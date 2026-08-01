package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestComparisonExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Cmp",
			got:  expr.Cmp(expr.Field("score"), 50),
			want: bson.D{{Key: "$cmp", Value: bson.A{"$score", 50}}},
		},
		{
			name: "Eq against a value",
			got:  expr.Eq(expr.Field("status"), "active"),
			want: bson.D{{Key: "$eq", Value: bson.A{"$status", "active"}}},
		},
		{
			name: "Eq between two fields",
			got:  expr.Eq(expr.Field("spent"), expr.Field("budget")),
			want: bson.D{{Key: "$eq", Value: bson.A{"$spent", "$budget"}}},
		},
		{
			name: "Ne",
			got:  expr.Ne(expr.Field("status"), "banned"),
			want: bson.D{{Key: "$ne", Value: bson.A{"$status", "banned"}}},
		},
		{
			name: "Gt",
			got:  expr.Gt(expr.Field("score"), 50),
			want: bson.D{{Key: "$gt", Value: bson.A{"$score", 50}}},
		},
		{
			name: "Gte",
			got:  expr.Gte(expr.Field("age"), 18),
			want: bson.D{{Key: "$gte", Value: bson.A{"$age", 18}}},
		},
		{
			name: "Lt",
			got:  expr.Lt(expr.Field("stock"), 10),
			want: bson.D{{Key: "$lt", Value: bson.A{"$stock", 10}}},
		},
		{
			name: "Lte",
			got:  expr.Lte(expr.Field("stock"), 10),
			want: bson.D{{Key: "$lte", Value: bson.A{"$stock", 10}}},
		},
		{
			name: "nested expressions on both sides",
			got:  expr.Gt(expr.Cmp(expr.Field("a"), expr.Field("b")), 0),
			want: bson.D{{Key: "$gt", Value: bson.A{
				bson.D{{Key: "$cmp", Value: bson.A{"$a", "$b"}}},
				0,
			}}},
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

func ExampleCmp() {
	e := expr.Cmp(expr.Field("score"), 50)

	printExpr(e)
	// Output: {"$cmp":["$score",50]}
}

func ExampleEq() {
	e := expr.Eq(expr.Field("status"), "active")

	printExpr(e)
	// Output: {"$eq":["$status","active"]}
}

func ExampleNe() {
	e := expr.Ne(expr.Field("status"), "banned")

	printExpr(e)
	// Output: {"$ne":["$status","banned"]}
}

func ExampleGt() {
	e := expr.Gt(expr.Field("score"), 50)

	printExpr(e)
	// Output: {"$gt":["$score",50]}
}

func ExampleGte() {
	e := expr.Gte(expr.Field("age"), 18)

	printExpr(e)
	// Output: {"$gte":["$age",18]}
}

func ExampleLt() {
	e := expr.Lt(expr.Field("stock"), 10)

	printExpr(e)
	// Output: {"$lt":["$stock",10]}
}

func ExampleLte() {
	e := expr.Lte(expr.Field("stock"), 10)

	printExpr(e)
	// Output: {"$lte":["$stock",10]}
}
