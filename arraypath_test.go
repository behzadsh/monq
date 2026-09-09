package monq_test

import (
	"fmt"
	"testing"

	"github.com/behzadsh/monq"
)

func TestArrayPath(t *testing.T) {
	items := monq.ArrayPath{Path: "items"}

	tests := []struct {
		name string
		got  monq.FieldPath
		want monq.FieldPath
	}{
		{name: "the array itself", got: items.Path, want: "items"},
		{name: "a fixed position", got: items.At(3), want: "items.3"},
		{name: "the first position", got: items.At(0), want: "items.0"},
		{name: "the element a query matched", got: items.Positional(), want: "items.$"},
		{name: "every element", got: items.All(), want: "items.$[]"},
		{name: "the elements an array filter picks", got: items.Filtered("cheap"), want: "items.$[cheap]"},
		{name: "a dotted array path", got: monq.ArrayPath{Path: "order.items"}.At(1), want: "order.items.1"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.got != tt.want {
					t.Fatalf("got %q, want %q", tt.got, tt.want)
				}
			},
		)
	}
}

func TestArrayPathComposesWithOperators(t *testing.T) {
	items := monq.ArrayPath{Path: "items"}

	got := monq.Update(
		monq.Set(items.Positional(), "sold"),
		monq.Inc(items.At(0), 1),
	)

	want := 2
	if len(got) != want {
		t.Fatalf("Update() = %v, want %d operators", got, want)
	}
}

func ExampleArrayPath() {
	tags := monq.ArrayPath{Path: "tags"}

	fmt.Println(tags.Path, tags.At(0), tags.Positional(), tags.All(), tags.Filtered("old"))
	// Output: tags tags.0 tags.$ tags.$[] tags.$[old]
}

func ExampleArrayPath_At() {
	update := monq.Set(monq.ArrayPath{Path: "tags"}.At(0), "go")

	printFilter(update)
	// Output: {"$set":{"tags.0":"go"}}
}

func ExampleArrayPath_Positional() {
	update := monq.Set(monq.ArrayPath{Path: "items"}.Positional(), "sold")

	printFilter(update)
	// Output: {"$set":{"items.$":"sold"}}
}

func ExampleArrayPath_All() {
	update := monq.Set(monq.ArrayPath{Path: "items"}.All(), "sold")

	printFilter(update)
	// Output: {"$set":{"items.$[]":"sold"}}
}

func ExampleArrayPath_Filtered() {
	update := monq.Inc(monq.ArrayPath{Path: "items"}.Filtered("cheap"), 1)

	printFilter(update)
	// Output: {"$inc":{"items.$[cheap]":1}}
}
