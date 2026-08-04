package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// Accumulator returns one output field of a [Group], [Bucket], or [BucketAuto] stage.
//
// Grouping stages describe their output as a document of name-to-accumulator pairs, such as
// {total: {$sum: "$amount"}}. Accumulator builds one of those pairs, and the stage merges however many it is
// given, so each piece stays a bson.D like everything else in monq. The name is an output field name: it cannot
// contain a dot, cannot start with a $, and in a [Group] it cannot be _id, which the group key already owns.
//
// It builds the same pair [Field] does and exists under its own name because grouping stages call these fields
// accumulators; use whichever reads better where you are.
//
// Example:
//
//	stage.Accumulator("total", bson.D{{Key: "$sum", Value: "$amount"}})
//	// bson.D{{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/group/#accumulator-operator
func Accumulator(name string, accumulator any) bson.D {
	return Field(monq.FieldPath(name), accumulator)
}

// BucketOption configures one optional field of a [Bucket] or [BucketAuto] stage.
//
// The two stages share this type because they share the output field. The options that belong to only one of them
// say so: [BucketDefault] is for [Bucket] and [BucketGranularity] for [BucketAuto]. Passing one to the wrong stage
// compiles and produces a document the server rejects, in keeping with monq validating nothing itself.
type BucketOption func(*bson.D)

// BucketDefault returns a [BucketOption] naming the bucket that collects documents outside the boundaries.
//
// It applies to [Bucket] only. Without it, a document whose groupBy value falls outside every boundary makes the
// whole aggregation fail, which is the usual surprise with $bucket. The value must sort outside the boundary
// values, either below the first or above the last, and is commonly a string such as "other" since strings sort
// after numbers.
func BucketDefault(value any) BucketOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "default", Value: value})
	}
}

// BucketGranularity returns a [BucketOption] setting the preferred number series for the bucket boundaries.
//
// It applies to [BucketAuto] only, and takes one of MongoDB's series names: "R5", "R10", "R20", "R40", "R80",
// "1-2-5", "E6", "E12", "E24", "E48", "E96", "E192", or "POWERSOF2". Boundaries then land on round numbers of that
// series, which can leave fewer buckets than asked for. The groupBy values have to be numeric for it to work.
func BucketGranularity(name string) BucketOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "granularity", Value: name})
	}
}

// BucketOutput returns a [BucketOption] describing the fields each bucket carries.
//
// Accumulators are merged into one output document, with a later field of the same name winning. Without this
// option a bucket holds only its _id and a count field; with it, the count has to be asked for explicitly, since
// naming any output at all replaces the default.
func BucketOutput(accumulators ...bson.D) BucketOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "output", Value: mergeFields(bson.D{}, accumulators)})
	}
}

// Bucket returns a stage that sorts documents into buckets with explicit boundaries.
//
// $bucket groups documents by the groupBy expression, putting each into the bucket whose boundary range covers its
// value, and emits one document per non-empty bucket with the lower boundary as its _id. Boundaries have to be
// sorted ascending and share one type, and each bucket covers its lower boundary up to but not including the next
// one. A document whose value falls outside every bucket is a hard error unless [BucketDefault] gives it somewhere
// to go.
//
// Example:
//
//	stage.Bucket("$price", []any{0, 50, 100}, stage.BucketDefault("other"))
//	// bson.D{{Key: "$bucket", Value: bson.D{
//	//     {Key: "groupBy", Value: "$price"},
//	//     {Key: "boundaries", Value: bson.A{0, 50, 100}},
//	//     {Key: "default", Value: "other"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucket/
func Bucket(groupBy any, boundaries []any, opts ...BucketOption) bson.D {
	spec := bson.D{
		{Key: "groupBy", Value: groupBy},
		{Key: "boundaries", Value: toArray(boundaries)},
	}

	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$bucket", Value: spec}}
}

// BucketAuto returns a stage that spreads documents across a given number of buckets.
//
// $bucketAuto picks the boundaries itself, aiming for buckets of roughly equal size, so no boundary list is
// needed. It may produce fewer buckets than asked for when the data cannot be split that evenly, and
// [BucketGranularity] makes that more likely by rounding boundaries to a number series. Each emitted document has
// an _id of {min, max}.
//
// Example:
//
//	stage.BucketAuto("$price", 4, stage.BucketGranularity("R20"))
//	// bson.D{{Key: "$bucketAuto", Value: bson.D{
//	//     {Key: "groupBy", Value: "$price"},
//	//     {Key: "buckets", Value: 4},
//	//     {Key: "granularity", Value: "R20"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bucketAuto/
func BucketAuto(groupBy any, buckets int, opts ...BucketOption) bson.D {
	spec := bson.D{
		{Key: "groupBy", Value: groupBy},
		{Key: "buckets", Value: buckets},
	}

	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$bucketAuto", Value: spec}}
}

// Facet returns a stage that runs several sub-pipelines over the same input documents.
//
// $facet feeds every incoming document to each sub-pipeline and emits a single document whose fields are the
// sub-pipeline names, each holding that pipeline's results as an array. It is how one pass over the data produces
// several summaries at once. Build the sub-pipelines with [FacetPipeline]; facets sharing a name are merged with
// the last one winning.
//
// Sub-pipelines cannot contain $facet, $out, $merge, $geoNear, or $indexStats. Nothing here checks for them, so
// the server is what reports the mistake.
//
// Example:
//
//	stage.Facet(
//		stage.FacetPipeline("newest", stage.Sort(bson.D{{Key: "created_at", Value: -1}}), stage.Limit(5)),
//		stage.FacetPipeline("total", stage.Count("n")),
//	)
//	// bson.D{{Key: "$facet", Value: bson.D{
//	//     {Key: "newest", Value: bson.A{...}},
//	//     {Key: "total", Value: bson.A{...}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/facet/
func Facet(facets ...bson.D) bson.D {
	return bson.D{{Key: "$facet", Value: mergeFields(bson.D{}, facets)}}
}

// FacetPipeline returns one named sub-pipeline of a [Facet] stage.
//
// The name becomes a field of the document $facet emits, holding that sub-pipeline's output as an array. Stages
// are taken in the order given, exactly as [Pipeline] takes them.
//
// Example:
//
//	stage.FacetPipeline("total", stage.Count("n"))
//	// bson.D{{Key: "total", Value: bson.A{bson.D{{Key: "$count", Value: "n"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/facet/
func FacetPipeline(name string, stages ...bson.D) bson.D {
	return bson.D{{Key: name, Value: toStageArray(stages)}}
}

// Group returns a stage that groups documents by a key and accumulates values over each group.
//
// $group emits one document per distinct value of the id expression, which becomes the group's _id, plus whatever
// the accumulators compute. The id is an aggregation expression, so a field reference is written in the "$field"
// form; passing nil sends every document into a single group, which is the usual way to total a whole collection
// and reads as an omission only until you have seen it once.
//
// Accumulators are built with [Accumulator] and merged into one document, with a later field of the same name
// winning. $group does not preserve the input order, so a $sort belongs after it rather than before, and grouping
// buffers each group in memory unless the aggregation is allowed to use disk.
//
// Example:
//
//	stage.Group("$category", stage.Accumulator("total", bson.D{{Key: "$sum", Value: "$amount"}}))
//	// bson.D{{Key: "$group", Value: bson.D{
//	//     {Key: "_id", Value: "$category"},
//	//     {Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/group/
func Group(id any, accumulators ...bson.D) bson.D {
	group := mergeFields(bson.D{{Key: "_id", Value: id}}, accumulators)

	return bson.D{{Key: "$group", Value: group}}
}

// SortByCount returns a stage that groups documents by an expression and sorts the groups by size.
//
// $sortByCount is $group by the expression plus a count, followed by a descending $sort on that count, in one
// stage. Each emitted document holds the grouped value as _id and the number of documents as count. The expression
// is an aggregation expression, so a field reference takes the "$field" form.
//
// Example:
//
//	stage.SortByCount("$category")
//	// bson.D{{Key: "$sortByCount", Value: "$category"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sortByCount/
func SortByCount(expression any) bson.D {
	return bson.D{{Key: "$sortByCount", Value: expression}}
}

// mergeFields folds the fields of docs into target, with a later field of the same name replacing an earlier one.
func mergeFields(target bson.D, docs []bson.D) bson.D {
	for _, d := range docs {
		for _, e := range d {
			if i := indexOfKey(target, e.Key); i >= 0 {
				target[i].Value = e.Value

				continue
			}

			target = append(target, e)
		}
	}

	return target
}

// indexOfKey returns the position of key in d, or -1 when d has no such key.
func indexOfKey(d bson.D, key string) int {
	for i, e := range d {
		if e.Key == key {
			return i
		}
	}

	return -1
}

// toArray copies values into the bson.A the driver marshals as a BSON array.
func toArray(values []any) bson.A {
	arr := make(bson.A, len(values))
	copy(arr, values)

	return arr
}
