package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// AllElementsTrue returns an expression that is true when every element of an array is true.
//
// $allElementsTrue takes its array wrapped in another array, a quirk of the operator rather than a choice here,
// and counts false, null, 0, and undefined as false. An empty array yields true.
//
// Example:
//
//	expr.AllElementsTrue(expr.Field("checks"))
//	// bson.D{{Key: "$allElementsTrue", Value: bson.A{"$checks"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/allElementsTrue/
func AllElementsTrue(array any) bson.D {
	return bson.D{{Key: "$allElementsTrue", Value: bson.A{array}}}
}

// AnyElementTrue returns an expression that is true when at least one element of an array is true.
//
// $anyElementTrue is wrapped the same way as [AllElementsTrue] and counts truth the same way. An empty array
// yields false.
//
// Example:
//
//	expr.AnyElementTrue(expr.Field("checks"))
//	// bson.D{{Key: "$anyElementTrue", Value: bson.A{"$checks"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/anyElementTrue/
func AnyElementTrue(array any) bson.D {
	return bson.D{{Key: "$anyElementTrue", Value: bson.A{array}}}
}

// SetDifference returns an expression yielding the elements of a that are not in b.
//
// $setDifference takes exactly two arrays and treats them as sets: order is ignored and duplicates are dropped, so
// the result never repeats an element. Use [ConcatArrays] and [Filter] when the array's order or repeats matter.
//
// Example:
//
//	expr.SetDifference(expr.Field("tags"), expr.Field("banned_tags"))
//	// bson.D{{Key: "$setDifference", Value: bson.A{"$tags", "$banned_tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setDifference/
func SetDifference(a, b any) bson.D {
	return bson.D{{Key: "$setDifference", Value: bson.A{a, b}}}
}

// SetEquals returns an expression that is true when its arrays hold the same elements.
//
// $setEquals compares as sets, so order and repeats do not count: [1, 1, 2] and [2, 1] are equal. It takes two or
// more arrays.
//
// Example:
//
//	expr.SetEquals(expr.Field("tags"), expr.Field("expected_tags"))
//	// bson.D{{Key: "$setEquals", Value: bson.A{"$tags", "$expected_tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setEquals/
func SetEquals(arrays ...any) bson.D {
	return bson.D{{Key: "$setEquals", Value: operands(arrays)}}
}

// SetIntersection returns an expression yielding the elements common to every array given.
//
// $setIntersection treats its arguments as sets, so the result has no duplicates and no meaningful order. With no
// arguments it yields an empty array.
//
// Example:
//
//	expr.SetIntersection(expr.Field("tags"), expr.Field("featured_tags"))
//	// bson.D{{Key: "$setIntersection", Value: bson.A{"$tags", "$featured_tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setIntersection/
func SetIntersection(arrays ...any) bson.D {
	return bson.D{{Key: "$setIntersection", Value: operands(arrays)}}
}

// SetIsSubset returns an expression that is true when every element of a is also in b.
//
// $setIsSubset takes exactly two arrays and compares them as sets, so repeats in a do not have to be repeated in
// b. An empty first array is a subset of anything.
//
// Example:
//
//	expr.SetIsSubset(expr.Field("required_tags"), expr.Field("tags"))
//	// bson.D{{Key: "$setIsSubset", Value: bson.A{"$required_tags", "$tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setIsSubset/
func SetIsSubset(a, b any) bson.D {
	return bson.D{{Key: "$setIsSubset", Value: bson.A{a, b}}}
}

// SetUnion returns an expression yielding every element appearing in any of the arrays given.
//
// $setUnion treats its arguments as sets, so the result is deduplicated and its order is not meaningful.
// [ConcatArrays] is the one that keeps both.
//
// Example:
//
//	expr.SetUnion(expr.Field("tags"), expr.Field("extra_tags"))
//	// bson.D{{Key: "$setUnion", Value: bson.A{"$tags", "$extra_tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setUnion/
func SetUnion(arrays ...any) bson.D {
	return bson.D{{Key: "$setUnion", Value: operands(arrays)}}
}
