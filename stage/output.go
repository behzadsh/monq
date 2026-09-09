package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// Documents returns a stage that produces the given documents out of nothing.
//
// $documents starts a pipeline from literal input rather than from a collection, which is what makes it useful for
// trying an expression out or for feeding a $lookup or $unionWith a small fixed set. It has to be the first stage,
// and since it takes no input it runs against a database rather than a collection: Database.Aggregate, not
// Collection.Aggregate.
//
// Example:
//
//	stage.Documents(bson.D{{Key: "x", Value: 1}}, bson.D{{Key: "x", Value: 2}})
//	// bson.D{
//	//     {
//	//         Key: "$documents",
//	//         Value: bson.A{
//	//             bson.D{{Key: "x", Value: 1}},
//	//             bson.D{{Key: "x", Value: 2}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/documents/
func Documents(docs ...any) bson.D {
	return bson.D{{Key: "$documents", Value: toArray(docs)}}
}

// MergeOption configures one optional field of a [Merge] stage.
type MergeOption func(*bson.D)

// MergeLet returns a [MergeOption] defining variables for a pipeline given to [MergeWhenMatched].
//
// Inside that pipeline a variable is written with two dollar signs, "$$new" rather than "$new". MongoDB defines
// $$new by itself, holding the incoming document, so this option is only needed for anything beyond it.
func MergeLet(vars bson.D) MergeOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "let", Value: vars})
	}
}

// MergeOn returns a [MergeOption] naming the fields that decide whether a document already exists.
//
// The target collection needs a unique index on exactly these fields, in any order. Without this option the match
// is on _id alone. One field is emitted as a string and several as an array, the two forms MongoDB accepts.
func MergeOn(fields ...monq.FieldPath) MergeOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "on", Value: fieldNames(fields)})
	}
}

// MergeWhenMatched returns a [MergeOption] saying what to do when a document already exists in the target.
//
// The action is one of "merge" (the default, which overlays the incoming fields), "replace", "keepExisting",
// "fail", or a pipeline of stages that computes the new document. A pipeline may only hold $addFields, $set,
// $project, $unset, $replaceRoot, and $replaceWith, and refers to the incoming document as "$$new".
func MergeWhenMatched(action any) MergeOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "whenMatched", Value: action})
	}
}

// MergeWhenNotMatched returns a [MergeOption] saying what to do when a document is not in the target yet.
//
// The action is one of "insert" (the default), "discard", or "fail".
func MergeWhenNotMatched(action string) MergeOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "whenNotMatched", Value: action})
	}
}

// Merge returns a stage that writes the pipeline's results into a collection, document by document.
//
// $merge decides per document whether it already exists in the target, matching on _id or on the fields given to
// [MergeOn], and then applies [MergeWhenMatched] and [MergeWhenNotMatched]. That is what separates it from [Out],
// which throws the whole target collection away and writes a new one: $merge can add to a collection, update it in
// place, or build a rolled-up summary that later runs extend.
//
// It has to be the last stage of the pipeline. The target may be in another database, and unlike $out it may be
// sharded. Into takes the collection name, or a [Namespace] for a collection in another database.
//
// Example:
//
//	stage.Merge("daily_totals", stage.MergeOn("date"), stage.MergeWhenMatched("replace"))
//	// bson.D{
//	//     {
//	//         Key: "$merge",
//	//         Value: bson.D{
//	//             {Key: "into", Value: "daily_totals"},
//	//             {Key: "on", Value: "date"},
//	//             {Key: "whenMatched", Value: "replace"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/merge/
func Merge(into any, opts ...MergeOption) bson.D {
	spec := bson.D{{Key: "into", Value: into}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$merge", Value: spec}}
}

// Namespace returns the {db, coll} target document that [Out] and [Merge] take for a collection in another
// database.
//
// Both stages also accept a plain collection name, which means the database the aggregation runs against. This
// builder is for the other case, and keeps the two stages taking one target argument instead of growing a second
// function each.
//
// Example:
//
//	stage.Namespace("reports", "daily_totals")
//	// bson.D{{Key: "db", Value: "reports"}, {Key: "coll", Value: "daily_totals"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/out/
func Namespace(db, coll string) bson.D {
	return bson.D{{Key: "db", Value: db}, {Key: "coll", Value: coll}}
}

// Out returns a stage that writes the pipeline's results into a collection, replacing whatever was there.
//
// $out drops the target collection and recreates it from the pipeline's output, atomically enough that readers see
// either the old collection or the new one. Anything already in the target is gone, indexes included, which is the
// difference from [Merge]: $out replaces, $merge incorporates.
//
// It has to be the last stage of the pipeline, and the target cannot be sharded. Target takes the collection name,
// or a [Namespace] for a collection in another database.
//
// Example:
//
//	stage.Out("daily_totals")
//	// bson.D{{Key: "$out", Value: "daily_totals"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/out/
func Out(target any) bson.D {
	return bson.D{{Key: "$out", Value: target}}
}

// fieldNames renders paths as the scalar MongoDB expects for one field and the array it expects for several.
func fieldNames(fields []monq.FieldPath) any {
	if len(fields) == 1 {
		return string(fields[0])
	}

	names := make(bson.A, len(fields))
	for i, f := range fields {
		names[i] = string(f)
	}

	return names
}
