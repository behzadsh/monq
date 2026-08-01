package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// GraphLookupOption configures one optional field of a [GraphLookup] stage.
type GraphLookupOption func(*bson.D)

// DepthField returns a [GraphLookupOption] that records how many hops each found document is from the start.
//
// The name is an output field name added to every document in the result array, holding the recursion depth as a
// long, with 0 for the documents found by the first lookup. Without it the depth is not reported at all.
func DepthField(name string) GraphLookupOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "depthField", Value: name})
	}
}

// MaxDepth returns a [GraphLookupOption] that stops the recursion after n further lookups.
//
// A value of 0 means only the documents matched by the first lookup, 1 adds one more round, and so on. Without it
// the search runs until it finds nothing new, so it is worth setting on data that could form long chains.
func MaxDepth(n int) GraphLookupOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "maxDepth", Value: n})
	}
}

// RestrictSearchWithMatch returns a [GraphLookupOption] limiting the recursion to documents matching filter.
//
// The filter is an ordinary query document, the same as [Match] takes, and it applies to every document the search
// considers, so a document it rejects also stops the search through that document. $expr and $text are not allowed
// in it.
func RestrictSearchWithMatch(filter bson.D) GraphLookupOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "restrictSearchWithMatch", Value: filter})
	}
}

// GraphLookup returns a stage that follows a chain of references recursively.
//
// $graphLookup starts from the startWith expression, looks for documents in the from collection whose
// connectToField matches it, then repeats using those documents' connectFromField values, until nothing new turns
// up or [MaxDepth] cuts it short. Everything it finds lands in the as array field, deduplicated and in no
// particular order, which is what makes it the way to walk hierarchies such as an employee reporting chain.
//
// The from collection has to be in the same database and cannot be sharded. The starting point is an aggregation
// expression, so a field reference there takes the "$field" form.
//
// Example:
//
//	stage.GraphLookup("employees", "$reports_to", "reports_to", "name", "chain", stage.MaxDepth(3))
//	// bson.D{{Key: "$graphLookup", Value: bson.D{
//	//     {Key: "from", Value: "employees"},
//	//     {Key: "startWith", Value: "$reports_to"},
//	//     {Key: "connectFromField", Value: "reports_to"},
//	//     {Key: "connectToField", Value: "name"},
//	//     {Key: "as", Value: "chain"},
//	//     {Key: "maxDepth", Value: 3},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/graphLookup/
func GraphLookup(
	from string,
	startWith any,
	connectFromField, connectToField monq.FieldPath,
	as string,
	opts ...GraphLookupOption,
) bson.D {
	spec := bson.D{
		{Key: "from", Value: from},
		{Key: "startWith", Value: startWith},
		{Key: "connectFromField", Value: string(connectFromField)},
		{Key: "connectToField", Value: string(connectToField)},
		{Key: "as", Value: as},
	}

	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$graphLookup", Value: spec}}
}

// Lookup returns a stage that joins documents from another collection on equality.
//
// $lookup adds an array field holding, for each input document, the documents of the from collection whose
// foreignField equals the input's localField. The array is empty when nothing matches, which makes the stage a
// left outer join; following it with an [Unwind] turns it into an inner join, or a left outer one with
// [PreserveNullAndEmptyArrays]. A missing localField is treated as null and matches foreign nulls.
//
// The from collection has to be in the same database. For a join on anything other than equality, or one that
// filters the joined documents, use [LookupPipeline].
//
// Example:
//
//	stage.Lookup("orders", "_id", "customer_id", "orders")
//	// bson.D{{Key: "$lookup", Value: bson.D{
//	//     {Key: "from", Value: "orders"},
//	//     {Key: "localField", Value: "_id"},
//	//     {Key: "foreignField", Value: "customer_id"},
//	//     {Key: "as", Value: "orders"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lookup/
func Lookup(from string, localField, foreignField monq.FieldPath, as string) bson.D {
	return bson.D{{Key: "$lookup", Value: bson.D{
		{Key: "from", Value: from},
		{Key: "localField", Value: string(localField)},
		{Key: "foreignField", Value: string(foreignField)},
		{Key: "as", Value: as},
	}}}
}

// LookupPipeline returns a stage that joins documents from another collection through a sub-pipeline.
//
// This is the form of $lookup that runs a pipeline against the from collection instead of matching one field, so
// the join condition can be anything and the joined documents can be filtered, sorted, or reshaped before they
// land in the as array.
//
// Values from the outer document reach the sub-pipeline through let, and referring to them takes two things that
// are easy to miss: inside the pipeline a let variable is written with two dollar signs, "$$order_id" rather than
// "$order_id", and a stage that compares one against a field of the joined collection has to go through $expr,
// since an ordinary query condition cannot see variables. Passing a nil let leaves the field out entirely.
//
// Example:
//
//	stage.LookupPipeline("orders",
//		bson.D{{Key: "customer", Value: "$_id"}},
//		[]bson.D{stage.Match(monq.Expr(bson.D{{Key: "$eq", Value: bson.A{"$customer_id", "$$customer"}}}))},
//		"orders",
//	)
//	// bson.D{{Key: "$lookup", Value: bson.D{
//	//     {Key: "from", Value: "orders"},
//	//     {Key: "let", Value: bson.D{{Key: "customer", Value: "$_id"}}},
//	//     {Key: "pipeline", Value: bson.A{...}},
//	//     {Key: "as", Value: "orders"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lookup/
func LookupPipeline(from string, let bson.D, pipeline []bson.D, as string) bson.D {
	spec := bson.D{{Key: "from", Value: from}}
	if let != nil {
		spec = append(spec, bson.E{Key: "let", Value: let})
	}

	spec = append(spec,
		bson.E{Key: "pipeline", Value: toStageArray(pipeline)},
		bson.E{Key: "as", Value: as},
	)

	return bson.D{{Key: "$lookup", Value: spec}}
}

// UnionWith returns a stage that appends the documents of another collection to the pipeline.
//
// $unionWith runs after the stages before it and concatenates the results, without deduplicating: it is a UNION
// ALL rather than a UNION. The optional sub-pipeline runs against the other collection first, which is how its
// documents get filtered or reshaped to match. With no stages the pipeline field is left out, and the whole
// collection is appended.
//
// The sub-pipeline cannot contain $out or $merge.
//
// Example:
//
//	stage.UnionWith("archived_orders", stage.Match(monq.Gte("created_at", "2026-01-01")))
//	// bson.D{{Key: "$unionWith", Value: bson.D{
//	//     {Key: "coll", Value: "archived_orders"},
//	//     {Key: "pipeline", Value: bson.A{...}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unionWith/
func UnionWith(coll string, pipeline ...bson.D) bson.D {
	spec := bson.D{{Key: "coll", Value: coll}}
	if len(pipeline) > 0 {
		spec = append(spec, bson.E{Key: "pipeline", Value: toStageArray(pipeline)})
	}

	return bson.D{{Key: "$unionWith", Value: spec}}
}
