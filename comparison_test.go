package monq_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestEq(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		value any
		want  bson.D
	}{
		{
			name:  "string value",
			field: "status",
			value: "active",
			want:  bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}},
		},
		{
			name:  "dotted field path",
			field: "stats.followers",
			value: 10000,
			want:  bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$eq", Value: 10000}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Eq(tt.field, tt.value)

			if len(got) != len(tt.want) || got[0].Key != tt.want[0].Key {
				t.Fatalf("Eq() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleEq() {
	filter := monq.Eq("status", "active")

	fmt.Println(filter)
	// Output: {"status":{"$eq":"active"}}
}

func TestNe(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		value any
		want  bson.D
	}{
		{
			name:  "string value",
			field: "status",
			value: "banned",
			want:  bson.D{{Key: "status", Value: bson.D{{Key: "$ne", Value: "banned"}}}},
		},
		{
			name:  "dotted field path",
			field: "profile.role",
			value: "guest",
			want:  bson.D{{Key: "profile.role", Value: bson.D{{Key: "$ne", Value: "guest"}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Ne(tt.field, tt.value)

			if len(got) != len(tt.want) || got[0].Key != tt.want[0].Key {
				t.Fatalf("Ne() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleNe() {
	filter := monq.Ne("status", "banned")

	fmt.Println(filter)
	// Output: {"status":{"$ne":"banned"}}
}

func TestGt(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		value any
		want  bson.D
	}{
		{
			name:  "numeric value",
			field: "age",
			value: 18,
			want:  bson.D{{Key: "age", Value: bson.D{{Key: "$gt", Value: 18}}}},
		},
		{
			name:  "dotted field path",
			field: "stats.followers",
			value: 10000,
			want:  bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$gt", Value: 10000}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Gt(tt.field, tt.value)

			if len(got) != len(tt.want) || got[0].Key != tt.want[0].Key {
				t.Fatalf("Gt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleGt() {
	filter := monq.Gt("age", 18)

	printFilter(filter)
	// Output: {"age":{"$gt":18}}
}

func TestGte(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		value any
		want  bson.D
	}{
		{
			name:  "numeric value",
			field: "age",
			value: 18,
			want:  bson.D{{Key: "age", Value: bson.D{{Key: "$gte", Value: 18}}}},
		},
		{
			name:  "dotted field path",
			field: "stats.followers",
			value: 10000,
			want:  bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$gte", Value: 10000}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Gte(tt.field, tt.value)

			if len(got) != len(tt.want) || got[0].Key != tt.want[0].Key {
				t.Fatalf("Gte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleGte() {
	filter := monq.Gte("stats.followers", 10000)

	printFilter(filter)
	// Output: {"stats.followers":{"$gte":10000}}
}

func TestLt(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		value any
		want  bson.D
	}{
		{
			name:  "numeric value",
			field: "age",
			value: 65,
			want:  bson.D{{Key: "age", Value: bson.D{{Key: "$lt", Value: 65}}}},
		},
		{
			name:  "dotted field path",
			field: "stats.followers",
			value: 100,
			want:  bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$lt", Value: 100}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Lt(tt.field, tt.value)

			if len(got) != len(tt.want) || got[0].Key != tt.want[0].Key {
				t.Fatalf("Lt() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleLt() {
	filter := monq.Lt("age", 65)

	printFilter(filter)
	// Output: {"age":{"$lt":65}}
}

func TestLte(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		value any
		want  bson.D
	}{
		{
			name:  "numeric value",
			field: "age",
			value: 65,
			want:  bson.D{{Key: "age", Value: bson.D{{Key: "$lte", Value: 65}}}},
		},
		{
			name:  "dotted field path",
			field: "stats.followers",
			value: 100,
			want:  bson.D{{Key: "stats.followers", Value: bson.D{{Key: "$lte", Value: 100}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Lte(tt.field, tt.value)

			if len(got) != len(tt.want) || got[0].Key != tt.want[0].Key {
				t.Fatalf("Lte() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleLte() {
	filter := monq.Lte("age", 65)

	printFilter(filter)
	// Output: {"age":{"$lte":65}}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name   string
		field  monq.FieldPath
		values []any
		want   bson.A
	}{
		{
			name:   "multiple spread values",
			field:  "status",
			values: []any{"active", "pending"},
			want:   bson.A{"active", "pending"},
		},
		{
			name:   "single value",
			field:  "status",
			values: []any{"active"},
			want:   bson.A{"active"},
		},
		{
			name:   "no values",
			field:  "status",
			values: nil,
			want:   bson.A{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.In(tt.field, tt.values...)

			inner, ok := got[0].Value.(bson.D)
			if !ok || inner[0].Key != "$in" {
				t.Fatalf("In() = %v, want a $in operator document", got)
			}

			gotArr, ok := inner[0].Value.(bson.A)
			if !ok || len(gotArr) != len(tt.want) {
				t.Fatalf("In() = %v, want values %v", got, tt.want)
			}
		})
	}

	t.Run("un-spread slice becomes one candidate, not one per element", func(t *testing.T) {
		ids := []string{"a", "b", "c"}

		got := monq.In("status", ids)

		gotArr, ok := got[0].Value.(bson.D)[0].Value.(bson.A)
		if !ok || len(gotArr) != 1 {
			t.Fatalf("In(field, ids) without spread = %v, want exactly one candidate holding the whole slice", got)
		}
	})
}

func ExampleIn() {
	filter := monq.In("status", "active", "pending")

	fmt.Println(filter)
	// Output: {"status":{"$in":["active","pending"]}}
}

func TestNin(t *testing.T) {
	tests := []struct {
		name   string
		field  monq.FieldPath
		values []any
		want   bson.A
	}{
		{
			name:   "multiple spread values",
			field:  "status",
			values: []any{"banned", "suspended"},
			want:   bson.A{"banned", "suspended"},
		},
		{
			name:   "single value",
			field:  "status",
			values: []any{"banned"},
			want:   bson.A{"banned"},
		},
		{
			name:   "no values",
			field:  "status",
			values: nil,
			want:   bson.A{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Nin(tt.field, tt.values...)

			inner, ok := got[0].Value.(bson.D)
			if !ok || inner[0].Key != "$nin" {
				t.Fatalf("Nin() = %v, want a $nin operator document", got)
			}

			gotArr, ok := inner[0].Value.(bson.A)
			if !ok || len(gotArr) != len(tt.want) {
				t.Fatalf("Nin() = %v, want values %v", got, tt.want)
			}
		})
	}
}

func ExampleNin() {
	filter := monq.Nin("status", "banned", "suspended")

	fmt.Println(filter)
	// Output: {"status":{"$nin":["banned","suspended"]}}
}
