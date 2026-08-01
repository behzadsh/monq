package monq_test

import (
	"fmt"
	"reflect"
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

func TestExpr(t *testing.T) {
	tests := []struct {
		name       string
		expression any
	}{
		{
			name:       "compares two fields",
			expression: bson.D{{Key: "$gt", Value: bson.A{"$spent", "$budget"}}},
		},
		{
			name:       "passes a non-document expression through",
			expression: "$isActive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Expr(tt.expression)

			if got[0].Key != "$expr" || !reflect.DeepEqual(got[0].Value, tt.expression) {
				t.Fatalf("Expr() = %v, want {$expr: %v}", got, tt.expression)
			}
		})
	}
}

func ExampleExpr() {
	filter := monq.Expr(bson.D{{Key: "$gt", Value: bson.A{"$spent", "$budget"}}})

	fmt.Println(filter)
	// Output: {"$expr":{"$gt":["$spent","$budget"]}}
}

func TestJSONSchema(t *testing.T) {
	tests := []struct {
		name   string
		schema bson.D
	}{
		{
			name:   "required fields",
			schema: bson.D{{Key: "required", Value: bson.A{"email"}}},
		},
		{
			name:   "empty schema",
			schema: bson.D{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.JSONSchema(tt.schema)

			if got[0].Key != "$jsonSchema" || !reflect.DeepEqual(got[0].Value, tt.schema) {
				t.Fatalf("JSONSchema() = %v, want {$jsonSchema: %v}", got, tt.schema)
			}
		})
	}
}

func ExampleJSONSchema() {
	filter := monq.JSONSchema(bson.D{{Key: "required", Value: bson.A{"email"}}})

	fmt.Println(filter)
	// Output: {"$jsonSchema":{"required":["email"]}}
}

func TestMod(t *testing.T) {
	tests := []struct {
		name      string
		field     monq.FieldPath
		divisor   int64
		remainder int64
	}{
		{
			name:      "even values",
			field:     "qty",
			divisor:   2,
			remainder: 0,
		},
		{
			name:      "every fourth value offset by one",
			field:     "qty",
			divisor:   4,
			remainder: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Mod(tt.field, tt.divisor, tt.remainder)

			inner, ok := got[0].Value.(bson.D)
			if !ok || got[0].Key != string(tt.field) || inner[0].Key != "$mod" {
				t.Fatalf("Mod() = %v, want a $mod operator document on %q", got, tt.field)
			}

			want := bson.A{tt.divisor, tt.remainder}
			if !reflect.DeepEqual(inner[0].Value, want) {
				t.Fatalf("Mod() value = %v, want %v", inner[0].Value, want)
			}
		})
	}
}

func ExampleMod() {
	filter := monq.Mod("qty", 4, 0)

	printFilter(filter)
	// Output: {"qty":{"$mod":[4,0]}}
}

func TestText(t *testing.T) {
	tests := []struct {
		name   string
		search string
		opts   []monq.TextOption
		want   bson.D
	}{
		{
			name:   "search terms only",
			search: "coffee shop",
			want:   bson.D{{Key: "$search", Value: "coffee shop"}},
		},
		{
			name:   "with a language",
			search: "coffee",
			opts:   []monq.TextOption{monq.Language("en")},
			want: bson.D{
				{Key: "$search", Value: "coffee"},
				{Key: "$language", Value: "en"},
			},
		},
		{
			name:   "with every option",
			search: "café",
			opts:   []monq.TextOption{monq.Language("fr"), monq.CaseSensitive(), monq.DiacriticSensitive()},
			want: bson.D{
				{Key: "$search", Value: "café"},
				{Key: "$language", Value: "fr"},
				{Key: "$caseSensitive", Value: true},
				{Key: "$diacriticSensitive", Value: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Text(tt.search, tt.opts...)

			inner, ok := got[0].Value.(bson.D)
			if !ok || got[0].Key != "$text" {
				t.Fatalf("Text() = %v, want a $text document", got)
			}

			if !reflect.DeepEqual(inner, tt.want) {
				t.Fatalf("Text() = %v, want %v", inner, tt.want)
			}
		})
	}
}

func ExampleText() {
	filter := monq.Text("coffee shop", monq.Language("en"), monq.CaseSensitive())

	fmt.Println(filter)
	// Output: {"$text":{"$search":"coffee shop","$language":"en","$caseSensitive":true}}
}
