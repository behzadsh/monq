package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// And returns a filter that matches documents matching all of filters.
//
// $and joins filters with a logical AND, matching documents that satisfy every one of them. Top-level filters
// passed to And already behave as an implicit AND without this operator, so And is mainly useful to force
// multiple conditions on the same field (e.g. two separate [Gt]/[Lt] bounds) or to nest inside [Or]/[Nor].
//
// Example:
//
//	monq.And(monq.Eq("status", "active"), monq.Gte("age", 18))
//	// bson.D{
//	//     {
//	//         Key: "$and",
//	//         Value: bson.A{
//	//             bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}},
//	//             bson.D{{Key: "age", Value: bson.D{{Key: "$gte", Value: 18}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/and/
func And(filters ...bson.D) bson.D {
	arr := make(bson.A, len(filters))
	for i, f := range filters {
		arr[i] = f
	}

	return bson.D{{Key: "$and", Value: arr}}
}

// Or returns a filter that matches documents matching at least one of filters.
//
// $or joins filters with a logical OR, matching documents that satisfy at least one of them.
//
// Example:
//
//	monq.Or(monq.Gte("stats.followers", 10000), monq.Exists("verified_at", true))
//	// bson.D{
//	//     {
//	//         Key: "$or",
//	//         Value: bson.A{
//	//             bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$gte", Value: 10000}}}},
//	//             bson.D{{Key: "verified_at", Value: bson.D{{Key: "$exists", Value: true}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/or/
func Or(filters ...bson.D) bson.D {
	arr := make(bson.A, len(filters))
	for i, f := range filters {
		arr[i] = f
	}

	return bson.D{{Key: "$or", Value: arr}}
}

// Nor returns a filter that matches documents matching none of filters.
//
// $nor matches documents that fail every one of filters, the logical negation of [Or]. A document with a field
// that filters test is only excluded if that field also fails to match; documents missing the field entirely can
// still match $nor, the same way [Ne] and [Nin] treat missing fields as satisfying a negative condition.
//
// Example:
//
//	monq.Nor(monq.Eq("status", "banned"), monq.Eq("status", "suspended"))
//	// bson.D{
//	//     {
//	//         Key: "$nor",
//	//         Value: bson.A{
//	//             bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "banned"}}}},
//	//             bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "suspended"}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/nor/
func Nor(filters ...bson.D) bson.D {
	arr := make(bson.A, len(filters))
	for i, f := range filters {
		arr[i] = f
	}

	return bson.D{{Key: "$nor", Value: arr}}
}

// Not returns a filter that matches documents that do not match expr.
//
// $not negates the single operator expression expr, which must be a single-key filter document such as the output
// of [Gt], [In], or [Regex] (e.g. {field: {$gt: 5}} becomes {field: {$not: {$gt: 5}}}). Unlike [And]/[Or]/[Nor],
// $not applies to exactly one field's operator expression, not a list of filters, so it cannot wrap a multi-field
// filter or plain equality shorthand. The shape of expr is not validated: a malformed or empty expr panics on
// index access or produces a query MongoDB rejects at execution time, not a monq-level error.
//
// Example:
//
//	monq.Not(monq.Gt("age", 5))
//	// bson.D{{Key: "age", Value: bson.D{{Key: "$not", Value: bson.D{{Key: "$gt", Value: 5}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/not/
func Not(expr bson.D) bson.D {
	return bson.D{{Key: expr[0].Key, Value: bson.D{{Key: "$not", Value: expr[0].Value}}}}
}
