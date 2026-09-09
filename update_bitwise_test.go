package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestUpdateBitwiseOperators(t *testing.T) {
	tests := []struct {
		name    string
		got     bson.D
		field   monq.FieldPath
		subOp   string
		operand any
	}{
		{
			name:    "BitAnd clears bits",
			got:     monq.BitAnd("flags", 6),
			field:   "flags",
			subOp:   "and",
			operand: 6,
		},
		{
			name:    "BitOr sets bits",
			got:     monq.BitOr("flags", 4),
			field:   "flags",
			subOp:   "or",
			operand: 4,
		},
		{
			name:    "BitXor toggles bits",
			got:     monq.BitXor("perms.bits", 2),
			field:   "perms.bits",
			subOp:   "xor",
			operand: 2,
		},
		{
			name:    "BitOr keeps an explicit int64 operand",
			got:     monq.BitOr("flags", int64(4)),
			field:   "flags",
			subOp:   "or",
			operand: int64(4),
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				// The sub-operator key is bare, "and" rather than "$and", which is how MongoDB spells it.
				sub := bson.D{{Key: tt.subOp, Value: tt.operand}}
				want := bson.D{{Key: "$bit", Value: bson.D{{Key: string(tt.field), Value: sub}}}}

				if !reflect.DeepEqual(tt.got, want) {
					t.Fatalf("got %v, want %v", tt.got, want)
				}
			},
		)
	}
}

func ExampleBitAnd() {
	update := monq.BitAnd("flags", 6)

	printFilter(update)
	// Output: {"$bit":{"flags":{"and":6}}}
}

func ExampleBitOr() {
	update := monq.BitOr("flags", 4)

	printFilter(update)
	// Output: {"$bit":{"flags":{"or":4}}}
}

func ExampleBitXor() {
	update := monq.BitXor("flags", 2)

	printFilter(update)
	// Output: {"$bit":{"flags":{"xor":2}}}
}
