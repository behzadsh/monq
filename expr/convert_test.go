package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestConvertExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Convert without fallbacks",
			got:  expr.Convert(expr.Field("legacy_id"), "objectId"),
			want: bson.D{
				{
					Key: "$convert",
					Value: bson.D{
						{Key: "input", Value: "$legacy_id"},
						{Key: "to", Value: "objectId"},
					},
				},
			},
		},
		{
			name: "Convert with both fallbacks",
			got: expr.Convert(
				expr.Field("legacy_id"), "objectId",
				expr.ConvertOnError(nil), expr.ConvertOnNull(nil),
			),
			want: bson.D{
				{
					Key: "$convert",
					Value: bson.D{
						{Key: "input", Value: "$legacy_id"},
						{Key: "to", Value: "objectId"},
						{Key: "onError", Value: nil},
						{Key: "onNull", Value: nil},
					},
				},
			},
		},
		{
			name: "IsNumber takes a bare expression",
			got:  expr.IsNumber(expr.Field("score")),
			want: bson.D{{Key: "$isNumber", Value: "$score"}},
		},
		{
			name: "Type reports a type name",
			got:  expr.Type(expr.Field("legacy_id")),
			want: bson.D{{Key: "$type", Value: "$legacy_id"}},
		},
		{
			name: "ToBool",
			got:  expr.ToBool(expr.Field("flag")),
			want: bson.D{{Key: "$toBool", Value: "$flag"}},
		},
		{
			name: "ToDate",
			got:  expr.ToDate(expr.Field("created_on")),
			want: bson.D{{Key: "$toDate", Value: "$created_on"}},
		},
		{
			name: "ToDecimal",
			got:  expr.ToDecimal(expr.Field("price")),
			want: bson.D{{Key: "$toDecimal", Value: "$price"}},
		},
		{
			name: "ToDouble",
			got:  expr.ToDouble(expr.Field("weight")),
			want: bson.D{{Key: "$toDouble", Value: "$weight"}},
		},
		{
			name: "ToInt",
			got:  expr.ToInt(expr.Field("quantity")),
			want: bson.D{{Key: "$toInt", Value: "$quantity"}},
		},
		{
			name: "ToLong",
			got:  expr.ToLong(expr.Field("quantity")),
			want: bson.D{{Key: "$toLong", Value: "$quantity"}},
		},
		{
			name: "ToObjectID emits the operator MongoDB spells with a lowercase d",
			got:  expr.ToObjectID(expr.Field("legacy_id")),
			want: bson.D{{Key: "$toObjectId", Value: "$legacy_id"}},
		},
		{
			name: "ToString",
			got:  expr.ToString(expr.Field("quantity")),
			want: bson.D{{Key: "$toString", Value: "$quantity"}},
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

func ExampleConvert() {
	e := expr.Convert(expr.Field("legacy_id"), "objectId", expr.ConvertOnError(nil))

	printExpr(e)
	// Output: {"$convert":{"input":"$legacy_id","to":"objectId","onError":null}}
}

func ExampleIsNumber() {
	e := expr.IsNumber(expr.Field("score"))

	printExpr(e)
	// Output: {"$isNumber":"$score"}
}

func ExampleType() {
	e := expr.Type(expr.Field("legacy_id"))

	printExpr(e)
	// Output: {"$type":"$legacy_id"}
}

func ExampleToBool() {
	e := expr.ToBool(expr.Field("flag"))

	printExpr(e)
	// Output: {"$toBool":"$flag"}
}

func ExampleToDate() {
	e := expr.ToDate(expr.Field("created_on"))

	printExpr(e)
	// Output: {"$toDate":"$created_on"}
}

func ExampleToDecimal() {
	e := expr.ToDecimal(expr.Field("price"))

	printExpr(e)
	// Output: {"$toDecimal":"$price"}
}

func ExampleToDouble() {
	e := expr.ToDouble(expr.Field("weight"))

	printExpr(e)
	// Output: {"$toDouble":"$weight"}
}

func ExampleToInt() {
	e := expr.ToInt(expr.Field("quantity"))

	printExpr(e)
	// Output: {"$toInt":"$quantity"}
}

func ExampleToLong() {
	e := expr.ToLong(expr.Field("quantity"))

	printExpr(e)
	// Output: {"$toLong":"$quantity"}
}

func ExampleToObjectID() {
	e := expr.ToObjectID(expr.Field("legacy_id"))

	printExpr(e)
	// Output: {"$toObjectId":"$legacy_id"}
}

func ExampleToString() {
	e := expr.ToString(expr.Field("quantity"))

	printExpr(e)
	// Output: {"$toString":"$quantity"}
}

func ExampleConvertOnError() {
	e := expr.Convert(expr.Field("legacy_id"), "objectId", expr.ConvertOnError(nil))

	printExpr(e)
	// Output: {"$convert":{"input":"$legacy_id","to":"objectId","onError":null}}
}

func ExampleConvertOnNull() {
	e := expr.Convert(expr.Field("quantity"), "int", expr.ConvertOnNull(0))

	printExpr(e)
	// Output: {"$convert":{"input":"$quantity","to":"int","onNull":0}}
}
