package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/stage"
)

func TestFilterStages(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Match carries a monq filter",
			got:  stage.Match(monq.Eq("status", "active")),
			want: bson.D{{Key: "$match", Value: bson.D{
				{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}},
			}}},
		},
		{
			name: "Match carries a composed filter",
			got:  stage.Match(monq.And(monq.Eq("status", "active"), monq.Gte("age", 18))),
			want: bson.D{{Key: "$match", Value: bson.D{{Key: "$and", Value: bson.A{
				bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}},
				bson.D{{Key: "age", Value: bson.D{{Key: "$gte", Value: 18}}}},
			}}}}},
		},
		{
			name: "Match with an empty filter",
			got:  stage.Match(bson.D{}),
			want: bson.D{{Key: "$match", Value: bson.D{}}},
		},
		{
			name: "Limit",
			got:  stage.Limit(20),
			want: bson.D{{Key: "$limit", Value: int64(20)}},
		},
		{
			name: "Skip",
			got:  stage.Skip(40),
			want: bson.D{{Key: "$skip", Value: int64(40)}},
		},
		{
			name: "Sample",
			got:  stage.Sample(10),
			want: bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: int64(10)}}}},
		},
		{
			name: "Count takes an output field name",
			got:  stage.Count("active_users"),
			want: bson.D{{Key: "$count", Value: "active_users"}},
		},
		{
			name: "Sort keeps the order of its keys",
			got:  stage.Sort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: 1}}),
			want: bson.D{{Key: "$sort", Value: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: 1},
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

func ExampleMatch() {
	s := stage.Match(monq.Eq("status", "active"))

	printStage(s)
	// Output: {"$match":{"status":{"$eq":"active"}}}
}

func ExampleLimit() {
	s := stage.Limit(20)

	printStage(s)
	// Output: {"$limit":20}
}

func ExampleSkip() {
	s := stage.Skip(40)

	printStage(s)
	// Output: {"$skip":40}
}

func ExampleSample() {
	s := stage.Sample(10)

	printStage(s)
	// Output: {"$sample":{"size":10}}
}

func ExampleCount() {
	s := stage.Count("active_users")

	printStage(s)
	// Output: {"$count":"active_users"}
}

func ExampleSort() {
	s := stage.Sort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: 1}})

	printStage(s)
	// Output: {"$sort":{"created_at":-1,"_id":1}}
}
