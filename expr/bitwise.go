package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// BitAnd returns an expression yielding the bitwise AND of its arguments.
//
// $bitAnd takes integers only, int or long, and fails the aggregation on a double or a decimal. A null or missing
// argument makes the whole result null. With no arguments the result is -1, every bit set, which is the identity
// for AND. MongoDB 6.3 added these operators, and an older server rejects them outright.
//
// This is one of three bitwise families monq covers, and they do different jobs: monq.BitsAllClear and its
// siblings are query operators testing a stored field's bits, monq.BitAnd is the $bit update operator rewriting
// one, and this computes a value inside an aggregation.
//
// Example:
//
//	expr.BitAnd(expr.Field("flags"), 6)
//	// bson.D{{Key: "$bitAnd", Value: bson.A{"$flags", 6}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bitAnd/
func BitAnd(values ...any) bson.D {
	return bson.D{{Key: "$bitAnd", Value: operands(values)}}
}

// BitNot returns an expression yielding the bitwise negation of a single integer.
//
// $bitNot takes one operand rather than a list, since flipping every bit of several numbers at once means nothing.
// The operand has to be an int or a long, and the result keeps that width, so negating an int gives an int.
// MongoDB 6.3 and later.
//
// Example:
//
//	expr.BitNot(expr.Field("flags"))
//	// bson.D{{Key: "$bitNot", Value: "$flags"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bitNot/
func BitNot(value any) bson.D {
	return bson.D{{Key: "$bitNot", Value: value}}
}

// BitOr returns an expression yielding the bitwise OR of its arguments.
//
// $bitOr carries the same restrictions as [BitAnd]: integers only, null in means null out, MongoDB 6.3 and later.
// With no arguments the result is 0, the identity for OR.
//
// Example:
//
//	expr.BitOr(expr.Field("flags"), 4)
//	// bson.D{{Key: "$bitOr", Value: bson.A{"$flags", 4}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bitOr/
func BitOr(values ...any) bson.D {
	return bson.D{{Key: "$bitOr", Value: operands(values)}}
}

// BitXor returns an expression yielding the bitwise exclusive OR of its arguments.
//
// $bitXor carries the same restrictions as [BitAnd] and, like [BitOr], has 0 as its identity. MongoDB 6.3 and
// later.
//
// Example:
//
//	expr.BitXor(expr.Field("flags"), 2)
//	// bson.D{{Key: "$bitXor", Value: bson.A{"$flags", 2}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bitXor/
func BitXor(values ...any) bson.D {
	return bson.D{{Key: "$bitXor", Value: operands(values)}}
}
