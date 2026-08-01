package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// Cmp returns an expression comparing two values and yielding -1, 0, or 1.
//
// $cmp reports whether a sorts before, with, or after b, using BSON comparison order across types. It is the
// building block for sorting logic inside a pipeline, where a boolean is not enough.
//
// Example:
//
//	expr.Cmp(expr.Field("score"), 50)
//	// bson.D{{Key: "$cmp", Value: bson.A{"$score", 50}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/cmp/
func Cmp(a, b any) bson.D {
	return bson.D{{Key: "$cmp", Value: bson.A{a, b}}}
}

// Eq returns an expression that is true when a and b are equal.
//
// $eq compares two expressions and yields a boolean, which is what separates it from the query operator of the
// same name: monq.Eq("status", "active") tests a field against a value, while expr.Eq compares whatever its two
// arguments evaluate to. A bare string is a value, not a field, so the field side needs [Field] or the "$status"
// form; expr.Eq("status", "active") compares two constants and is false for every document.
//
// Comparison is across types, using BSON comparison order, so 1 and "1" are never equal.
//
// Example:
//
//	expr.Eq(expr.Field("status"), "active")
//	// bson.D{{Key: "$eq", Value: bson.A{"$status", "active"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/eq/
func Eq(a, b any) bson.D {
	return bson.D{{Key: "$eq", Value: bson.A{a, b}}}
}

// Gt returns an expression that is true when a is greater than b.
//
// $gt compares two expressions in BSON comparison order and yields a boolean. Both sides are expressions, so a
// field goes in as [Field] output or a "$field" string.
//
// Example:
//
//	expr.Gt(expr.Field("score"), 50)
//	// bson.D{{Key: "$gt", Value: bson.A{"$score", 50}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/gt/
func Gt(a, b any) bson.D {
	return bson.D{{Key: "$gt", Value: bson.A{a, b}}}
}

// Gte returns an expression that is true when a is greater than or equal to b.
//
// $gte compares two expressions in BSON comparison order and yields a boolean.
//
// Example:
//
//	expr.Gte(expr.Field("age"), 18)
//	// bson.D{{Key: "$gte", Value: bson.A{"$age", 18}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/gte/
func Gte(a, b any) bson.D {
	return bson.D{{Key: "$gte", Value: bson.A{a, b}}}
}

// Lt returns an expression that is true when a is less than b.
//
// $lt compares two expressions in BSON comparison order and yields a boolean.
//
// Example:
//
//	expr.Lt(expr.Field("stock"), 10)
//	// bson.D{{Key: "$lt", Value: bson.A{"$stock", 10}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lt/
func Lt(a, b any) bson.D {
	return bson.D{{Key: "$lt", Value: bson.A{a, b}}}
}

// Lte returns an expression that is true when a is less than or equal to b.
//
// $lte compares two expressions in BSON comparison order and yields a boolean.
//
// Example:
//
//	expr.Lte(expr.Field("stock"), 10)
//	// bson.D{{Key: "$lte", Value: bson.A{"$stock", 10}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lte/
func Lte(a, b any) bson.D {
	return bson.D{{Key: "$lte", Value: bson.A{a, b}}}
}

// Ne returns an expression that is true when a and b differ.
//
// $ne compares two expressions in BSON comparison order and yields a boolean, the negation of [Eq].
//
// Example:
//
//	expr.Ne(expr.Field("status"), "banned")
//	// bson.D{{Key: "$ne", Value: bson.A{"$status", "banned"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ne/
func Ne(a, b any) bson.D {
	return bson.D{{Key: "$ne", Value: bson.A{a, b}}}
}
