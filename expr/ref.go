package expr

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// Field returns the aggregation reference to a document field.
//
// Aggregation expressions name a field with a string that starts with a dollar sign, so the reference to age is
// "$age". Field builds that string from a monq.FieldPath, which spells the intent out at the call site and lets a
// generated path constant be used directly. Writing "$age" by hand is exactly equivalent; both forms appear
// throughout the MongoDB documentation.
//
// The rule works in the other direction too: a string starting with a dollar sign is always read as a reference,
// which is what [Literal] exists to escape. A field that does not exist in a document resolves to missing, and
// most operators treat that as null.
//
// Example:
//
//	expr.Field("profile.age")
//	// "$profile.age"
//
// MongoDB docs: https://www.mongodb.com/docs/manual/meta/aggregation-quick-reference/#expressions
func Field(f monq.FieldPath) string {
	return "$" + string(f)
}

// Literal returns an expression that evaluates to value exactly as given.
//
// $literal stops MongoDB from interpreting what it wraps, which matters for strings that begin with a dollar sign:
// "$total" is a field reference, while Literal("$total") is the four-character string. It also protects a document
// that would otherwise look like an operator expression, such as one whose only key starts with a dollar sign.
// Values that cannot be mistaken for either need no wrapping.
//
// Example:
//
//	expr.Literal("$total")
//	// bson.D{{Key: "$literal", Value: "$total"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/literal/
func Literal(value any) bson.D {
	return bson.D{{Key: "$literal", Value: value}}
}

// operands renders the arguments of an expression operator as the BSON array MongoDB expects.
func operands(values []any) bson.A {
	arr := make(bson.A, len(values))
	copy(arr, values)

	return arr
}
