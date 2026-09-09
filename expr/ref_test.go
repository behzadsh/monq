package expr_test

import (
	"fmt"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
)

func TestField(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		want  string
	}{
		{
			name:  "top-level field",
			field: "age",
			want:  "$age",
		},
		{
			name:  "dotted path",
			field: "profile.age",
			want:  "$profile.age",
		},
		{
			name:  "empty path",
			field: "",
			want:  "$",
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := expr.Field(tt.field); got != tt.want {
					t.Fatalf("Field() = %q, want %q", got, tt.want)
				}
			},
		)
	}
}

func TestLiteral(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  bson.D
	}{
		{
			name:  "string that would otherwise be a field reference",
			value: "$total",
			want:  bson.D{{Key: "$literal", Value: "$total"}},
		},
		{
			name:  "number",
			value: 1,
			want:  bson.D{{Key: "$literal", Value: 1}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := expr.Literal(tt.value); !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("Literal() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func ExampleField() {
	reference := expr.Field("profile.age")

	fmt.Println(reference)
	// Output: $profile.age
}

func ExampleLiteral() {
	e := expr.Literal("$total")

	printExpr(e)
	// Output: {"$literal":"$total"}
}
