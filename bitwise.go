package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// BitsAllClear returns a filter that matches documents where every bit selected by mask is clear in field.
//
// $bitsAllClear tests the bit pattern of a numeric or binData field. The mask is a bitmask number, a BSON binary
// value, or an array of bit positions counted from the least significant bit (bson.A{0, 3}). Only integers, doubles
// that hold a whole number, and binData values are considered; every other type, including a missing field, fails
// to match. A negative bitmask or a non-integral double as the mask is a server-side error.
//
// This family tests bits. The other two do different jobs with similar names: [BitAnd] and its siblings are update
// operators rewriting a stored field, and the expression operators in monq/expr compute a value from bits inside an
// aggregation.
//
// Example:
//
//	monq.BitsAllClear("flags", 6)
//	// bson.D{{Key: "flags", Value: bson.D{{Key: "$bitsAllClear", Value: 6}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAllClear/
func BitsAllClear(field FieldPath, mask any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$bitsAllClear", Value: mask}}}}
}

// BitsAllSet returns a filter that matches documents where every bit selected by mask is set in field.
//
// $bitsAllSet is the counterpart of [BitsAllClear] and takes the same mask forms: a bitmask number, a BSON binary
// value, or an array of bit positions. It matches only numeric and binData fields, so documents missing the field
// never match.
//
// Example:
//
//	monq.BitsAllSet("flags", bson.A{1, 5})
//	// bson.D{{Key: "flags", Value: bson.D{{Key: "$bitsAllSet", Value: bson.A{1, 5}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAllSet/
func BitsAllSet(field FieldPath, mask any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$bitsAllSet", Value: mask}}}}
}

// BitsAnyClear returns a filter that matches documents where at least one bit selected by mask is clear in field.
//
// $bitsAnyClear takes the same mask forms as [BitsAllClear] and matches as soon as a single selected bit is 0, so
// it is the negation of [BitsAllSet] over the same mask, except that it still requires the field to be numeric or
// binData.
//
// Example:
//
//	monq.BitsAnyClear("flags", 6)
//	// bson.D{{Key: "flags", Value: bson.D{{Key: "$bitsAnyClear", Value: 6}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAnyClear/
func BitsAnyClear(field FieldPath, mask any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$bitsAnyClear", Value: mask}}}}
}

// BitsAnySet returns a filter that matches documents where at least one bit selected by mask is set in field.
//
// $bitsAnySet takes the same mask forms as [BitsAllClear] and matches as soon as a single selected bit is 1, which
// makes it the usual choice for "has any of these flags" over a flag field.
//
// Example:
//
//	monq.BitsAnySet("flags", bson.A{1, 5})
//	// bson.D{{Key: "flags", Value: bson.D{{Key: "$bitsAnySet", Value: bson.A{1, 5}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/bitsAnySet/
func BitsAnySet(field FieldPath, mask any) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$bitsAnySet", Value: mask}}}}
}
