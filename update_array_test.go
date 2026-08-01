package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestUpdateArrayOperators(t *testing.T) {
	tests := []struct {
		name  string
		got   bson.D
		op    string
		field monq.FieldPath
		value any
	}{
		{
			name:  "Push appends one value",
			got:   monq.Push("tags", "go"),
			op:    "$push",
			field: "tags",
			value: "go",
		},
		{
			name:  "Push keeps a slice as one element",
			got:   monq.Push("tags", []string{"go", "mongodb"}),
			op:    "$push",
			field: "tags",
			value: []string{"go", "mongodb"},
		},
		{
			name:  "PushEach without modifiers",
			got:   monq.PushEach("tags", []any{"go", "mongodb"}),
			op:    "$push",
			field: "tags",
			value: bson.D{{Key: "$each", Value: bson.A{"go", "mongodb"}}},
		},
		{
			name:  "PushEach with every modifier",
			got:   monq.PushEach("scores", []any{90, 80}, monq.PushPosition(0), monq.PushSort(-1), monq.PushSlice(3)),
			op:    "$push",
			field: "scores",
			value: bson.D{
				{Key: "$each", Value: bson.A{90, 80}},
				{Key: "$position", Value: 0},
				{Key: "$sort", Value: -1},
				{Key: "$slice", Value: 3},
			},
		},
		{
			name:  "PushEach sorts documents by a sort document",
			got:   monq.PushEach("items", []any{"a"}, monq.PushSort(bson.D{{Key: "score", Value: -1}})),
			op:    "$push",
			field: "items",
			value: bson.D{
				{Key: "$each", Value: bson.A{"a"}},
				{Key: "$sort", Value: bson.D{{Key: "score", Value: -1}}},
			},
		},
		{
			name:  "PushEach with no values",
			got:   monq.PushEach("tags", nil),
			op:    "$push",
			field: "tags",
			value: bson.D{{Key: "$each", Value: bson.A{}}},
		},
		{
			name:  "AddToSet appends one value",
			got:   monq.AddToSet("tags", "go"),
			op:    "$addToSet",
			field: "tags",
			value: "go",
		},
		{
			name:  "AddToSetEach appends several",
			got:   monq.AddToSetEach("tags", []any{"go", "mongodb"}),
			op:    "$addToSet",
			field: "tags",
			value: bson.D{{Key: "$each", Value: bson.A{"go", "mongodb"}}},
		},
		{
			name:  "Pull takes a monq filter for arrays of documents",
			got:   monq.Pull("items", monq.Eq("sku", "abc")),
			op:    "$pull",
			field: "items",
			value: bson.D{{Key: "sku", Value: bson.D{{Key: "$eq", Value: "abc"}}}},
		},
		{
			name:  "Pull takes a bare operator expression for arrays of scalars",
			got:   monq.Pull("scores", monq.Raw(bson.D{{Key: "$gte", Value: 80}})),
			op:    "$pull",
			field: "scores",
			value: bson.D{{Key: "$gte", Value: 80}},
		},
		{
			name:  "Pull takes a plain value",
			got:   monq.Pull("tags", "draft"),
			op:    "$pull",
			field: "tags",
			value: "draft",
		},
		{
			name:  "PullAll removes known values",
			got:   monq.PullAll("tags", []any{"draft", "wip"}),
			op:    "$pullAll",
			field: "tags",
			value: bson.A{"draft", "wip"},
		},
		{
			name:  "PopFirst removes the leading element",
			got:   monq.PopFirst("queue"),
			op:    "$pop",
			field: "queue",
			value: -1,
		},
		{
			name:  "PopLast removes the trailing element",
			got:   monq.PopLast("history"),
			op:    "$pop",
			field: "history",
			value: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.got) != 1 || tt.got[0].Key != tt.op {
				t.Fatalf("got %v, want a %s update document", tt.got, tt.op)
			}

			fields, ok := tt.got[0].Value.(bson.D)
			if !ok || len(fields) != 1 || fields[0].Key != string(tt.field) {
				t.Fatalf("got %v, want %s to hold %q", tt.got, tt.op, tt.field)
			}

			if !reflect.DeepEqual(fields[0].Value, tt.value) {
				t.Fatalf("value = %v, want %v", fields[0].Value, tt.value)
			}
		})
	}
}

func TestArrayUpdatesMergeThroughUpdate(t *testing.T) {
	got := monq.Update(
		monq.Push("history", "login"),
		monq.AddToSet("tags", "go"),
		monq.Push("audit", "login"),
	)

	want := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "history", Value: "login"},
			{Key: "audit", Value: "login"},
		}},
		{Key: "$addToSet", Value: bson.D{{Key: "tags", Value: "go"}}},
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Update() = %v, want %v", got, want)
	}
}

func ExamplePush() {
	update := monq.Push("tags", "go")

	printFilter(update)
	// Output: {"$push":{"tags":"go"}}
}

func ExamplePushEach() {
	update := monq.PushEach("scores", []any{90, 80}, monq.PushSort(-1), monq.PushSlice(3))

	printFilter(update)
	// Output: {"$push":{"scores":{"$each":[90,80],"$sort":-1,"$slice":3}}}
}

func ExamplePushPosition() {
	update := monq.PushEach("queue", []any{"urgent"}, monq.PushPosition(0))

	printFilter(update)
	// Output: {"$push":{"queue":{"$each":["urgent"],"$position":0}}}
}

func ExamplePushSlice() {
	update := monq.PushEach("recent", []any{"a"}, monq.PushSlice(-5))

	printFilter(update)
	// Output: {"$push":{"recent":{"$each":["a"],"$slice":-5}}}
}

func ExamplePushSort() {
	update := monq.PushEach("items", []any{"a"}, monq.PushSort(bson.D{{Key: "score", Value: -1}}))

	printFilter(update)
	// Output: {"$push":{"items":{"$each":["a"],"$sort":{"score":-1}}}}
}

func ExampleAddToSet() {
	update := monq.AddToSet("tags", "go")

	printFilter(update)
	// Output: {"$addToSet":{"tags":"go"}}
}

func ExampleAddToSetEach() {
	update := monq.AddToSetEach("tags", []any{"go", "mongodb"})

	printFilter(update)
	// Output: {"$addToSet":{"tags":{"$each":["go","mongodb"]}}}
}

func ExamplePull() {
	update := monq.Pull("items", monq.Eq("sku", "abc"))

	printFilter(update)
	// Output: {"$pull":{"items":{"sku":{"$eq":"abc"}}}}
}

func ExamplePullAll() {
	update := monq.PullAll("tags", []any{"draft", "wip"})

	printFilter(update)
	// Output: {"$pullAll":{"tags":["draft","wip"]}}
}

func ExamplePopFirst() {
	update := monq.PopFirst("queue")

	printFilter(update)
	// Output: {"$pop":{"queue":-1}}
}

func ExamplePopLast() {
	update := monq.PopLast("history")

	printFilter(update)
	// Output: {"$pop":{"history":1}}
}
