package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Exclude returns a projection entry leaving the given fields out of the returned documents.
//
// Excluding fields returns everything else, so a projection built only from Exclude entries needs no matching
// [Include]. The two cannot be mixed, with one exception: "_id" is returned unless asked otherwise, so excluding
// it alongside inclusions is allowed and is the usual way to drop it.
//
// Example:
//
//	monq.Exclude("password_hash", "internal.notes")
//	// bson.D{{Key: "password_hash", Value: 0}, {Key: "internal.notes", Value: 0}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/tutorial/project-fields-from-query-results/
func Exclude(fields ...FieldPath) bson.D {
	return fieldKeys(fields, 0)
}

// Include returns a projection entry keeping the given fields in the returned documents.
//
// Everything not named is left out, apart from "_id", which comes back unless [Exclude] drops it. Paths reach into
// subdocuments with dots, and naming a field inside an array keeps that field of every element. Combining this with
// [ArrayPath.Positional] is what produces the "$" positional projection, which returns only the array element the
// query matched.
//
// Example:
//
//	monq.Include("email", "profile.name")
//	// bson.D{{Key: "email", Value: 1}, {Key: "profile.name", Value: 1}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/tutorial/project-fields-from-query-results/
func Include(fields ...FieldPath) bson.D {
	return fieldKeys(fields, 1)
}

// Meta returns an entry holding a value MongoDB computes about the document rather than stores in it.
//
// The kind is "textScore" for the relevance of a $text search, "indexKey" for the index entry a document was found
// through, or one of the Atlas search kinds such as "searchScore" and "searchHighlights". [TextScore] is the same
// thing spelled for the common case.
//
// It works in a projection, where it adds the value under the given name, and in a sort document, where it orders
// by that value. Asking for a text score without a $text search in the same query is an error.
//
// Example:
//
//	monq.Meta("indexKey", "indexKey")
//	// bson.D{{Key: "indexKey", Value: bson.D{{Key: "$meta", Value: "indexKey"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/meta/
func Meta(field FieldPath, kind string) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$meta", Value: kind}}}}
}

// Projection returns the document deciding which fields come back from a query.
//
// Entries come from [Include], [Exclude], [Slice], [SliceFrom], [Meta], and [ElemMatch], and are applied in the
// order given, a later entry replacing an earlier one for the same field while keeping its position. The result
// goes into options.Find().SetProjection and the same option on the other read and find-and-modify operations.
//
// A projection is either inclusions or exclusions, not both, since the two answer different questions: one says
// what to keep, the other what to drop. "_id" is the exception, returned unless excluded, so dropping it beside a
// list of inclusions is allowed. An empty Projection returns whole documents.
//
// Only one positional entry is allowed per projection, and a field cannot have both a positional and an
// [ElemMatch] entry.
//
// Example:
//
//	monq.Projection(monq.Include("email", "items.sku"), monq.Exclude("_id"))
//	// bson.D{{Key: "email", Value: 1}, {Key: "items.sku", Value: 1}, {Key: "_id", Value: 0}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/tutorial/project-fields-from-query-results/
func Projection(entries ...bson.D) bson.D {
	return mergeEntries(entries)
}

// Slice returns a projection entry returning only part of an array field.
//
// A positive n takes the first n elements and a negative n the last n, so Slice("comments", -5) is the five most
// recent entries of an append-ordered array. [SliceFrom] is the form that skips before taking.
//
// This is the projection $slice, which trims what a query returns. Two others share the name and do different
// jobs: [PushSlice] caps an array during an update, and the Slice of monq/expr takes part of an array inside an
// aggregation expression.
//
// Whether a $slice entry may sit beside exclusions of other fields has varied between server versions. Keeping a
// projection all inclusions or all exclusions, with "_id" the one exception, avoids the question.
//
// Example:
//
//	monq.Slice("comments", -5)
//	// bson.D{{Key: "comments", Value: bson.D{{Key: "$slice", Value: -5}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/projection/slice/
func Slice(field FieldPath, n int) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$slice", Value: n}}}}
}

// SliceFrom returns a projection entry returning part of an array field, starting at an offset.
//
// The skip counts from the start of the array, or from the end when negative, and the limit is how many elements
// to take from there and has to be positive. It is the two-element form of the projection $slice, and the way to
// page through a long array; [Slice] is the shorter form that takes from one end.
//
// Example:
//
//	monq.SliceFrom("comments", 10, 5)
//	// bson.D{{Key: "comments", Value: bson.D{{Key: "$slice", Value: bson.A{10, 5}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/projection/slice/
func SliceFrom(field FieldPath, skip, limit int) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$slice", Value: bson.A{skip, limit}}}}}
}
