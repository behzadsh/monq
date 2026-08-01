package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestUpdate(t *testing.T) {
	tests := []struct {
		name string
		ops  []bson.D
		want bson.D
	}{
		{
			name: "no operators",
			ops:  nil,
			want: bson.D{},
		},
		{
			name: "one operator passes through",
			ops:  []bson.D{monq.Set("status", "active")},
			want: bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}},
		},
		{
			name: "same operator merges its fields",
			ops:  []bson.D{monq.Set("status", "active"), monq.Set("name", "ada")},
			want: bson.D{{Key: "$set", Value: bson.D{
				{Key: "status", Value: "active"},
				{Key: "name", Value: "ada"},
			}}},
		},
		{
			name: "different operators keep first-seen order",
			ops: []bson.D{
				monq.Set("status", "active"),
				monq.Inc("logins", 1),
				monq.Set("name", "ada"),
				monq.Unset("deleted_at"),
			},
			want: bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "status", Value: "active"},
					{Key: "name", Value: "ada"},
				}},
				{Key: "$inc", Value: bson.D{{Key: "logins", Value: 1}}},
				{Key: "$unset", Value: bson.D{{Key: "deleted_at", Value: ""}}},
			},
		},
		{
			name: "same field twice under one operator keeps the last value",
			ops:  []bson.D{monq.Set("status", "active"), monq.Set("status", "banned")},
			want: bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "banned"}}}},
		},
		{
			name: "merges an already merged update",
			ops: []bson.D{
				monq.Update(monq.Set("status", "active"), monq.Inc("logins", 1)),
				monq.Set("name", "ada"),
			},
			want: bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "status", Value: "active"},
					{Key: "name", Value: "ada"},
				}},
				{Key: "$inc", Value: bson.D{{Key: "logins", Value: 1}}},
			},
		},
		{
			name: "a hand-written operator composes like the rest",
			ops: []bson.D{
				monq.Set("status", "active"),
				monq.Raw(bson.D{{Key: "$set", Value: bson.D{{Key: "legacy", Value: 1}}}}),
			},
			want: bson.D{{Key: "$set", Value: bson.D{
				{Key: "status", Value: "active"},
				{Key: "legacy", Value: 1},
			}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Update(tt.ops...)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Update() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUpdateLeavesOperandsUntouched(t *testing.T) {
	status := monq.Set("status", "active")
	name := monq.Set("name", "ada")
	logins := monq.Inc("logins", 1)

	first := monq.Update(status, name, logins)
	second := monq.Update(status, name, logins)

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Update() = %v on the second call, want %v", second, first)
	}

	want := []bson.D{
		{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}},
		{{Key: "$set", Value: bson.D{{Key: "name", Value: "ada"}}}},
		{{Key: "$inc", Value: bson.D{{Key: "logins", Value: 1}}}},
	}

	for i, op := range []bson.D{status, name, logins} {
		if !reflect.DeepEqual(op, want[i]) {
			t.Fatalf("operand %d = %v, want %v unchanged", i, op, want[i])
		}
	}
}

func ExampleUpdate() {
	update := monq.Update(
		monq.Set("status", "active"),
		monq.Inc("logins", 1),
		monq.Set("name", "ada"),
	)

	printFilter(update)
	// Output: {"$set":{"status":"active","name":"ada"},"$inc":{"logins":1}}
}
