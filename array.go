package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// All returns a filter that matches documents where field is an array containing all of values.
//
// $all matches an array field that contains every one of the given elements, in any order and regardless of extra
// elements the array also holds. Against a non-array field it matches only when the field equals the single element
// of a one-element $all list. Values is variadic with the same un-spread-slice gotcha as [In]: All(field, tags) where
// tags is a []string produces one candidate whose value is the slice itself. Spread it: All(field, tags...).
//
// Example:
//
//	monq.All("tags", "go", "mongodb")
//	// bson.D{{Key: "tags", Value: bson.D{{Key: "$all", Value: bson.A{"go", "mongodb"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/all/
func All(field FieldPath, values ...any) bson.D {
	arr := make(bson.A, len(values))
	copy(arr, values)

	return bson.D{{Key: string(field), Value: bson.D{{Key: "$all", Value: arr}}}}
}

// ElemMatch returns a filter that matches documents where field is an array with at least one element satisfying
// all of filters.
//
// $elemMatch requires a single array element to satisfy every criterion at once, unlike plain dotted conditions on
// the same array, which may be satisfied by different elements. Filters are concatenated into one criteria document,
// so two conditions on the same field (ElemMatch("items", Gt("qty", 2), Lt("qty", 9))) produce a duplicate key that
// MongoDB does not read as an AND: express both bounds in one filter instead. Arrays of scalars use a different form
// where the criteria are bare operator expressions with no field name, which no monq comparison operator emits; build
// those with [Raw], e.g. ElemMatch("scores", Raw(bson.D{{Key: "$gte", Value: 80}})).
//
// The same document works as a [Projection] entry, where it returns only the first element matching the criteria
// instead of selecting documents. A field cannot carry both that and a positional entry.
//
// Example:
//
//	monq.ElemMatch("items", monq.Eq("sku", "abc"), monq.Gte("qty", 2))
//	// bson.D{{Key: "items", Value: bson.D{{Key: "$elemMatch", Value: bson.D{
//	//     {Key: "sku", Value: bson.D{{Key: "$eq", Value: "abc"}}},
//	//     {Key: "qty", Value: bson.D{{Key: "$gte", Value: 2}}},
//	// }}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/elemMatch/
func ElemMatch(field FieldPath, filters ...bson.D) bson.D {
	criteria := bson.D{}
	for _, f := range filters {
		criteria = append(criteria, f...)
	}

	return bson.D{{Key: string(field), Value: bson.D{{Key: "$elemMatch", Value: criteria}}}}
}

// Size returns a filter that matches documents where field is an array of exactly size elements.
//
// $size matches on exact array length only. It accepts no range expression, so "more than n elements" needs a
// different approach, such as [Exists] on the nth index ("tags.3"). Fields that are not arrays never match.
//
// Example:
//
//	monq.Size("tags", 3)
//	// bson.D{{Key: "tags", Value: bson.D{{Key: "$size", Value: 3}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/size/
func Size(field FieldPath, size int) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$size", Value: size}}}}
}
