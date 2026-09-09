package monq_test

import (
	"fmt"
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestExists(t *testing.T) {
	tests := []struct {
		name   string
		field  monq.FieldPath
		exists bool
		want   bson.D
	}{
		{
			name:   "field must be present",
			field:  "verified_at",
			exists: true,
			want:   bson.D{{Key: "verified_at", Value: bson.D{{Key: "$exists", Value: true}}}},
		},
		{
			name:   "field must be absent",
			field:  "deleted_at",
			exists: false,
			want:   bson.D{{Key: "deleted_at", Value: bson.D{{Key: "$exists", Value: false}}}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := monq.Exists(tt.field, tt.exists)

				inner, ok := got[0].Value.(bson.D)
				if !ok || inner[0].Key != "$exists" || inner[0].Value != tt.exists {
					t.Fatalf("Exists() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func ExampleExists() {
	filter := monq.Exists("verified_at", true)

	fmt.Println(filter)
	// Output: {"verified_at":{"$exists":true}}
}

func TestType(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		types []any
		want  any
	}{
		{
			name:  "single alias is a scalar",
			field: "legacyId",
			types: []any{"string"},
			want:  "string",
		},
		{
			name:  "single numeric code is a scalar",
			field: "legacyId",
			types: []any{2},
			want:  2,
		},
		{
			name:  "several types become an array",
			field: "legacyId",
			types: []any{"string", "objectId"},
			want:  bson.A{"string", "objectId"},
		},
		{
			name:  "mixed alias and code",
			field: "legacyId",
			types: []any{"string", 7},
			want:  bson.A{"string", 7},
		},
		{
			name:  "no types is an empty array",
			field: "legacyId",
			types: nil,
			want:  bson.A{},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := monq.Type(tt.field, tt.types...)

				inner, ok := got[0].Value.(bson.D)
				if !ok || got[0].Key != string(tt.field) || inner[0].Key != "$type" {
					t.Fatalf("Type() = %v, want a $type operator document on %q", got, tt.field)
				}

				if !reflect.DeepEqual(inner[0].Value, tt.want) {
					t.Fatalf("Type() value = %v, want %v", inner[0].Value, tt.want)
				}
			},
		)
	}
}

func ExampleType() {
	filter := monq.Type("legacyId", "string", "objectId")

	fmt.Println(filter)
	// Output: {"legacyId":{"$type":["string","objectId"]}}
}
