package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// And returns an expression that is true when every one of expressions is true.
//
// $and evaluates its arguments in order and stops at the first false one. What counts as false is narrower than it
// looks: only false, null, 0, and missing are, so an empty string and an empty array are both true. With no
// arguments the result is true.
//
// Example:
//
//	expr.And(expr.Gte(expr.Field("age"), 18), expr.Eq(expr.Field("status"), "active"))
//	// bson.D{{Key: "$and", Value: bson.A{
//	//     bson.D{{Key: "$gte", Value: bson.A{"$age", 18}}},
//	//     bson.D{{Key: "$eq", Value: bson.A{"$status", "active"}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/and/
func And(expressions ...any) bson.D {
	return bson.D{{Key: "$and", Value: operands(expressions)}}
}

// Not returns an expression that is true when expression is false.
//
// $not takes a single expression, unlike [And] and [Or], and yields the opposite boolean, treating false, null, 0,
// and missing as false and everything else as true. It is not the query operator monq.Not, which negates a field's
// operator expression and cannot wrap arbitrary conditions; this one negates any expression at all.
//
// Example:
//
//	expr.Not(expr.Eq(expr.Field("status"), "banned"))
//	// bson.D{{Key: "$not", Value: bson.A{
//	//     bson.D{{Key: "$eq", Value: bson.A{"$status", "banned"}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/not/
func Not(expression any) bson.D {
	return bson.D{{Key: "$not", Value: bson.A{expression}}}
}

// Or returns an expression that is true when at least one of expressions is true.
//
// $or evaluates its arguments in order and stops at the first true one, counting false, null, 0, and missing as
// false. With no arguments the result is false.
//
// Example:
//
//	expr.Or(expr.Gte(expr.Field("score"), 90), expr.Eq(expr.Field("staff"), true))
//	// bson.D{{Key: "$or", Value: bson.A{
//	//     bson.D{{Key: "$gte", Value: bson.A{"$score", 90}}},
//	//     bson.D{{Key: "$eq", Value: bson.A{"$staff", true}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/or/
func Or(expressions ...any) bson.D {
	return bson.D{{Key: "$or", Value: operands(expressions)}}
}
