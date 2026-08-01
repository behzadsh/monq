package monq_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestAll(t *testing.T) {
	tests := []struct {
		name   string
		field  monq.FieldPath
		values []any
		want   bson.A
	}{
		{
			name:   "two values",
			field:  "tags",
			values: []any{"go", "mongodb"},
			want:   bson.A{"go", "mongodb"},
		},
		{
			name:   "single value",
			field:  "tags",
			values: []any{"go"},
			want:   bson.A{"go"},
		},
		{
			name:   "un-spread slice stays one element",
			field:  "tags",
			values: []any{[]string{"go", "mongodb"}},
			want:   bson.A{[]string{"go", "mongodb"}},
		},
		{
			name:   "no values",
			field:  "tags",
			values: nil,
			want:   bson.A{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.All(tt.field, tt.values...)

			inner, ok := got[0].Value.(bson.D)
			if !ok || got[0].Key != string(tt.field) || inner[0].Key != "$all" {
				t.Fatalf("All() = %v, want a $all operator document on %q", got, tt.field)
			}

			arr, ok := inner[0].Value.(bson.A)
			if !ok || len(arr) != len(tt.want) {
				t.Fatalf("All() values = %v, want %v", inner[0].Value, tt.want)
			}
		})
	}
}

func ExampleAll() {
	filter := monq.All("tags", "go", "mongodb")

	fmt.Println(filter)
	// Output: {"tags":{"$all":["go","mongodb"]}}
}

func TestElemMatch(t *testing.T) {
	tests := []struct {
		name    string
		field   monq.FieldPath
		filters []bson.D
		want    bson.D
	}{
		{
			name:  "two criteria on different fields",
			field: "items",
			filters: []bson.D{
				monq.Eq("sku", "abc"),
				monq.Gte("qty", 2),
			},
			want: bson.D{
				{Key: "sku", Value: bson.D{{Key: "$eq", Value: "abc"}}},
				{Key: "qty", Value: bson.D{{Key: "$gte", Value: 2}}},
			},
		},
		{
			name:    "bare operator expression for arrays of scalars",
			field:   "scores",
			filters: []bson.D{monq.Raw(bson.D{{Key: "$gte", Value: 80}, {Key: "$lt", Value: 85}})},
			want:    bson.D{{Key: "$gte", Value: 80}, {Key: "$lt", Value: 85}},
		},
		{
			name:    "no filters",
			field:   "items",
			filters: nil,
			want:    bson.D{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.ElemMatch(tt.field, tt.filters...)

			inner, ok := got[0].Value.(bson.D)
			if !ok || got[0].Key != string(tt.field) || inner[0].Key != "$elemMatch" {
				t.Fatalf("ElemMatch() = %v, want a $elemMatch operator document on %q", got, tt.field)
			}

			criteria, ok := inner[0].Value.(bson.D)
			if !ok || len(criteria) != len(tt.want) {
				t.Fatalf("ElemMatch() criteria = %v, want %v", inner[0].Value, tt.want)
			}

			for i, e := range tt.want {
				if criteria[i].Key != e.Key {
					t.Fatalf("ElemMatch() criteria[%d] = %q, want %q", i, criteria[i].Key, e.Key)
				}
			}
		})
	}
}

func ExampleElemMatch() {
	filter := monq.ElemMatch("items", monq.Eq("sku", "abc"), monq.Gte("qty", 2))

	printFilter(filter)
	// Output: {"items":{"$elemMatch":{"sku":{"$eq":"abc"},"qty":{"$gte":2}}}}
}

func TestSize(t *testing.T) {
	tests := []struct {
		name  string
		field monq.FieldPath
		size  int
		want  bson.D
	}{
		{
			name:  "three elements",
			field: "tags",
			size:  3,
			want:  bson.D{{Key: "tags", Value: bson.D{{Key: "$size", Value: 3}}}},
		},
		{
			name:  "empty array",
			field: "tags",
			size:  0,
			want:  bson.D{{Key: "tags", Value: bson.D{{Key: "$size", Value: 0}}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Size(tt.field, tt.size)

			inner, ok := got[0].Value.(bson.D)
			if !ok || got[0].Key != string(tt.field) || inner[0].Key != "$size" || inner[0].Value != tt.size {
				t.Fatalf("Size() = %v, want %v", got, tt.want)
			}
		})
	}
}

func ExampleSize() {
	filter := monq.Size("tags", 3)

	printFilter(filter)
	// Output: {"tags":{"$size":3}}
}

func TestArrayOperatorsComposeWithNot(t *testing.T) {
	tests := []struct {
		name string
		expr bson.D
		op   string
	}{
		{name: "negates All", expr: monq.All("tags", "go"), op: "$all"},
		{name: "negates Size", expr: monq.Size("tags", 3), op: "$size"},
		{name: "negates ElemMatch", expr: monq.ElemMatch("items", monq.Eq("sku", "abc")), op: "$elemMatch"},
		{name: "negates Type", expr: monq.Type("legacyId", "string"), op: "$type"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Not(tt.expr)

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
