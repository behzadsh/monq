package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestBitwiseExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "BitAnd over two operands",
			got:  expr.BitAnd(expr.Field("flags"), 6),
			want: bson.D{{Key: "$bitAnd", Value: bson.A{"$flags", 6}}},
		},
		{
			name: "BitAnd over three operands",
			got:  expr.BitAnd(expr.Field("a"), expr.Field("b"), 1),
			want: bson.D{{Key: "$bitAnd", Value: bson.A{"$a", "$b", 1}}},
		},
		{
			name: "BitAnd with no operands",
			got:  expr.BitAnd(),
			want: bson.D{{Key: "$bitAnd", Value: bson.A{}}},
		},
		{
			name: "BitOr",
			got:  expr.BitOr(expr.Field("flags"), 4),
			want: bson.D{{Key: "$bitOr", Value: bson.A{"$flags", 4}}},
		},
		{
			name: "BitXor",
			got:  expr.BitXor(expr.Field("flags"), 2),
			want: bson.D{{Key: "$bitXor", Value: bson.A{"$flags", 2}}},
		},
		{
			name: "BitNot takes a single bare operand",
			got:  expr.BitNot(expr.Field("flags")),
			want: bson.D{{Key: "$bitNot", Value: "$flags"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func ExampleBitAnd() {
	e := expr.BitAnd(expr.Field("flags"), 6)

	printExpr(e)
	// Output: {"$bitAnd":["$flags",6]}
}

func ExampleBitOr() {
	e := expr.BitOr(expr.Field("flags"), 4)

	printExpr(e)
	// Output: {"$bitOr":["$flags",4]}
}

func ExampleBitXor() {
	e := expr.BitXor(expr.Field("flags"), 2)

	printExpr(e)
	// Output: {"$bitXor":["$flags",2]}
}

func ExampleBitNot() {
	e := expr.BitNot(expr.Field("flags"))

	printExpr(e)
	// Output: {"$bitNot":"$flags"}
}
