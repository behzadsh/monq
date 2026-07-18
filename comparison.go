package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Eq returns a filter that matches documents where field equals value.
//
// $eq matches documents where the field's value equals the specified value. It is functionally the same as the
// implicit equality shorthand ({field: value}), spelled out explicitly so every comparison operator monq builds
// has the same shape.
//
// Example:
//
//	monq.Eq("status", "active")
//	// bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/eq/
func Eq(field FieldPath, value any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$eq", Value: value}}}}
}

// Ne returns a filter that matches documents where field does not equal value.
//
// $ne matches all documents where the field's value is not equal to the specified value, including documents
// where the field is missing entirely. Combine with [Exists] if the field must also be present.
//
// Example:
//
//	monq.Ne("status", "banned")
//	// bson.D{{Key: "status", Value: bson.D{{Key: "$ne", Value: "banned"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/ne/
func Ne(field FieldPath, value any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$ne", Value: value}}}}
}

// Gt returns a filter that matches documents where field is greater than value.
//
// $gt matches documents where the field's value is strictly greater than the specified value, using BSON
// comparison order for the field's type.
//
// Example:
//
//	monq.Gt("age", 18)
//	// bson.D{{Key: "age", Value: bson.D{{Key: "$gt", Value: 18}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/gt/
func Gt(field FieldPath, value any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$gt", Value: value}}}}
}

// Gte returns a filter that matches documents where field is greater than or equal to value.
//
// $gte matches documents where the field's value is greater than or equal to the specified value, using BSON
// comparison order for the field's type.
//
// Example:
//
//	monq.Gte("stats.followers", 10000)
//	// bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$gte", Value: 10000}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/gte/
func Gte(field FieldPath, value any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$gte", Value: value}}}}
}

// Lt returns a filter that matches documents where field is less than value.
//
// $lt matches documents where the field's value is strictly less than the specified value, using BSON comparison
// order for the field's type.
//
// Example:
//
//	monq.Lt("age", 65)
//	// bson.D{{Key: "age", Value: bson.D{{Key: "$lt", Value: 65}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/lt/
func Lt(field FieldPath, value any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$lt", Value: value}}}}
}

// Lte returns a filter that matches documents where field is less than or equal to value.
//
// $lte matches documents where the field's value is less than or equal to the specified value, using BSON
// comparison order for the field's type.
//
// Example:
//
//	monq.Lte("age", 65)
//	// bson.D{{Key: "age", Value: bson.D{{Key: "$lte", Value: 65}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/lte/
func Lte(field FieldPath, value any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$lte", Value: value}}}}
}

// In returns a filter that matches documents where field's value equals any value in values.
//
// $in matches documents where the field's value equals any element in the given array, including matching an
// array field against any of its own elements. Values is variadic, and a single un-spread slice argument is not
// flattened: In(field, ids) where ids is a []string produces exactly one $in candidate whose value is the slice
// itself, not one candidate per element. Spread it: In(field, ids...).
//
// Example:
//
//	monq.In("status", "active", "pending")
//	// bson.D{{Key: "status", Value: bson.D{{Key: "$in", Value: bson.A{"active", "pending"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/in/
func In(field FieldPath, values ...any) bson.D {
	arr := make(bson.A, len(values))
	copy(arr, values)

	return bson.D{{Key: string(field), Value: bson.D{{Key: "$in", Value: arr}}}}
}

// Nin returns a filter that matches documents where field's value equals none of values.
//
// $nin matches documents where the field's value is not equal to any element in the given array, including
// documents where the field is missing entirely, the same as [Ne]. Values is variadic with the same
// un-spread-slice gotcha as [In]: spread it, Nin(field, ids...), rather than passing the slice directly.
//
// Example:
//
//	monq.Nin("status", "banned", "suspended")
//	// bson.D{{Key: "status", Value: bson.D{{Key: "$nin", Value: bson.A{"banned", "suspended"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/nin/
func Nin(field FieldPath, values ...any) bson.D {
	arr := make(bson.A, len(values))
	copy(arr, values)

	return bson.D{{Key: string(field), Value: bson.D{{Key: "$nin", Value: arr}}}}
}
