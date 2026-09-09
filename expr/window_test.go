package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestWindowExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Rank takes an empty document",
			got:  expr.Rank(),
			want: bson.D{{Key: "$rank", Value: bson.D{}}},
		},
		{
			name: "DenseRank takes an empty document",
			got:  expr.DenseRank(),
			want: bson.D{{Key: "$denseRank", Value: bson.D{}}},
		},
		{
			name: "DocumentNumber takes an empty document",
			got:  expr.DocumentNumber(),
			want: bson.D{{Key: "$documentNumber", Value: bson.D{}}},
		},
		{
			name: "Shift without a default",
			got:  expr.Shift(expr.Field("price"), -1),
			want: bson.D{
				{
					Key: "$shift",
					Value: bson.D{
						{Key: "output", Value: "$price"},
						{Key: "by", Value: -1},
					},
				},
			},
		},
		{
			name: "Shift forward with a default",
			got:  expr.Shift(expr.Field("price"), 1, expr.ShiftDefault(0)),
			want: bson.D{
				{
					Key: "$shift",
					Value: bson.D{
						{Key: "output", Value: "$price"},
						{Key: "by", Value: 1},
						{Key: "default", Value: 0},
					},
				},
			},
		},
		{
			name: "Locf takes a bare expression",
			got:  expr.Locf(expr.Field("price")),
			want: bson.D{{Key: "$locf", Value: "$price"}},
		},
		{
			name: "LinearFill takes a bare expression",
			got:  expr.LinearFill(expr.Field("price")),
			want: bson.D{{Key: "$linearFill", Value: "$price"}},
		},
		{
			name: "Derivative without a unit",
			got:  expr.Derivative(expr.Field("odometer")),
			want: bson.D{{Key: "$derivative", Value: bson.D{{Key: "input", Value: "$odometer"}}}},
		},
		{
			name: "Derivative per hour",
			got:  expr.Derivative(expr.Field("odometer"), expr.TimeUnit("hour")),
			want: bson.D{
				{
					Key: "$derivative",
					Value: bson.D{
						{Key: "input", Value: "$odometer"},
						{Key: "unit", Value: "hour"},
					},
				},
			},
		},
		{
			name: "Integral per hour",
			got:  expr.Integral(expr.Field("power"), expr.TimeUnit("hour")),
			want: bson.D{
				{
					Key: "$integral",
					Value: bson.D{
						{Key: "input", Value: "$power"},
						{Key: "unit", Value: "hour"},
					},
				},
			},
		},
		{
			name: "ExpMovingAvgN uses a capital N as MongoDB spells it",
			got:  expr.ExpMovingAvgN(expr.Field("price"), 5),
			want: bson.D{
				{
					Key: "$expMovingAvg",
					Value: bson.D{
						{Key: "input", Value: "$price"},
						{Key: "N", Value: 5},
					},
				},
			},
		},
		{
			name: "ExpMovingAvgAlpha",
			got:  expr.ExpMovingAvgAlpha(expr.Field("price"), 0.3),
			want: bson.D{
				{
					Key: "$expMovingAvg",
					Value: bson.D{
						{Key: "input", Value: "$price"},
						{Key: "alpha", Value: 0.3},
					},
				},
			},
		},
		{
			name: "CovariancePop",
			got:  expr.CovariancePop(expr.Field("x"), expr.Field("y")),
			want: bson.D{{Key: "$covariancePop", Value: bson.A{"$x", "$y"}}},
		},
		{
			name: "CovarianceSamp",
			got:  expr.CovarianceSamp(expr.Field("x"), expr.Field("y")),
			want: bson.D{{Key: "$covarianceSamp", Value: bson.A{"$x", "$y"}}},
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

func ExampleRank() {
	e := expr.Rank()

	printExpr(e)
	// Output: {"$rank":{}}
}

func ExampleDenseRank() {
	e := expr.DenseRank()

	printExpr(e)
	// Output: {"$denseRank":{}}
}

func ExampleDocumentNumber() {
	e := expr.DocumentNumber()

	printExpr(e)
	// Output: {"$documentNumber":{}}
}

func ExampleShift() {
	e := expr.Shift(expr.Field("price"), -1, expr.ShiftDefault(0))

	printExpr(e)
	// Output: {"$shift":{"output":"$price","by":-1,"default":0}}
}

func ExampleShiftDefault() {
	e := expr.Shift(expr.Field("price"), 1, expr.ShiftDefault("none"))

	printExpr(e)
	// Output: {"$shift":{"output":"$price","by":1,"default":"none"}}
}

func ExampleLocf() {
	e := expr.Locf(expr.Field("price"))

	printExpr(e)
	// Output: {"$locf":"$price"}
}

func ExampleLinearFill() {
	e := expr.LinearFill(expr.Field("price"))

	printExpr(e)
	// Output: {"$linearFill":"$price"}
}

func ExampleDerivative() {
	e := expr.Derivative(expr.Field("odometer"), expr.TimeUnit("hour"))

	printExpr(e)
	// Output: {"$derivative":{"input":"$odometer","unit":"hour"}}
}

func ExampleTimeUnit() {
	e := expr.Integral(expr.Field("power"), expr.TimeUnit("hour"))

	printExpr(e)
	// Output: {"$integral":{"input":"$power","unit":"hour"}}
}

func ExampleIntegral() {
	e := expr.Integral(expr.Field("power"))

	printExpr(e)
	// Output: {"$integral":{"input":"$power"}}
}

func ExampleExpMovingAvgN() {
	e := expr.ExpMovingAvgN(expr.Field("price"), 5)

	printExpr(e)
	// Output: {"$expMovingAvg":{"input":"$price","N":5}}
}

func ExampleExpMovingAvgAlpha() {
	e := expr.ExpMovingAvgAlpha(expr.Field("price"), 0.3)

	printExpr(e)
	// Output: {"$expMovingAvg":{"input":"$price","alpha":0.3}}
}

func ExampleCovariancePop() {
	e := expr.CovariancePop(expr.Field("x"), expr.Field("y"))

	printExpr(e)
	// Output: {"$covariancePop":["$x","$y"]}
}

func ExampleCovarianceSamp() {
	e := expr.CovarianceSamp(expr.Field("x"), expr.Field("y"))

	printExpr(e)
	// Output: {"$covarianceSamp":["$x","$y"]}
}
