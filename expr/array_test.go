package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestArrayExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "ArrayElemAt",
			got:  expr.ArrayElemAt(expr.Field("scores"), 0),
			want: bson.D{{Key: "$arrayElemAt", Value: bson.A{"$scores", 0}}},
		},
		{
			name: "ArrayToObject takes a bare expression",
			got:  expr.ArrayToObject(expr.Field("settings")),
			want: bson.D{{Key: "$arrayToObject", Value: "$settings"}},
		},
		{
			name: "ObjectToArray takes a bare expression",
			got:  expr.ObjectToArray(expr.Field("settings")),
			want: bson.D{{Key: "$objectToArray", Value: "$settings"}},
		},
		{
			name: "ConcatArrays",
			got:  expr.ConcatArrays(expr.Field("tags"), expr.Field("extra_tags")),
			want: bson.D{{Key: "$concatArrays", Value: bson.A{"$tags", "$extra_tags"}}},
		},
		{
			name: "Filter binds the element to $$this",
			got:  expr.Filter(expr.Field("scores"), expr.Gte("$$this", 80)),
			want: bson.D{{Key: "$filter", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "cond", Value: bson.D{{Key: "$gte", Value: bson.A{"$$this", 80}}}},
			}}},
		},
		{
			name: "Filter with a name and a limit",
			got: expr.Filter(expr.Field("scores"), expr.Gte("$$score", 80),
				expr.FilterAs("score"), expr.FilterLimit(3)),
			want: bson.D{{Key: "$filter", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "cond", Value: bson.D{{Key: "$gte", Value: bson.A{"$$score", 80}}}},
				{Key: "as", Value: "score"},
				{Key: "limit", Value: 3},
			}}},
		},
		{
			name: "Map binds the element to $$this",
			got:  expr.Map(expr.Field("prices"), expr.Multiply("$$this", 1.1)),
			want: bson.D{{Key: "$map", Value: bson.D{
				{Key: "input", Value: "$prices"},
				{Key: "in", Value: bson.D{{Key: "$multiply", Value: bson.A{"$$this", 1.1}}}},
			}}},
		},
		{
			name: "Map with a named variable keeps in last",
			got:  expr.Map(expr.Field("prices"), expr.Multiply("$$price", 1.1), expr.MapAs("price")),
			want: bson.D{{Key: "$map", Value: bson.D{
				{Key: "input", Value: "$prices"},
				{Key: "as", Value: "price"},
				{Key: "in", Value: bson.D{{Key: "$multiply", Value: bson.A{"$$price", 1.1}}}},
			}}},
		},
		{
			name: "First takes a bare expression",
			got:  expr.First(expr.Field("scores")),
			want: bson.D{{Key: "$first", Value: "$scores"}},
		},
		{
			name: "Last takes a bare expression",
			got:  expr.Last(expr.Field("scores")),
			want: bson.D{{Key: "$last", Value: "$scores"}},
		},
		{
			name: "FirstN",
			got:  expr.FirstN(expr.Field("scores"), 3),
			want: bson.D{{Key: "$firstN", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "n", Value: 3},
			}}},
		},
		{
			name: "LastN",
			got:  expr.LastN(expr.Field("scores"), 3),
			want: bson.D{{Key: "$lastN", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "n", Value: 3},
			}}},
		},
		{
			name: "MaxN",
			got:  expr.MaxN(expr.Field("scores"), 3),
			want: bson.D{{Key: "$maxN", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "n", Value: 3},
			}}},
		},
		{
			name: "MinN",
			got:  expr.MinN(expr.Field("scores"), 3),
			want: bson.D{{Key: "$minN", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "n", Value: 3},
			}}},
		},
		{
			name: "In takes the value first",
			got:  expr.In("draft", expr.Field("tags")),
			want: bson.D{{Key: "$in", Value: bson.A{"draft", "$tags"}}},
		},
		{
			name: "IndexOfArray over the whole array",
			got:  expr.IndexOfArray(expr.Field("tags"), "draft"),
			want: bson.D{{Key: "$indexOfArray", Value: bson.A{"$tags", "draft"}}},
		},
		{
			name: "IndexOfArray between a start and an end",
			got:  expr.IndexOfArray(expr.Field("tags"), "draft", 1, 5),
			want: bson.D{{Key: "$indexOfArray", Value: bson.A{"$tags", "draft", 1, 5}}},
		},
		{
			name: "IsArray wraps its argument",
			got:  expr.IsArray(expr.Field("tags")),
			want: bson.D{{Key: "$isArray", Value: bson.A{"$tags"}}},
		},
		{
			name: "Range without a step",
			got:  expr.Range(0, 10),
			want: bson.D{{Key: "$range", Value: bson.A{0, 10}}},
		},
		{
			name: "Range with a step",
			got:  expr.Range(0, 10, 2),
			want: bson.D{{Key: "$range", Value: bson.A{0, 10, 2}}},
		},
		{
			name: "Reduce binds $$value and $$this",
			got:  expr.Reduce(expr.Field("scores"), 0, expr.Add("$$value", "$$this")),
			want: bson.D{{Key: "$reduce", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "initialValue", Value: 0},
				{Key: "in", Value: bson.D{{Key: "$add", Value: bson.A{"$$value", "$$this"}}}},
			}}},
		},
		{
			name: "ReverseArray takes a bare expression",
			got:  expr.ReverseArray(expr.Field("history")),
			want: bson.D{{Key: "$reverseArray", Value: "$history"}},
		},
		{
			name: "Size takes a bare expression",
			got:  expr.Size(expr.Field("tags")),
			want: bson.D{{Key: "$size", Value: "$tags"}},
		},
		{
			name: "Slice with a count",
			got:  expr.Slice(expr.Field("history"), 5),
			want: bson.D{{Key: "$slice", Value: bson.A{"$history", 5}}},
		},
		{
			name: "Slice with a negative count takes from the end",
			got:  expr.Slice(expr.Field("history"), -3),
			want: bson.D{{Key: "$slice", Value: bson.A{"$history", -3}}},
		},
		{
			name: "SliceFrom with a position",
			got:  expr.SliceFrom(expr.Field("history"), 10, 5),
			want: bson.D{{Key: "$slice", Value: bson.A{"$history", 10, 5}}},
		},
		{
			name: "SortArray with a sort document",
			got:  expr.SortArray(expr.Field("items"), bson.D{{Key: "price", Value: -1}}),
			want: bson.D{{Key: "$sortArray", Value: bson.D{
				{Key: "input", Value: "$items"},
				{Key: "sortBy", Value: bson.D{{Key: "price", Value: -1}}},
			}}},
		},
		{
			name: "SortArray with a direction for scalars",
			got:  expr.SortArray(expr.Field("scores"), 1),
			want: bson.D{{Key: "$sortArray", Value: bson.D{
				{Key: "input", Value: "$scores"},
				{Key: "sortBy", Value: 1},
			}}},
		},
		{
			name: "Zip with no options",
			got:  expr.Zip([]any{expr.Field("names"), expr.Field("scores")}),
			want: bson.D{{Key: "$zip", Value: bson.D{
				{Key: "inputs", Value: bson.A{"$names", "$scores"}},
			}}},
		},
		{
			name: "Zip to the longest input with defaults",
			got: expr.Zip([]any{expr.Field("names"), expr.Field("scores")},
				expr.ZipUseLongestLength(), expr.ZipDefaults(bson.A{"", 0})),
			want: bson.D{{Key: "$zip", Value: bson.D{
				{Key: "inputs", Value: bson.A{"$names", "$scores"}},
				{Key: "useLongestLength", Value: true},
				{Key: "defaults", Value: bson.A{"", 0}},
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

func ExampleArrayElemAt() {
	e := expr.ArrayElemAt(expr.Field("scores"), 0)

	printExpr(e)
	// Output: {"$arrayElemAt":["$scores",0]}
}

func ExampleArrayToObject() {
	e := expr.ArrayToObject(expr.Field("settings"))

	printExpr(e)
	// Output: {"$arrayToObject":"$settings"}
}

func ExampleObjectToArray() {
	e := expr.ObjectToArray(expr.Field("settings"))

	printExpr(e)
	// Output: {"$objectToArray":"$settings"}
}

func ExampleConcatArrays() {
	e := expr.ConcatArrays(expr.Field("tags"), expr.Field("extra_tags"))

	printExpr(e)
	// Output: {"$concatArrays":["$tags","$extra_tags"]}
}

func ExampleFilter() {
	e := expr.Filter(expr.Field("scores"), expr.Gte("$$this", 80))

	printExpr(e)
	// Output: {"$filter":{"input":"$scores","cond":{"$gte":["$$this",80]}}}
}

func ExampleFilterAs() {
	e := expr.Filter(expr.Field("scores"), expr.Gte("$$score", 80), expr.FilterAs("score"))

	printExpr(e)
	// Output: {"$filter":{"input":"$scores","cond":{"$gte":["$$score",80]},"as":"score"}}
}

func ExampleFilterLimit() {
	e := expr.Filter(expr.Field("scores"), expr.Gte("$$this", 80), expr.FilterLimit(3))

	printExpr(e)
	// Output: {"$filter":{"input":"$scores","cond":{"$gte":["$$this",80]},"limit":3}}
}

func ExampleMap() {
	e := expr.Map(expr.Field("prices"), expr.Multiply("$$this", 1.1))

	printExpr(e)
	// Output: {"$map":{"input":"$prices","in":{"$multiply":["$$this",1.1]}}}
}

func ExampleMapAs() {
	e := expr.Map(expr.Field("prices"), expr.Multiply("$$price", 1.1), expr.MapAs("price"))

	printExpr(e)
	// Output: {"$map":{"input":"$prices","as":"price","in":{"$multiply":["$$price",1.1]}}}
}

func ExampleFirst() {
	e := expr.First(expr.Field("scores"))

	printExpr(e)
	// Output: {"$first":"$scores"}
}

func ExampleLast() {
	e := expr.Last(expr.Field("scores"))

	printExpr(e)
	// Output: {"$last":"$scores"}
}

func ExampleFirstN() {
	e := expr.FirstN(expr.Field("scores"), 3)

	printExpr(e)
	// Output: {"$firstN":{"input":"$scores","n":3}}
}

func ExampleLastN() {
	e := expr.LastN(expr.Field("scores"), 3)

	printExpr(e)
	// Output: {"$lastN":{"input":"$scores","n":3}}
}

func ExampleMaxN() {
	e := expr.MaxN(expr.Field("scores"), 3)

	printExpr(e)
	// Output: {"$maxN":{"input":"$scores","n":3}}
}

func ExampleMinN() {
	e := expr.MinN(expr.Field("scores"), 3)

	printExpr(e)
	// Output: {"$minN":{"input":"$scores","n":3}}
}

func ExampleIn() {
	e := expr.In("draft", expr.Field("tags"))

	printExpr(e)
	// Output: {"$in":["draft","$tags"]}
}

func ExampleIndexOfArray() {
	e := expr.IndexOfArray(expr.Field("tags"), "draft")

	printExpr(e)
	// Output: {"$indexOfArray":["$tags","draft"]}
}

func ExampleIsArray() {
	e := expr.IsArray(expr.Field("tags"))

	printExpr(e)
	// Output: {"$isArray":["$tags"]}
}

func ExampleRange() {
	e := expr.Range(0, 10, 2)

	printExpr(e)
	// Output: {"$range":[0,10,2]}
}

func ExampleReduce() {
	e := expr.Reduce(expr.Field("scores"), 0, expr.Add("$$value", "$$this"))

	printExpr(e)
	// Output: {"$reduce":{"input":"$scores","initialValue":0,"in":{"$add":["$$value","$$this"]}}}
}

func ExampleReverseArray() {
	e := expr.ReverseArray(expr.Field("history"))

	printExpr(e)
	// Output: {"$reverseArray":"$history"}
}

func ExampleSize() {
	e := expr.Size(expr.Field("tags"))

	printExpr(e)
	// Output: {"$size":"$tags"}
}

func ExampleSlice() {
	e := expr.Slice(expr.Field("history"), 5)

	printExpr(e)
	// Output: {"$slice":["$history",5]}
}

func ExampleSliceFrom() {
	e := expr.SliceFrom(expr.Field("history"), 10, 5)

	printExpr(e)
	// Output: {"$slice":["$history",10,5]}
}

func ExampleSortArray() {
	e := expr.SortArray(expr.Field("items"), bson.D{{Key: "price", Value: -1}})

	printExpr(e)
	// Output: {"$sortArray":{"input":"$items","sortBy":{"price":-1}}}
}

func ExampleZip() {
	e := expr.Zip([]any{expr.Field("names"), expr.Field("scores")})

	printExpr(e)
	// Output: {"$zip":{"inputs":["$names","$scores"]}}
}

func ExampleZipUseLongestLength() {
	e := expr.Zip([]any{expr.Field("names"), expr.Field("scores")}, expr.ZipUseLongestLength())

	printExpr(e)
	// Output: {"$zip":{"inputs":["$names","$scores"],"useLongestLength":true}}
}

func ExampleZipDefaults() {
	e := expr.Zip([]any{expr.Field("names"), expr.Field("scores")},
		expr.ZipUseLongestLength(), expr.ZipDefaults(bson.A{"", 0}))

	printExpr(e)
	// Output: {"$zip":{"inputs":["$names","$scores"],"useLongestLength":true,"defaults":["",0]}}
}
