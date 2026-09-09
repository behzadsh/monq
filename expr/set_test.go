package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestSetExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "AllElementsTrue wraps its argument",
			got:  expr.AllElementsTrue(expr.Field("checks")),
			want: bson.D{{Key: "$allElementsTrue", Value: bson.A{"$checks"}}},
		},
		{
			name: "AnyElementTrue wraps its argument",
			got:  expr.AnyElementTrue(expr.Field("checks")),
			want: bson.D{{Key: "$anyElementTrue", Value: bson.A{"$checks"}}},
		},
		{
			name: "SetDifference takes two arrays",
			got:  expr.SetDifference(expr.Field("tags"), expr.Field("banned_tags")),
			want: bson.D{{Key: "$setDifference", Value: bson.A{"$tags", "$banned_tags"}}},
		},
		{
			name: "SetIsSubset takes two arrays",
			got:  expr.SetIsSubset(expr.Field("required_tags"), expr.Field("tags")),
			want: bson.D{{Key: "$setIsSubset", Value: bson.A{"$required_tags", "$tags"}}},
		},
		{
			name: "SetEquals",
			got:  expr.SetEquals(expr.Field("tags"), expr.Field("expected_tags")),
			want: bson.D{{Key: "$setEquals", Value: bson.A{"$tags", "$expected_tags"}}},
		},
		{
			name: "SetIntersection over three arrays",
			got:  expr.SetIntersection(expr.Field("a"), expr.Field("b"), expr.Field("c")),
			want: bson.D{{Key: "$setIntersection", Value: bson.A{"$a", "$b", "$c"}}},
		},
		{
			name: "SetUnion",
			got:  expr.SetUnion(expr.Field("tags"), expr.Field("extra_tags")),
			want: bson.D{{Key: "$setUnion", Value: bson.A{"$tags", "$extra_tags"}}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if !reflect.DeepEqual(tt.got, tt.want) {
					t.Fatalf("got %v, want %v", tt.got, tt.want)
				}
			},
		)
	}
}

func ExampleAllElementsTrue() {
	e := expr.AllElementsTrue(expr.Field("checks"))

	printExpr(e)
	// Output: {"$allElementsTrue":["$checks"]}
}

func ExampleAnyElementTrue() {
	e := expr.AnyElementTrue(expr.Field("checks"))

	printExpr(e)
	// Output: {"$anyElementTrue":["$checks"]}
}

func ExampleSetDifference() {
	e := expr.SetDifference(expr.Field("tags"), expr.Field("banned_tags"))

	printExpr(e)
	// Output: {"$setDifference":["$tags","$banned_tags"]}
}

func ExampleSetEquals() {
	e := expr.SetEquals(expr.Field("tags"), expr.Field("expected_tags"))

	printExpr(e)
	// Output: {"$setEquals":["$tags","$expected_tags"]}
}

func ExampleSetIntersection() {
	e := expr.SetIntersection(expr.Field("tags"), expr.Field("featured_tags"))

	printExpr(e)
	// Output: {"$setIntersection":["$tags","$featured_tags"]}
}

func ExampleSetIsSubset() {
	e := expr.SetIsSubset(expr.Field("required_tags"), expr.Field("tags"))

	printExpr(e)
	// Output: {"$setIsSubset":["$required_tags","$tags"]}
}

func ExampleSetUnion() {
	e := expr.SetUnion(expr.Field("tags"), expr.Field("extra_tags"))

	printExpr(e)
	// Output: {"$setUnion":["$tags","$extra_tags"]}
}
