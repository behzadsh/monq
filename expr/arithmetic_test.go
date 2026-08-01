package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestArithmeticExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Abs takes a bare expression",
			got:  expr.Abs(expr.Field("balance")),
			want: bson.D{{Key: "$abs", Value: "$balance"}},
		},
		{
			name: "Add with two operands",
			got:  expr.Add(expr.Field("price"), expr.Field("tax")),
			want: bson.D{{Key: "$add", Value: bson.A{"$price", "$tax"}}},
		},
		{
			name: "Add with several operands",
			got:  expr.Add(expr.Field("a"), expr.Field("b"), 10),
			want: bson.D{{Key: "$add", Value: bson.A{"$a", "$b", 10}}},
		},
		{
			name: "Ceil takes a bare expression",
			got:  expr.Ceil(expr.Field("rating")),
			want: bson.D{{Key: "$ceil", Value: "$rating"}},
		},
		{
			name: "Divide",
			got:  expr.Divide(expr.Field("total"), expr.Field("count")),
			want: bson.D{{Key: "$divide", Value: bson.A{"$total", "$count"}}},
		},
		{
			name: "Exp takes a bare expression",
			got:  expr.Exp(expr.Field("rate")),
			want: bson.D{{Key: "$exp", Value: "$rate"}},
		},
		{
			name: "Floor takes a bare expression",
			got:  expr.Floor(expr.Field("rating")),
			want: bson.D{{Key: "$floor", Value: "$rating"}},
		},
		{
			name: "Ln takes a bare expression",
			got:  expr.Ln(expr.Field("views")),
			want: bson.D{{Key: "$ln", Value: "$views"}},
		},
		{
			name: "Log takes a number and a base",
			got:  expr.Log(expr.Field("views"), 2),
			want: bson.D{{Key: "$log", Value: bson.A{"$views", 2}}},
		},
		{
			name: "Log10 takes a bare expression",
			got:  expr.Log10(expr.Field("views")),
			want: bson.D{{Key: "$log10", Value: "$views"}},
		},
		{
			name: "Mod computes a remainder from two operands",
			got:  expr.Mod(expr.Field("total"), 4),
			want: bson.D{{Key: "$mod", Value: bson.A{"$total", 4}}},
		},
		{
			name: "Multiply",
			got:  expr.Multiply(expr.Field("price"), expr.Field("quantity")),
			want: bson.D{{Key: "$multiply", Value: bson.A{"$price", "$quantity"}}},
		},
		{
			name: "Pow",
			got:  expr.Pow(expr.Field("side"), 2),
			want: bson.D{{Key: "$pow", Value: bson.A{"$side", 2}}},
		},
		{
			name: "Round to decimals",
			got:  expr.Round(expr.Field("price"), 2),
			want: bson.D{{Key: "$round", Value: bson.A{"$price", 2}}},
		},
		{
			name: "Round to an integer",
			got:  expr.Round(expr.Field("price"), 0),
			want: bson.D{{Key: "$round", Value: bson.A{"$price", 0}}},
		},
		{
			name: "Round to tens with a negative place",
			got:  expr.Round(expr.Field("price"), -1),
			want: bson.D{{Key: "$round", Value: bson.A{"$price", -1}}},
		},
		{
			name: "Sqrt takes a bare expression",
			got:  expr.Sqrt(expr.Field("area")),
			want: bson.D{{Key: "$sqrt", Value: "$area"}},
		},
		{
			name: "Subtract two numbers",
			got:  expr.Subtract(expr.Field("total"), expr.Field("discount")),
			want: bson.D{{Key: "$subtract", Value: bson.A{"$total", "$discount"}}},
		},
		{
			name: "Subtract two dates for the milliseconds between them",
			got:  expr.Subtract(expr.Field("finished_at"), expr.Field("started_at")),
			want: bson.D{{Key: "$subtract", Value: bson.A{"$finished_at", "$started_at"}}},
		},
		{
			name: "Trunc",
			got:  expr.Trunc(expr.Field("price"), 2),
			want: bson.D{{Key: "$trunc", Value: bson.A{"$price", 2}}},
		},
		{
			name: "Trunc to hundreds with a negative place",
			got:  expr.Trunc(expr.Field("price"), -2),
			want: bson.D{{Key: "$trunc", Value: bson.A{"$price", -2}}},
		},
		{
			name: "operators nest",
			got:  expr.Round(expr.Divide(expr.Field("total"), expr.Field("count")), 2),
			want: bson.D{{Key: "$round", Value: bson.A{
				bson.D{{Key: "$divide", Value: bson.A{"$total", "$count"}}},
				2,
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

func ExampleAbs() {
	e := expr.Abs(expr.Field("balance"))

	printExpr(e)
	// Output: {"$abs":"$balance"}
}

func ExampleAdd() {
	e := expr.Add(expr.Field("price"), expr.Field("tax"))

	printExpr(e)
	// Output: {"$add":["$price","$tax"]}
}

func ExampleCeil() {
	e := expr.Ceil(expr.Field("rating"))

	printExpr(e)
	// Output: {"$ceil":"$rating"}
}

func ExampleDivide() {
	e := expr.Divide(expr.Field("total"), expr.Field("count"))

	printExpr(e)
	// Output: {"$divide":["$total","$count"]}
}

func ExampleExp() {
	e := expr.Exp(expr.Field("rate"))

	printExpr(e)
	// Output: {"$exp":"$rate"}
}

func ExampleFloor() {
	e := expr.Floor(expr.Field("rating"))

	printExpr(e)
	// Output: {"$floor":"$rating"}
}

func ExampleLn() {
	e := expr.Ln(expr.Field("views"))

	printExpr(e)
	// Output: {"$ln":"$views"}
}

func ExampleLog() {
	e := expr.Log(expr.Field("views"), 2)

	printExpr(e)
	// Output: {"$log":["$views",2]}
}

func ExampleLog10() {
	e := expr.Log10(expr.Field("views"))

	printExpr(e)
	// Output: {"$log10":"$views"}
}

func ExampleMod() {
	e := expr.Mod(expr.Field("total"), 4)

	printExpr(e)
	// Output: {"$mod":["$total",4]}
}

func ExampleMultiply() {
	e := expr.Multiply(expr.Field("price"), expr.Field("quantity"))

	printExpr(e)
	// Output: {"$multiply":["$price","$quantity"]}
}

func ExamplePow() {
	e := expr.Pow(expr.Field("side"), 2)

	printExpr(e)
	// Output: {"$pow":["$side",2]}
}

func ExampleRound() {
	e := expr.Round(expr.Field("price"), 2)

	printExpr(e)
	// Output: {"$round":["$price",2]}
}

func ExampleSqrt() {
	e := expr.Sqrt(expr.Field("area"))

	printExpr(e)
	// Output: {"$sqrt":"$area"}
}

func ExampleSubtract() {
	e := expr.Subtract(expr.Field("total"), expr.Field("discount"))

	printExpr(e)
	// Output: {"$subtract":["$total","$discount"]}
}

func ExampleTrunc() {
	e := expr.Trunc(expr.Field("price"), 2)

	printExpr(e)
	// Output: {"$trunc":["$price",2]}
}
