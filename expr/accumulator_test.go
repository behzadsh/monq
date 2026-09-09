package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestAccumulatorExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Sum with one argument accumulates over a group",
			got:  expr.Sum(expr.Field("amount")),
			want: bson.D{{Key: "$sum", Value: "$amount"}},
		},
		{
			name: "Sum with several arguments adds within a document",
			got:  expr.Sum(expr.Field("price"), expr.Field("tax")),
			want: bson.D{{Key: "$sum", Value: bson.A{"$price", "$tax"}}},
		},
		{
			name: "Sum of a constant counts documents",
			got:  expr.Sum(1),
			want: bson.D{{Key: "$sum", Value: 1}},
		},
		{
			name: "Avg with one argument",
			got:  expr.Avg(expr.Field("score")),
			want: bson.D{{Key: "$avg", Value: "$score"}},
		},
		{
			name: "Avg with several arguments",
			got:  expr.Avg(expr.Field("a"), expr.Field("b")),
			want: bson.D{{Key: "$avg", Value: bson.A{"$a", "$b"}}},
		},
		{
			name: "Max with one argument",
			got:  expr.Max(expr.Field("score")),
			want: bson.D{{Key: "$max", Value: "$score"}},
		},
		{
			name: "Min with one argument",
			got:  expr.Min(expr.Field("score")),
			want: bson.D{{Key: "$min", Value: "$score"}},
		},
		{
			name: "StdDevPop",
			got:  expr.StdDevPop(expr.Field("score")),
			want: bson.D{{Key: "$stdDevPop", Value: "$score"}},
		},
		{
			name: "StdDevSamp",
			got:  expr.StdDevSamp(expr.Field("score")),
			want: bson.D{{Key: "$stdDevSamp", Value: "$score"}},
		},
		{
			name: "Push",
			got:  expr.Push(expr.Field("name")),
			want: bson.D{{Key: "$push", Value: "$name"}},
		},
		{
			name: "AddToSet",
			got:  expr.AddToSet(expr.Field("category")),
			want: bson.D{{Key: "$addToSet", Value: "$category"}},
		},
		{
			name: "Count takes an empty document",
			got:  expr.Count(),
			want: bson.D{{Key: "$count", Value: bson.D{}}},
		},
		{
			name: "Top",
			got:  expr.Top(bson.D{{Key: "score", Value: -1}}, expr.Field("name")),
			want: bson.D{
				{
					Key: "$top",
					Value: bson.D{
						{Key: "sortBy", Value: bson.D{{Key: "score", Value: -1}}},
						{Key: "output", Value: "$name"},
					},
				},
			},
		},
		{
			name: "TopN puts n first",
			got:  expr.TopN(3, bson.D{{Key: "score", Value: -1}}, expr.Field("name")),
			want: bson.D{
				{
					Key: "$topN",
					Value: bson.D{
						{Key: "n", Value: 3},
						{Key: "sortBy", Value: bson.D{{Key: "score", Value: -1}}},
						{Key: "output", Value: "$name"},
					},
				},
			},
		},
		{
			name: "Bottom",
			got:  expr.Bottom(bson.D{{Key: "score", Value: 1}}, expr.Field("name")),
			want: bson.D{
				{
					Key: "$bottom",
					Value: bson.D{
						{Key: "sortBy", Value: bson.D{{Key: "score", Value: 1}}},
						{Key: "output", Value: "$name"},
					},
				},
			},
		},
		{
			name: "BottomN puts n first",
			got:  expr.BottomN(3, bson.D{{Key: "score", Value: 1}}, expr.Field("name")),
			want: bson.D{
				{
					Key: "$bottomN",
					Value: bson.D{
						{Key: "n", Value: 3},
						{Key: "sortBy", Value: bson.D{{Key: "score", Value: 1}}},
						{Key: "output", Value: "$name"},
					},
				},
			},
		},
		{
			name: "Median",
			got:  expr.Median(expr.Field("score"), "approximate"),
			want: bson.D{
				{
					Key: "$median",
					Value: bson.D{
						{Key: "input", Value: "$score"},
						{Key: "method", Value: "approximate"},
					},
				},
			},
		},
		{
			name: "Percentile",
			got:  expr.Percentile(expr.Field("score"), bson.A{0.5, 0.95}, "approximate"),
			want: bson.D{
				{
					Key: "$percentile",
					Value: bson.D{
						{Key: "input", Value: "$score"},
						{Key: "p", Value: bson.A{0.5, 0.95}},
						{Key: "method", Value: "approximate"},
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

func TestTopNLeavesItsSortDocumentUntouched(t *testing.T) {
	sortBy := bson.D{{Key: "score", Value: -1}}

	expr.TopN(3, sortBy, expr.Field("name"))
	expr.BottomN(3, sortBy, expr.Field("name"))

	want := bson.D{{Key: "score", Value: -1}}
	if !reflect.DeepEqual(sortBy, want) {
		t.Fatalf("sortBy = %v, want %v unchanged", sortBy, want)
	}
}

func ExampleSum() {
	e := expr.Sum(expr.Field("amount"))

	printExpr(e)
	// Output: {"$sum":"$amount"}
}

func ExampleAvg() {
	e := expr.Avg(expr.Field("score"))

	printExpr(e)
	// Output: {"$avg":"$score"}
}

func ExampleMax() {
	e := expr.Max(expr.Field("score"))

	printExpr(e)
	// Output: {"$max":"$score"}
}

func ExampleMin() {
	e := expr.Min(expr.Field("score"))

	printExpr(e)
	// Output: {"$min":"$score"}
}

func ExampleStdDevPop() {
	e := expr.StdDevPop(expr.Field("score"))

	printExpr(e)
	// Output: {"$stdDevPop":"$score"}
}

func ExampleStdDevSamp() {
	e := expr.StdDevSamp(expr.Field("score"))

	printExpr(e)
	// Output: {"$stdDevSamp":"$score"}
}

func ExamplePush() {
	e := expr.Push(expr.Field("name"))

	printExpr(e)
	// Output: {"$push":"$name"}
}

func ExampleAddToSet() {
	e := expr.AddToSet(expr.Field("category"))

	printExpr(e)
	// Output: {"$addToSet":"$category"}
}

func ExampleCount() {
	e := expr.Count()

	printExpr(e)
	// Output: {"$count":{}}
}

func ExampleTop() {
	e := expr.Top(bson.D{{Key: "score", Value: -1}}, expr.Field("name"))

	printExpr(e)
	// Output: {"$top":{"sortBy":{"score":-1},"output":"$name"}}
}

func ExampleTopN() {
	e := expr.TopN(3, bson.D{{Key: "score", Value: -1}}, expr.Field("name"))

	printExpr(e)
	// Output: {"$topN":{"n":3,"sortBy":{"score":-1},"output":"$name"}}
}

func ExampleBottom() {
	e := expr.Bottom(bson.D{{Key: "score", Value: 1}}, expr.Field("name"))

	printExpr(e)
	// Output: {"$bottom":{"sortBy":{"score":1},"output":"$name"}}
}

func ExampleBottomN() {
	e := expr.BottomN(3, bson.D{{Key: "score", Value: 1}}, expr.Field("name"))

	printExpr(e)
	// Output: {"$bottomN":{"n":3,"sortBy":{"score":1},"output":"$name"}}
}

func ExampleMedian() {
	e := expr.Median(expr.Field("score"), "approximate")

	printExpr(e)
	// Output: {"$median":{"input":"$score","method":"approximate"}}
}

func ExamplePercentile() {
	e := expr.Percentile(expr.Field("score"), bson.A{0.5, 0.95}, "approximate")

	printExpr(e)
	// Output: {"$percentile":{"input":"$score","p":[0.5,0.95],"method":"approximate"}}
}
