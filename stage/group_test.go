package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/stage"
)

func sumOf(field string) bson.D {
	return bson.D{{Key: "$sum", Value: field}}
}

func TestGroupingStages(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Accumulator builds one output field",
			got:  stage.Accumulator("total", sumOf("$amount")),
			want: bson.D{{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}}},
		},
		{
			name: "Group with a field reference and one accumulator",
			got:  stage.Group("$category", stage.Accumulator("total", sumOf("$amount"))),
			want: bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$category"},
				{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
			}}},
		},
		{
			name: "Group over everything with a nil key",
			got:  stage.Group(nil, stage.Accumulator("total", sumOf("$amount"))),
			want: bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: nil},
				{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
			}}},
		},
		{
			name: "Group without accumulators keeps only the key",
			got:  stage.Group("$category"),
			want: bson.D{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$category"}}}},
		},
		{
			name: "Group keeps the last of two accumulators sharing a name",
			got: stage.Group("$category",
				stage.Accumulator("total", sumOf("$amount")),
				stage.Accumulator("total", sumOf("$net")),
			),
			want: bson.D{{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "$category"},
				{Key: "total", Value: bson.D{{Key: "$sum", Value: "$net"}}},
			}}},
		},
		{
			name: "Bucket with a default bucket",
			got:  stage.Bucket("$price", []any{0, 50, 100}, stage.BucketDefault("other")),
			want: bson.D{{Key: "$bucket", Value: bson.D{
				{Key: "groupBy", Value: "$price"},
				{Key: "boundaries", Value: bson.A{0, 50, 100}},
				{Key: "default", Value: "other"},
			}}},
		},
		{
			name: "Bucket with output fields",
			got: stage.Bucket("$price", []any{0, 100},
				stage.BucketOutput(stage.Accumulator("total", sumOf("$amount"))),
			),
			want: bson.D{{Key: "$bucket", Value: bson.D{
				{Key: "groupBy", Value: "$price"},
				{Key: "boundaries", Value: bson.A{0, 100}},
				{Key: "output", Value: bson.D{
					{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}},
				}},
			}}},
		},
		{
			name: "BucketAuto with a granularity",
			got:  stage.BucketAuto("$price", 4, stage.BucketGranularity("R20")),
			want: bson.D{{Key: "$bucketAuto", Value: bson.D{
				{Key: "groupBy", Value: "$price"},
				{Key: "buckets", Value: 4},
				{Key: "granularity", Value: "R20"},
			}}},
		},
		{
			name: "SortByCount",
			got:  stage.SortByCount("$category"),
			want: bson.D{{Key: "$sortByCount", Value: "$category"}},
		},
		{
			name: "FacetPipeline holds its stages as an array",
			got:  stage.FacetPipeline("total", stage.Count("n")),
			want: bson.D{{Key: "total", Value: bson.A{
				bson.D{{Key: "$count", Value: "n"}},
			}}},
		},
		{
			name: "Facet merges its sub-pipelines",
			got: stage.Facet(
				stage.FacetPipeline("newest", stage.Limit(5)),
				stage.FacetPipeline("total", stage.Count("n")),
			),
			want: bson.D{{Key: "$facet", Value: bson.D{
				{Key: "newest", Value: bson.A{bson.D{{Key: "$limit", Value: int64(5)}}}},
				{Key: "total", Value: bson.A{bson.D{{Key: "$count", Value: "n"}}}},
			}}},
		},
		{
			name: "Facet keeps the last of two sub-pipelines sharing a name",
			got: stage.Facet(
				stage.FacetPipeline("summary", stage.Limit(5)),
				stage.FacetPipeline("summary", stage.Count("n")),
			),
			want: bson.D{{Key: "$facet", Value: bson.D{
				{Key: "summary", Value: bson.A{bson.D{{Key: "$count", Value: "n"}}}},
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

func ExampleAccumulator() {
	field := stage.Accumulator("total", bson.D{{Key: "$sum", Value: "$amount"}})

	printStage(field)
	// Output: {"total":{"$sum":"$amount"}}
}

func ExampleGroup() {
	s := stage.Group("$category", stage.Accumulator("total", bson.D{{Key: "$sum", Value: "$amount"}}))

	printStage(s)
	// Output: {"$group":{"_id":"$category","total":{"$sum":"$amount"}}}
}

func ExampleBucket() {
	s := stage.Bucket("$price", []any{0, 50, 100}, stage.BucketDefault("other"))

	printStage(s)
	// Output: {"$bucket":{"groupBy":"$price","boundaries":[0,50,100],"default":"other"}}
}

func ExampleBucketAuto() {
	s := stage.BucketAuto("$price", 4, stage.BucketGranularity("R20"))

	printStage(s)
	// Output: {"$bucketAuto":{"groupBy":"$price","buckets":4,"granularity":"R20"}}
}

func ExampleSortByCount() {
	s := stage.SortByCount("$category")

	printStage(s)
	// Output: {"$sortByCount":"$category"}
}

func ExampleFacetPipeline() {
	facet := stage.FacetPipeline("total", stage.Count("n"))

	printStage(facet)
	// Output: {"total":[{"$count":"n"}]}
}

func ExampleFacet() {
	s := stage.Facet(
		stage.FacetPipeline("newest", stage.Sort(bson.D{{Key: "created_at", Value: -1}}), stage.Limit(5)),
		stage.FacetPipeline("total", stage.Count("n")),
	)

	printStage(s)
	// Output: {"$facet":{"newest":[{"$sort":{"created_at":-1}},{"$limit":5}],"total":[{"$count":"n"}]}}
}
