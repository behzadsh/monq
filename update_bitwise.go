package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// BitAnd returns an update that replaces field with its bitwise AND against operand.
//
// $bit works on integer fields only: int32 and int64 are fine, a double or a decimal is a server error, and a
// missing field is left alone rather than created. The operand must be an integer too, and the Go driver encodes
// an untyped int as int32 when it fits, so pass an int64 explicitly when the stored field is a long and the value
// is meant to be one. Note the inner key is a bare "and" with no $ prefix, which is how MongoDB spells the $bit
// sub-operators. Combine it with other operators through [Update]; two operator documents concatenated by hand
// keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.BitAnd("flags", 6)
//	// bson.D{{Key: "$bit", Value: bson.D{{Key: "flags", Value: bson.D{{Key: "and", Value: 6}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/bit/
func BitAnd(field FieldPath, operand any) bson.D {
	return bitUpdate(field, "and", operand)
}

// BitOr returns an update that replaces field with its bitwise OR against operand.
//
// It carries the same restrictions as [BitAnd]: integer fields and operands only, missing fields untouched, and a
// bare "or" as the inner key. Setting flag bits without disturbing the rest is what it is usually for. Combine it
// with other operators through [Update]; two operator documents concatenated by hand keep two separate keys, which
// MongoDB does not merge.
//
// Example:
//
//	monq.BitOr("flags", 4)
//	// bson.D{{Key: "$bit", Value: bson.D{{Key: "flags", Value: bson.D{{Key: "or", Value: 4}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/bit/
func BitOr(field FieldPath, operand any) bson.D {
	return bitUpdate(field, "or", operand)
}

// BitXor returns an update that replaces field with its bitwise XOR against operand.
//
// It carries the same restrictions as [BitAnd] and uses a bare "xor" as the inner key. Toggling flag bits is what
// it is usually for. Combine it with other operators through [Update]; two operator documents concatenated by hand
// keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.BitXor("flags", 2)
//	// bson.D{{Key: "$bit", Value: bson.D{{Key: "flags", Value: bson.D{{Key: "xor", Value: 2}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/bit/
func BitXor(field FieldPath, operand any) bson.D {
	return bitUpdate(field, "xor", operand)
}

// bitUpdate builds the $bit document for one of the bare "and", "or", and "xor" sub-operators.
func bitUpdate(field FieldPath, op string, operand any) bson.D {
	value := bson.D{{Key: op, Value: operand}}

	return bson.D{{Key: "$bit", Value: bson.D{{Key: string(field), Value: value}}}}
}
