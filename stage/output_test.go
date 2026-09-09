package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/stage"
)

func TestOutputStages(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Namespace holds a database and collection",
			got:  stage.Namespace("reports", "daily_totals"),
			want: bson.D{{Key: "db", Value: "reports"}, {Key: "coll", Value: "daily_totals"}},
		},
		{
			name: "Out to a collection in the same database",
			got:  stage.Out("daily_totals"),
			want: bson.D{{Key: "$out", Value: "daily_totals"}},
		},
		{
			name: "Out to another database",
			got:  stage.Out(stage.Namespace("reports", "daily_totals")),
			want: bson.D{
				{
					Key: "$out", Value: bson.D{
						{Key: "db", Value: "reports"},
						{Key: "coll", Value: "daily_totals"},
					},
				},
			},
		},
		{
			name: "Documents",
			got:  stage.Documents(bson.D{{Key: "x", Value: 1}}, bson.D{{Key: "x", Value: 2}}),
			want: bson.D{
				{
					Key: "$documents", Value: bson.A{
						bson.D{{Key: "x", Value: 1}},
						bson.D{{Key: "x", Value: 2}},
					},
				},
			},
		},
		{
			name: "Documents with nothing",
			got:  stage.Documents(),
			want: bson.D{{Key: "$documents", Value: bson.A{}}},
		},
		{
			name: "Merge into a collection with no options",
			got:  stage.Merge("daily_totals"),
			want: bson.D{{Key: "$merge", Value: bson.D{{Key: "into", Value: "daily_totals"}}}},
		},
		{
			name: "Merge on one field",
			got:  stage.Merge("daily_totals", stage.MergeOn("date")),
			want: bson.D{
				{
					Key: "$merge", Value: bson.D{
						{Key: "into", Value: "daily_totals"},
						{Key: "on", Value: "date"},
					},
				},
			},
		},
		{
			name: "Merge on several fields",
			got:  stage.Merge("daily_totals", stage.MergeOn("date", "region")),
			want: bson.D{
				{
					Key: "$merge", Value: bson.D{
						{Key: "into", Value: "daily_totals"},
						{Key: "on", Value: bson.A{"date", "region"}},
					},
				},
			},
		},
		{
			name: "Merge with a named action when a document matches",
			got: stage.Merge(
				"daily_totals",
				stage.MergeWhenMatched("replace"),
				stage.MergeWhenNotMatched("insert"),
			),
			want: bson.D{
				{
					Key: "$merge", Value: bson.D{
						{Key: "into", Value: "daily_totals"},
						{Key: "whenMatched", Value: "replace"},
						{Key: "whenNotMatched", Value: "insert"},
					},
				},
			},
		},
		{
			name: "Merge with a pipeline when a document matches",
			got: stage.Merge(
				"daily_totals",
				stage.MergeLet(bson.D{{Key: "amount", Value: "$total"}}),
				stage.MergeWhenMatched(
					[]bson.D{
						stage.Set(stage.Field("total", "$$amount")),
					},
				),
			),
			want: bson.D{
				{
					Key: "$merge", Value: bson.D{
						{Key: "into", Value: "daily_totals"},
						{Key: "let", Value: bson.D{{Key: "amount", Value: "$total"}}},
						{
							Key: "whenMatched", Value: []bson.D{
								{{Key: "$set", Value: bson.D{{Key: "total", Value: "$$amount"}}}},
							},
						},
					},
				},
			},
		},
		{
			name: "Merge into another database",
			got:  stage.Merge(stage.Namespace("reports", "daily_totals")),
			want: bson.D{
				{
					Key: "$merge", Value: bson.D{
						{
							Key: "into", Value: bson.D{
								{Key: "db", Value: "reports"},
								{Key: "coll", Value: "daily_totals"},
							},
						},
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

func ExampleNamespace() {
	target := stage.Namespace("reports", "daily_totals")

	printStage(target)
	// Output: {"db":"reports","coll":"daily_totals"}
}

func ExampleOut() {
	s := stage.Out("daily_totals")

	printStage(s)
	// Output: {"$out":"daily_totals"}
}

func ExampleDocuments() {
	s := stage.Documents(bson.D{{Key: "x", Value: 1}}, bson.D{{Key: "x", Value: 2}})

	printStage(s)
	// Output: {"$documents":[{"x":1},{"x":2}]}
}

func ExampleMerge() {
	s := stage.Merge("daily_totals", stage.MergeOn("date"), stage.MergeWhenMatched("replace"))

	printStage(s)
	// Output: {"$merge":{"into":"daily_totals","on":"date","whenMatched":"replace"}}
}

func ExampleMergeOn() {
	s := stage.Merge("daily_totals", stage.MergeOn("date", "region"))

	printStage(s)
	// Output: {"$merge":{"into":"daily_totals","on":["date","region"]}}
}

func ExampleMergeWhenMatched() {
	s := stage.Merge("daily_totals", stage.MergeWhenMatched("replace"))

	printStage(s)
	// Output: {"$merge":{"into":"daily_totals","whenMatched":"replace"}}
}

func ExampleMergeWhenNotMatched() {
	s := stage.Merge("daily_totals", stage.MergeWhenNotMatched("discard"))

	printStage(s)
	// Output: {"$merge":{"into":"daily_totals","whenNotMatched":"discard"}}
}

func ExampleMergeLet() {
	s := stage.Merge(
		"daily_totals",
		stage.MergeLet(bson.D{{Key: "amount", Value: "$total"}}),
		stage.MergeWhenMatched([]bson.D{stage.Set(stage.Field("total", "$$amount"))}),
	)

	printStage(s)
	// Output: {"$merge":{"into":"daily_totals","let":{"amount":"$total"},"whenMatched":[{"$set":{"total":"$$amount"}}]}}
}
