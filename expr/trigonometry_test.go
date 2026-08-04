package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestTrigonometryExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Sin takes a bare expression",
			got:  expr.Sin(expr.Field("angle")),
			want: bson.D{{Key: "$sin", Value: "$angle"}},
		},
		{
			name: "Cos takes a bare expression",
			got:  expr.Cos(expr.Field("angle")),
			want: bson.D{{Key: "$cos", Value: "$angle"}},
		},
		{
			name: "Tan takes a bare expression",
			got:  expr.Tan(expr.Field("angle")),
			want: bson.D{{Key: "$tan", Value: "$angle"}},
		},
		{
			name: "Asin takes a bare expression",
			got:  expr.Asin(expr.Field("ratio")),
			want: bson.D{{Key: "$asin", Value: "$ratio"}},
		},
		{
			name: "Acos takes a bare expression",
			got:  expr.Acos(expr.Field("ratio")),
			want: bson.D{{Key: "$acos", Value: "$ratio"}},
		},
		{
			name: "Atan takes a bare expression",
			got:  expr.Atan(expr.Field("slope")),
			want: bson.D{{Key: "$atan", Value: "$slope"}},
		},
		{
			name: "Atan2 takes y before x",
			got:  expr.Atan2(expr.Field("dy"), expr.Field("dx")),
			want: bson.D{{Key: "$atan2", Value: bson.A{"$dy", "$dx"}}},
		},
		{
			name: "Sinh takes a bare expression",
			got:  expr.Sinh(expr.Field("angle")),
			want: bson.D{{Key: "$sinh", Value: "$angle"}},
		},
		{
			name: "Cosh takes a bare expression",
			got:  expr.Cosh(expr.Field("angle")),
			want: bson.D{{Key: "$cosh", Value: "$angle"}},
		},
		{
			name: "Tanh takes a bare expression",
			got:  expr.Tanh(expr.Field("angle")),
			want: bson.D{{Key: "$tanh", Value: "$angle"}},
		},
		{
			name: "Asinh takes a bare expression",
			got:  expr.Asinh(expr.Field("ratio")),
			want: bson.D{{Key: "$asinh", Value: "$ratio"}},
		},
		{
			name: "Acosh takes a bare expression",
			got:  expr.Acosh(expr.Field("ratio")),
			want: bson.D{{Key: "$acosh", Value: "$ratio"}},
		},
		{
			name: "Atanh takes a bare expression",
			got:  expr.Atanh(expr.Field("ratio")),
			want: bson.D{{Key: "$atanh", Value: "$ratio"}},
		},
		{
			name: "DegreesToRadians",
			got:  expr.DegreesToRadians(expr.Field("heading")),
			want: bson.D{{Key: "$degreesToRadians", Value: "$heading"}},
		},
		{
			name: "RadiansToDegrees",
			got:  expr.RadiansToDegrees(expr.Field("angle")),
			want: bson.D{{Key: "$radiansToDegrees", Value: "$angle"}},
		},
		{
			name: "degrees convert on the way in",
			got:  expr.Sin(expr.DegreesToRadians(expr.Field("angle"))),
			want: bson.D{{Key: "$sin", Value: bson.D{
				{Key: "$degreesToRadians", Value: "$angle"},
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

func ExampleSin() {
	e := expr.Sin(expr.DegreesToRadians(expr.Field("angle")))

	printExpr(e)
	// Output: {"$sin":{"$degreesToRadians":"$angle"}}
}

func ExampleCos() {
	e := expr.Cos(expr.Field("angle"))

	printExpr(e)
	// Output: {"$cos":"$angle"}
}

func ExampleTan() {
	e := expr.Tan(expr.Field("angle"))

	printExpr(e)
	// Output: {"$tan":"$angle"}
}

func ExampleAsin() {
	e := expr.Asin(expr.Field("ratio"))

	printExpr(e)
	// Output: {"$asin":"$ratio"}
}

func ExampleAcos() {
	e := expr.Acos(expr.Field("ratio"))

	printExpr(e)
	// Output: {"$acos":"$ratio"}
}

func ExampleAtan() {
	e := expr.Atan(expr.Field("slope"))

	printExpr(e)
	// Output: {"$atan":"$slope"}
}

func ExampleAtan2() {
	e := expr.Atan2(expr.Field("dy"), expr.Field("dx"))

	printExpr(e)
	// Output: {"$atan2":["$dy","$dx"]}
}

func ExampleSinh() {
	e := expr.Sinh(expr.Field("angle"))

	printExpr(e)
	// Output: {"$sinh":"$angle"}
}

func ExampleCosh() {
	e := expr.Cosh(expr.Field("angle"))

	printExpr(e)
	// Output: {"$cosh":"$angle"}
}

func ExampleTanh() {
	e := expr.Tanh(expr.Field("angle"))

	printExpr(e)
	// Output: {"$tanh":"$angle"}
}

func ExampleAsinh() {
	e := expr.Asinh(expr.Field("ratio"))

	printExpr(e)
	// Output: {"$asinh":"$ratio"}
}

func ExampleAcosh() {
	e := expr.Acosh(expr.Field("ratio"))

	printExpr(e)
	// Output: {"$acosh":"$ratio"}
}

func ExampleAtanh() {
	e := expr.Atanh(expr.Field("ratio"))

	printExpr(e)
	// Output: {"$atanh":"$ratio"}
}

func ExampleDegreesToRadians() {
	e := expr.DegreesToRadians(expr.Field("heading"))

	printExpr(e)
	// Output: {"$degreesToRadians":"$heading"}
}

func ExampleRadiansToDegrees() {
	e := expr.RadiansToDegrees(expr.Atan2(expr.Field("dy"), expr.Field("dx")))

	printExpr(e)
	// Output: {"$radiansToDegrees":{"$atan2":["$dy","$dx"]}}
}
