package monq_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestRegex(t *testing.T) {
	tests := []struct {
		name    string
		field   monq.FieldPath
		pattern string
		options string
		want    bson.D
	}{
		{
			name:    "case-insensitive prefix match",
			field:   "email",
			pattern: "^alice",
			options: "i",
			want: bson.D{
				{Key: "email", Value: bson.D{{Key: "$regex", Value: "^alice"}, {Key: "$options", Value: "i"}}},
			},
		},
		{
			name:    "no flags",
			field:   "name",
			pattern: "^Bob$",
			options: "",
			want: bson.D{
				{Key: "name", Value: bson.D{{Key: "$regex", Value: "^Bob$"}, {Key: "$options", Value: ""}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Regex(tt.field, tt.pattern, tt.options)

			inner, ok := got[0].Value.(bson.D)
			if !ok || inner[0].Key != "$regex" || inner[0].Value != tt.pattern {
				t.Fatalf("Regex() = %v, want %v", got, tt.want)
			}

			if inner[1].Key != "$options" || inner[1].Value != tt.options {
				t.Fatalf("Regex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleRegex() {
	filter := monq.Regex("email", "^alice", "i")

	fmt.Println(filter)
	// Output: {"email":{"$regex":"^alice","$options":"i"}}
}
