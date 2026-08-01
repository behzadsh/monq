package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "And with two conditions",
			got:  expr.And(expr.Gte(expr.Field("age"), 18), expr.Eq(expr.Field("status"), "active")),
			want: bson.D{{Key: "$and", Value: bson.A{
				bson.D{{Key: "$gte", Value: bson.A{"$age", 18}}},
				bson.D{{Key: "$eq", Value: bson.A{"$status", "active"}}},
			}}},
		},
		{
			name: "And with none",
			got:  expr.And(),
			want: bson.D{{Key: "$and", Value: bson.A{}}},
		},
		{
			name: "Or with two conditions",
			got:  expr.Or(expr.Gte(expr.Field("score"), 90), expr.Eq(expr.Field("staff"), true)),
			want: bson.D{{Key: "$or", Value: bson.A{
				bson.D{{Key: "$gte", Value: bson.A{"$score", 90}}},
				bson.D{{Key: "$eq", Value: bson.A{"$staff", true}}},
			}}},
		},
		{
			name: "Not wraps a single expression in an array",
			got:  expr.Not(expr.Eq(expr.Field("status"), "banned")),
			want: bson.D{{Key: "$not", Value: bson.A{
				bson.D{{Key: "$eq", Value: bson.A{"$status", "banned"}}},
			}}},
		},
		{
			name: "Not over a bare field reference",
			got:  expr.Not(expr.Field("archived")),
			want: bson.D{{Key: "$not", Value: bson.A{"$archived"}}},
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

func ExampleAnd() {
	e := expr.And(expr.Gte(expr.Field("age"), 18), expr.Eq(expr.Field("status"), "active"))

	printExpr(e)
	// Output: {"$and":[{"$gte":["$age",18]},{"$eq":["$status","active"]}]}
}

func ExampleOr() {
	e := expr.Or(expr.Gte(expr.Field("score"), 90), expr.Eq(expr.Field("staff"), true))

	printExpr(e)
	// Output: {"$or":[{"$gte":["$score",90]},{"$eq":["$staff",true]}]}
}

func ExampleNot() {
	e := expr.Not(expr.Eq(expr.Field("status"), "banned"))

	printExpr(e)
	// Output: {"$not":[{"$eq":["$status","banned"]}]}
}
