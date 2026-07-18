package monq_test

import (
	"fmt"
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
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Exists(tt.field, tt.exists)

			inner, ok := got[0].Value.(bson.D)
			if !ok || inner[0].Key != "$exists" || inner[0].Value != tt.exists {
				t.Fatalf("Exists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleExists() {
	filter := monq.Exists("verified_at", true)

	fmt.Println(filter)
	// Output: {"verified_at":{"$exists":true}}
}
