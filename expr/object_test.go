package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestObjectExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "MergeObjects",
			got:  expr.MergeObjects(expr.Field("defaults"), expr.Field("overrides")),
			want: bson.D{{Key: "$mergeObjects", Value: bson.A{"$defaults", "$overrides"}}},
		},
		{
			name: "MergeObjects with one document",
			got:  expr.MergeObjects(expr.Field("settings")),
			want: bson.D{{Key: "$mergeObjects", Value: bson.A{"$settings"}}},
		},
		{
			name: "GetField reads from the current document",
			got:  expr.GetField("price.usd", "$$CURRENT"),
			want: bson.D{
				{
					Key: "$getField",
					Value: bson.D{
						{Key: "field", Value: "price.usd"},
						{Key: "input", Value: "$$CURRENT"},
					},
				},
			},
		},
		{
			name: "GetField reads from another document",
			got:  expr.GetField("usd", expr.Field("price")),
			want: bson.D{
				{
					Key: "$getField",
					Value: bson.D{
						{Key: "field", Value: "usd"},
						{Key: "input", Value: "$price"},
					},
				},
			},
		},
		{
			name: "SetField",
			got:  expr.SetField("price.usd", "$$CURRENT", 42),
			want: bson.D{
				{
					Key: "$setField",
					Value: bson.D{
						{Key: "field", Value: "price.usd"},
						{Key: "input", Value: "$$CURRENT"},
						{Key: "value", Value: 42},
					},
				},
			},
		},
		{
			name: "UnsetField",
			got:  expr.UnsetField("price.usd", "$$CURRENT"),
			want: bson.D{
				{
					Key: "$unsetField",
					Value: bson.D{
						{Key: "field", Value: "price.usd"},
						{Key: "input", Value: "$$CURRENT"},
					},
				},
			},
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

func ExampleMergeObjects() {
	e := expr.MergeObjects(expr.Field("defaults"), expr.Field("overrides"))

	printExpr(e)
	// Output: {"$mergeObjects":["$defaults","$overrides"]}
}

func ExampleGetField() {
	e := expr.GetField("price.usd", "$$CURRENT")

	printExpr(e)
	// Output: {"$getField":{"field":"price.usd","input":"$$CURRENT"}}
}

func ExampleSetField() {
	e := expr.SetField("price.usd", "$$CURRENT", 42)

	printExpr(e)
	// Output: {"$setField":{"field":"price.usd","input":"$$CURRENT","value":42}}
}

func ExampleUnsetField() {
	e := expr.UnsetField("price.usd", "$$CURRENT")

	printExpr(e)
	// Output: {"$unsetField":{"field":"price.usd","input":"$$CURRENT"}}
}
