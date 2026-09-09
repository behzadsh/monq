package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestUpdateFieldOperators(t *testing.T) {
	tests := []struct {
		name  string
		got   bson.D
		op    string
		field monq.FieldPath
		value any
	}{
		{
			name:  "Set writes a value",
			got:   monq.Set("status", "active"),
			op:    "$set",
			field: "status",
			value: "active",
		},
		{
			name:  "Set reaches a dotted path",
			got:   monq.Set("profile.display_name", "ada"),
			op:    "$set",
			field: "profile.display_name",
			value: "ada",
		},
		{
			name:  "SetOnInsert writes a value",
			got:   monq.SetOnInsert("created_at", "2026-08-01"),
			op:    "$setOnInsert",
			field: "created_at",
			value: "2026-08-01",
		},
		{
			name:  "Unset uses an empty operand",
			got:   monq.Unset("deleted_at"),
			op:    "$unset",
			field: "deleted_at",
			value: "",
		},
		{
			name:  "Inc adds",
			got:   monq.Inc("logins", 1),
			op:    "$inc",
			field: "logins",
			value: 1,
		},
		{
			name:  "Inc subtracts with a negative amount",
			got:   monq.Inc("credits", -5),
			op:    "$inc",
			field: "credits",
			value: -5,
		},
		{
			name:  "Mul multiplies",
			got:   monq.Mul("price", 1.1),
			op:    "$mul",
			field: "price",
			value: 1.1,
		},
		{
			name:  "Min lowers",
			got:   monq.Min("best_time", 42),
			op:    "$min",
			field: "best_time",
			value: 42,
		},
		{
			name:  "Max raises",
			got:   monq.Max("high_score", 250),
			op:    "$max",
			field: "high_score",
			value: 250,
		},
		{
			name:  "Rename carries the new path as a string",
			got:   monq.Rename("nickname", "profile.display_name"),
			op:    "$rename",
			field: "nickname",
			value: "profile.display_name",
		},
		{
			name:  "CurrentDate stores a date",
			got:   monq.CurrentDate("updated_at"),
			op:    "$currentDate",
			field: "updated_at",
			value: true,
		},
		{
			name:  "CurrentDateTimestamp stores a timestamp",
			got:   monq.CurrentDateTimestamp("synced_at"),
			op:    "$currentDate",
			field: "synced_at",
			value: bson.D{{Key: "$type", Value: "timestamp"}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
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
			},
		)
	}
}

func ExampleSet() {
	update := monq.Set("status", "active")

	printFilter(update)
	// Output: {"$set":{"status":"active"}}
}

func ExampleSetOnInsert() {
	update := monq.SetOnInsert("created_at", "2026-08-01")

	printFilter(update)
	// Output: {"$setOnInsert":{"created_at":"2026-08-01"}}
}

func ExampleUnset() {
	update := monq.Unset("deleted_at")

	printFilter(update)
	// Output: {"$unset":{"deleted_at":""}}
}

func ExampleInc() {
	update := monq.Inc("logins", 1)

	printFilter(update)
	// Output: {"$inc":{"logins":1}}
}

func ExampleMul() {
	update := monq.Mul("price", 1.1)

	printFilter(update)
	// Output: {"$mul":{"price":1.1}}
}

func ExampleMin() {
	update := monq.Min("best_time", 42)

	printFilter(update)
	// Output: {"$min":{"best_time":42}}
}

func ExampleMax() {
	update := monq.Max("high_score", 250)

	printFilter(update)
	// Output: {"$max":{"high_score":250}}
}

func ExampleRename() {
	update := monq.Rename("nickname", "profile.display_name")

	printFilter(update)
	// Output: {"$rename":{"nickname":"profile.display_name"}}
}

func ExampleCurrentDate() {
	update := monq.CurrentDate("updated_at")

	printFilter(update)
	// Output: {"$currentDate":{"updated_at":true}}
}

func ExampleCurrentDateTimestamp() {
	update := monq.CurrentDateTimestamp("synced_at")

	printFilter(update)
	// Output: {"$currentDate":{"synced_at":{"$type":"timestamp"}}}
}
