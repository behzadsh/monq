package monq_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestAnd(t *testing.T) {
	tests := []struct {
		name    string
		filters []bson.D
		want    int
	}{
		{
			name: "two filters",
			filters: []bson.D{
				monq.Eq("status", "active"),
				monq.Gte("age", 18),
			},
			want: 2,
		},
		{
			name:    "no filters",
			filters: nil,
			want:    0,
		},
		{
			name: "composes with Raw",
			filters: []bson.D{
				monq.Eq("status", "active"),
				monq.Raw(bson.D{{Key: "legacyField", Value: bson.D{{Key: "$type", Value: "string"}}}}),
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.And(tt.filters...)

			inner, ok := got[0].Value.(bson.A)
			if !ok || got[0].Key != "$and" || len(inner) != tt.want {
				t.Fatalf("And() = %v, want %d filters under $and", got, tt.want)
			}
		})
	}
}

func ExampleAnd() {
	filter := monq.And(monq.Eq("status", "active"), monq.Gte("age", 18))

	fmt.Println(filter)
	// Output: {"$and":[{"status":{"$eq":"active"}},{"age":{"$gte":{"$numberInt":"18"}}}]}
}

func TestOr(t *testing.T) {
	tests := []struct {
		name    string
		filters []bson.D
		want    int
	}{
		{
			name: "two filters",
			filters: []bson.D{
				monq.Gte("stats.followers", 10000),
				monq.Exists("verified_at", true),
			},
			want: 2,
		},
		{
			name:    "no filters",
			filters: nil,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Or(tt.filters...)

			inner, ok := got[0].Value.(bson.A)
			if !ok || got[0].Key != "$or" || len(inner) != tt.want {
				t.Fatalf("Or() = %v, want %d filters under $or", got, tt.want)
			}
		})
	}
}

func ExampleOr() {
	filter := monq.Or(monq.Gte("stats.followers", 10000), monq.Exists("verified_at", true))

	fmt.Println(filter)
	// Output: {"$or":[{"stats.followers":{"$gte":{"$numberInt":"10000"}}},{"verified_at":{"$exists":true}}]}
}

func TestNor(t *testing.T) {
	tests := []struct {
		name    string
		filters []bson.D
		want    int
	}{
		{
			name: "two filters",
			filters: []bson.D{
				monq.Eq("status", "banned"),
				monq.Eq("status", "suspended"),
			},
			want: 2,
		},
		{
			name:    "no filters",
			filters: nil,
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Nor(tt.filters...)

			inner, ok := got[0].Value.(bson.A)
			if !ok || got[0].Key != "$nor" || len(inner) != tt.want {
				t.Fatalf("Nor() = %v, want %d filters under $nor", got, tt.want)
			}
		})
	}
}

func ExampleNor() {
	filter := monq.Nor(monq.Eq("status", "banned"), monq.Eq("status", "suspended"))

	fmt.Println(filter)
	// Output: {"$nor":[{"status":{"$eq":"banned"}},{"status":{"$eq":"suspended"}}]}
}

func TestNot(t *testing.T) {
	tests := []struct {
		name  string
		expr  bson.D
		field string
		op    string
	}{
		{
			name:  "negates Gt",
			expr:  monq.Gt("age", 5),
			field: "age",
			op:    "$gt",
		},
		{
			name:  "negates In",
			expr:  monq.In("status", "banned", "suspended"),
			field: "status",
			op:    "$in",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Not(tt.expr)

			if got[0].Key != tt.field {
				t.Fatalf("Not() field = %q, want %q", got[0].Key, tt.field)
			}

			outer, ok := got[0].Value.(bson.D)
			if !ok || outer[0].Key != "$not" {
				t.Fatalf("Not() = %v, want a $not operator document", got)
			}

			innerOp, ok := outer[0].Value.(bson.D)
			if !ok || innerOp[0].Key != tt.op {
				t.Fatalf("Not() inner operator = %v, want %q", outer[0].Value, tt.op)
			}
		})
	}
}

func ExampleNot() {
	filter := monq.Not(monq.Gt("age", 5))

	printFilter(filter)
	// Output: {"age":{"$not":{"$gt":5}}}
}
