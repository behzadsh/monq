package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// AddFields returns a stage that adds fields to the documents passing through it.
//
// $addFields keeps every existing field and adds the ones given, which is what separates it from [Project], where
// anything left out is dropped. A field that already exists is overwritten, so it doubles as a way to recompute
// one. Fields are built with [Field] and merged into one document, with a later field of the same name winning.
// [Set] is the same stage under MongoDB's newer name.
//
// Example:
//
//	stage.AddFields(stage.Field("total", bson.D{{Key: "$sum", Value: "$items.price"}}))
//	// bson.D{
//	//     {
//	//         Key: "$addFields",
//	//         Value: bson.D{
//	//             {Key: "total", Value: bson.D{{Key: "$sum", Value: "$items.price"}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/addFields/
func AddFields(fields ...bson.D) bson.D {
	return bson.D{{Key: "$addFields", Value: mergeFields(bson.D{}, fields)}}
}

// Field returns one name-to-value pair of a stage that describes its output field by field.
//
// It is the building block [Project], [AddFields], and [Set] take, and the same pair [Accumulator] builds for the
// grouping stages. The name is a path into the document, dots included, which is why it is a monq.FieldPath: these
// stages name fields that already exist as often as they create new ones. It does not take the "$field" form
// though, which belongs on the value side. The value is an aggregation expression, or 1 and 0 in a [Project] to
// include and exclude.
//
// Example:
//
//	stage.Field("name", "$profile.display_name")
//	// bson.D{{Key: "name", Value: "$profile.display_name"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/project/
func Field(name monq.FieldPath, value any) bson.D {
	return bson.D{{Key: string(name), Value: value}}
}

// Project returns a stage that reshapes documents, keeping, renaming, and computing fields.
//
// $project emits a new document per input document holding only what it names. Passing 1 keeps a field and 0
// removes it, and the two cannot be mixed in one $project except for _id, which is included by default and is the
// one field that can be switched off alongside inclusions. Naming a field with an expression computes it. Fields
// are built with [Field] and merged into one document, with a later field of the same name winning; a
// hand-written bson.D works just as well, since it is the same type.
//
// Adding fields without dropping the rest is [AddFields] or [Set]; removing a few and keeping the rest is [Unset].
//
// Example:
//
//	stage.Project(stage.Field("name", 1), stage.Field("_id", 0))
//	// bson.D{{Key: "$project", Value: bson.D{{Key: "name", Value: 1}, {Key: "_id", Value: 0}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/project/
func Project(fields ...bson.D) bson.D {
	return bson.D{{Key: "$project", Value: mergeFields(bson.D{}, fields)}}
}

// ReplaceRoot returns a stage that promotes an embedded document to be the whole document.
//
// $replaceRoot swaps each document for the one newRoot evaluates to, so everything not inside it is gone. The
// expression has to produce a document: on an input where it resolves to nothing or to a non-document, the
// aggregation fails rather than skipping that document, which is the usual surprise with "$subdoc" as the
// argument. Guard it with $ifNull or $mergeObjects when the field may be missing. [ReplaceWith] is the same stage
// under MongoDB's newer name.
//
// Example:
//
//	stage.ReplaceRoot("$profile")
//	// bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$profile"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceRoot/
func ReplaceRoot(newRoot any) bson.D {
	return bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: newRoot}}}}
}

// ReplaceWith returns a stage that promotes an embedded document to be the whole document.
//
// $replaceWith is [ReplaceRoot] without the wrapper document around the expression, added in MongoDB 4.2. The two
// are interchangeable and carry the same requirement that the expression resolve to a document.
//
// Example:
//
//	stage.ReplaceWith("$profile")
//	// bson.D{{Key: "$replaceWith", Value: "$profile"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceWith/
func ReplaceWith(expression any) bson.D {
	return bson.D{{Key: "$replaceWith", Value: expression}}
}

// Set returns a stage that adds fields to the documents passing through it.
//
// $set is [AddFields] under the name MongoDB gave it in 4.2, with identical behavior: existing fields stay, named
// ones are added or overwritten. It is also where the package split earns its keep, since monq.Set is the $set
// update operator and stage.Set is this stage.
//
// Example:
//
//	stage.Set(stage.Field("status", "active"))
//	// bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/set/
func Set(fields ...bson.D) bson.D {
	return bson.D{{Key: "$set", Value: mergeFields(bson.D{}, fields)}}
}

// Unset returns a stage that removes fields from the documents passing through it.
//
// $unset drops the named fields and keeps everything else, the complement of an inclusion [Project]. Paths reach
// into subdocuments with dots; a path through an array removes the field from every element. One field is emitted
// as a string and several as an array, the two forms MongoDB accepts.
//
// Example:
//
//	stage.Unset("password_hash", "internal.notes")
//	// bson.D{{Key: "$unset", Value: bson.A{"password_hash", "internal.notes"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unset/
func Unset(fields ...monq.FieldPath) bson.D {
	return bson.D{{Key: "$unset", Value: fieldNames(fields)}}
}
