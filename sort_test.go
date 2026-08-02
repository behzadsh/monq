package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestSortEntries(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Asc",
			got:  monq.Asc("created_at"),
			want: bson.D{{Key: "created_at", Value: 1}},
		},
		{
			name: "Desc",
			got:  monq.Desc("created_at"),
			want: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			name: "TextScore",
			got:  monq.TextScore("score"),
			want: bson.D{{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}}},
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

func TestSort(t *testing.T) {
	tests := []struct {
		name    string
		entries []bson.D
		want    bson.D
	}{
		{
			name:    "no entries",
			entries: nil,
			want:    bson.D{},
		},
		{
			name:    "one entry",
			entries: []bson.D{monq.Desc("created_at")},
			want:    bson.D{{Key: "created_at", Value: -1}},
		},
		{
			name:    "keeps the order it is given",
			entries: []bson.D{monq.Desc("created_at"), monq.Asc("_id")},
			want: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: 1},
			},
		},
		{
			name:    "a field named twice keeps its position and the later direction",
			entries: []bson.D{monq.Asc("created_at"), monq.Asc("_id"), monq.Desc("created_at")},
			want: bson.D{
				{Key: "created_at", Value: -1},
				{Key: "_id", Value: 1},
			},
		},
		{
			name:    "mixes a text score with a field",
			entries: []bson.D{monq.TextScore("score"), monq.Asc("_id")},
			want: bson.D{
				{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}},
				{Key: "_id", Value: 1},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Sort(tt.entries...)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Sort() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSortLeavesEntriesUntouched(t *testing.T) {
	createdAt := monq.Desc("created_at")
	id := monq.Asc("_id")

	first := monq.Sort(createdAt, id)
	second := monq.Sort(createdAt, id)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Sort() = %v on the second call, want %v", second, first)
	}

	if !reflect.DeepEqual(createdAt, bson.D{{Key: "created_at", Value: -1}}) {
		t.Fatalf("entry = %v, want the caller's document unchanged", createdAt)
	}
}

func ExampleSort() {
	sort := monq.Sort(monq.Desc("created_at"), monq.Asc("_id"))

	printFilter(sort)
	// Output: {"created_at":-1,"_id":1}
}

func ExampleAsc() {
	entry := monq.Asc("created_at")

	printFilter(entry)
	// Output: {"created_at":1}
}

func ExampleDesc() {
	entry := monq.Desc("created_at")

	printFilter(entry)
	// Output: {"created_at":-1}
}

func ExampleTextScore() {
	sort := monq.Sort(monq.TextScore("score"))

	printFilter(sort)
	// Output: {"score":{"$meta":"textScore"}}
}
