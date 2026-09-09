package example_test

import (
	"fmt"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/cmd/monqgen/internal/example"
	"github.com/behzadsh/monq/expr"
	"github.com/behzadsh/monq/index"
	"github.com/behzadsh/monq/stage"
)

// printDoc prints a document as relaxed extended JSON, the same helper the other packages' tests carry.
func printDoc(d bson.D) {
	b, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

func TestGeneratedPaths(t *testing.T) {
	tests := []struct {
		name string
		got  monq.FieldPath
		want monq.FieldPath
	}{
		{name: "a tagged field", got: example.UserPaths.Email, want: "email"},
		{name: "an id", got: example.UserPaths.ID, want: "_id"},
		{name: "a date is a leaf", got: example.UserPaths.CreatedAt, want: "created_at"},
		{name: "a nested document", got: example.UserPaths.Address.Street, want: "address.street"},
		{name: "an array of scalars", got: example.UserPaths.Tags.Path, want: "tags"},
		{name: "an array of documents", got: example.UserPaths.Items.Path, want: "items"},
		{name: "the dotted element form", got: example.UserPaths.Items.SKU, want: "items.sku"},
		{name: "a fixed position", got: example.UserPaths.Items.At(3).Quantity, want: "items.3.qty"},
		{name: "the matched element", got: example.UserPaths.Items.Positional().Price, want: "items.$.price"},
		{name: "every element", got: example.UserPaths.Items.All().SKU, want: "items.$[].sku"},
		{name: "an array filter", got: example.UserPaths.Items.Filtered("cheap").Price, want: "items.$[cheap].price"},
		{name: "a scalar array position", got: example.UserPaths.Tags.At(0), want: "tags.0"},
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

func TestGeneratedPathsBuildAFilter(t *testing.T) {
	filter := monq.And(
		monq.Eq(example.UserPaths.Email, "ada@example.com"),
		monq.ElemMatch(
			example.UserPaths.Items.Path,
			monq.Eq(example.UserPaths.Items.SKU, "abc"),
		),
	)

	if len(filter) != 1 || filter[0].Key != "$and" {
		t.Fatalf("And() = %v, want one $and", filter)
	}
}

func TestGeneratedPathsBuildAnUpdate(t *testing.T) {
	update := monq.Update(
		monq.Set(example.UserPaths.Items.Positional().Quantity, 5),
		monq.Inc(example.UserPaths.Items.Filtered("cheap").Price, 1),
		monq.AddToSet(example.UserPaths.Tags.Path, "vip"),
	)

	want := 3
	if len(update) != want {
		t.Fatalf("Update() = %v, want %d operators", update, want)
	}
}

func TestGeneratedPathsBuildAPipeline(t *testing.T) {
	pipeline := stage.Pipeline(
		stage.Match(monq.Gte(example.UserPaths.CreatedAt, "2026-01-01")),
		stage.Unwind(expr.Field(example.UserPaths.Items.Path)),
		stage.Group(
			expr.Field(example.UserPaths.Items.SKU),
			stage.Accumulator("revenue", expr.Sum(expr.Field(example.UserPaths.Items.Price))),
		),
		stage.Sort(monq.Sort(monq.Desc("revenue"))),
	)

	want := 4
	if len(pipeline) != want {
		t.Fatalf("Pipeline() = %v stages, want %d", len(pipeline), want)
	}
}

func TestGeneratedPathsBuildAnIndex(t *testing.T) {
	model := index.New(
		index.Asc(example.UserPaths.Email),
		index.Desc(example.UserPaths.CreatedAt),
		index.Unique(),
	)

	keys, ok := model.Keys.(bson.D)
	if !ok || len(keys) != 2 || keys[0].Key != "email" {
		t.Fatalf("Keys = %v, want email and created_at", model.Keys)
	}
}

func ExampleUserPaths() {
	filter := monq.And(
		monq.Eq(example.UserPaths.Email, "ada@example.com"),
		monq.Gte(example.UserPaths.Items.Price, 10),
	)

	printDoc(filter)
	// Output: {"$and":[{"email":{"$eq":"ada@example.com"}},{"items.price":{"$gte":10}}]}
}

func ExampleUserPaths_array() {
	update := monq.Update(
		monq.Set(example.UserPaths.Items.Positional().Quantity, 5),
		monq.Inc(example.UserPaths.Items.Filtered("cheap").Price, 1),
	)

	printDoc(update)
	// Output: {"$set":{"items.$.qty":5},"$inc":{"items.$[cheap].price":1}}
}
