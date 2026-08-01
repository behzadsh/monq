package monq_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestRaw(t *testing.T) {
	tests := []struct {
		name string
		in   bson.D
		want bson.D
	}{
		{
			name: "passes a simple document through unchanged",
			in:   bson.D{{Key: "status", Value: "active"}},
			want: bson.D{{Key: "status", Value: "active"}},
		},
		{
			name: "passes an empty document through unchanged",
			in:   bson.D{},
			want: bson.D{},
		},
		{
			name: "passes a hand-written operator expression through unchanged",
			in:   bson.D{{Key: "age", Value: bson.D{{Key: "$gt", Value: 18}}}},
			want: bson.D{{Key: "age", Value: bson.D{{Key: "$gt", Value: 18}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Raw(tt.in)

			if len(got) != len(tt.want) {
				t.Fatalf("Raw() = %v, want %v", got, tt.want)
			}

			for i := range got {
				if got[i].Key != tt.want[i].Key {
					t.Errorf("Raw()[%d].Key = %q, want %q", i, got[i].Key, tt.want[i].Key)
				}
			}
		})
	}
}

func ExampleRaw() {
	filter := monq.Raw(bson.D{{Key: "$where", Value: "this.credits == this.debits"}})

	fmt.Println(filter)
	// Output: {"$where":"this.credits == this.debits"}
}
