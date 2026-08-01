package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/stage"
)

func TestUnwind(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "path only, still in document form",
			got:  stage.Unwind("$items"),
			want: bson.D{{Key: "$unwind", Value: bson.D{{Key: "path", Value: "$items"}}}},
		},
		{
			name: "keeps documents with nothing to unwind",
			got:  stage.Unwind("$items", stage.PreserveNullAndEmptyArrays()),
			want: bson.D{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$items"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
			}}},
		},
		{
			name: "records the element index",
			got:  stage.Unwind("$items", stage.IncludeArrayIndex("idx")),
			want: bson.D{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$items"},
				{Key: "includeArrayIndex", Value: "idx"},
			}}},
		},
		{
			name: "both options, in the order given",
			got:  stage.Unwind("$items", stage.IncludeArrayIndex("idx"), stage.PreserveNullAndEmptyArrays()),
			want: bson.D{{Key: "$unwind", Value: bson.D{
				{Key: "path", Value: "$items"},
				{Key: "includeArrayIndex", Value: "idx"},
				{Key: "preserveNullAndEmptyArrays", Value: true},
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

func ExampleUnwind() {
	s := stage.Unwind("$items", stage.PreserveNullAndEmptyArrays())

	printStage(s)
	// Output: {"$unwind":{"path":"$items","preserveNullAndEmptyArrays":true}}
}

func ExampleIncludeArrayIndex() {
	s := stage.Unwind("$items", stage.IncludeArrayIndex("position"))

	printStage(s)
	// Output: {"$unwind":{"path":"$items","includeArrayIndex":"position"}}
}

func ExamplePreserveNullAndEmptyArrays() {
	s := stage.Unwind("$items", stage.PreserveNullAndEmptyArrays())

	printStage(s)
	// Output: {"$unwind":{"path":"$items","preserveNullAndEmptyArrays":true}}
}
