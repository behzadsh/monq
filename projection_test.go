package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestProjectionEntries(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Include one field",
			got:  monq.Include("email"),
			want: bson.D{{Key: "email", Value: 1}},
		},
		{
			name: "Include several fields and a dotted path",
			got:  monq.Include("email", "profile.name"),
			want: bson.D{{Key: "email", Value: 1}, {Key: "profile.name", Value: 1}},
		},
		{
			name: "Include nothing",
			got:  monq.Include(),
			want: bson.D{},
		},
		{
			name: "Exclude",
			got:  monq.Exclude("password_hash", "internal.notes"),
			want: bson.D{{Key: "password_hash", Value: 0}, {Key: "internal.notes", Value: 0}},
		},
		{
			name: "Slice from the start",
			got:  monq.Slice("comments", 5),
			want: bson.D{{Key: "comments", Value: bson.D{{Key: "$slice", Value: 5}}}},
		},
		{
			name: "Slice from the end",
			got:  monq.Slice("comments", -5),
			want: bson.D{{Key: "comments", Value: bson.D{{Key: "$slice", Value: -5}}}},
		},
		{
			name: "SliceFrom an offset",
			got:  monq.SliceFrom("comments", 10, 5),
			want: bson.D{{Key: "comments", Value: bson.D{{Key: "$slice", Value: bson.A{10, 5}}}}},
		},
		{
			name: "Meta names a computed value",
			got:  monq.Meta("indexKey", "indexKey"),
			want: bson.D{{Key: "indexKey", Value: bson.D{{Key: "$meta", Value: "indexKey"}}}},
		},
		{
			name: "TextScore is Meta with the text kind",
			got:  monq.TextScore("score"),
			want: monq.Meta("score", "textScore"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestProjection(t *testing.T) {
	tests := []struct {
		name    string
		entries []bson.D
		want    bson.D
	}{
		{
			name:    "no entries returns whole documents",
			entries: nil,
			want:    bson.D{},
		},
		{
			name:    "inclusions keep the order they are given",
			entries: []bson.D{monq.Include("email", "items.sku")},
			want:    bson.D{{Key: "email", Value: 1}, {Key: "items.sku", Value: 1}},
		},
		{
			name:    "dropping _id beside inclusions is the one legal mix",
			entries: []bson.D{monq.Include("email"), monq.Exclude("_id")},
			want:    bson.D{{Key: "email", Value: 1}, {Key: "_id", Value: 0}},
		},
		{
			name:    "a later entry replaces an earlier one in place",
			entries: []bson.D{monq.Include("email", "name"), monq.Exclude("email")},
			want:    bson.D{{Key: "email", Value: 0}, {Key: "name", Value: 1}},
		},
		{
			name:    "a slice sits beside inclusions",
			entries: []bson.D{monq.Include("email"), monq.Slice("comments", -5)},
			want: bson.D{
				{Key: "email", Value: 1},
				{Key: "comments", Value: bson.D{{Key: "$slice", Value: -5}}},
			},
		},
		{
			name:    "the query operator doubles as an entry",
			entries: []bson.D{monq.ElemMatch("items", monq.Gt("qty", 5))},
			want: bson.D{{Key: "items", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{{Key: "qty", Value: bson.D{{Key: "$gt", Value: 5}}}}},
			}}},
		},
		{
			name:    "a text score comes back alongside the fields",
			entries: []bson.D{monq.Include("title"), monq.TextScore("score")},
			want: bson.D{
				{Key: "title", Value: 1},
				{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Projection(tt.entries...)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Projection() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProjectionTakesAPositionalPath(t *testing.T) {
	items := monq.ArrayPath{Path: "items"}

	got := monq.Projection(monq.Include(items.Positional()))
	want := bson.D{{Key: "items.$", Value: 1}}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Projection() = %v, want %v", got, want)
	}
}

func TestProjectionLeavesEntriesUntouched(t *testing.T) {
	include := monq.Include("email")

	first := monq.Projection(include, monq.Exclude("_id"))
	second := monq.Projection(include, monq.Exclude("_id"))

	if !reflect.DeepEqual(first, second) {
		t.Fatalf("Projection() = %v on the second call, want %v", second, first)
	}

	if !reflect.DeepEqual(include, bson.D{{Key: "email", Value: 1}}) {
		t.Fatalf("entry = %v, want the caller's document unchanged", include)
	}
}

func ExampleProjection() {
	projection := monq.Projection(monq.Include("email", "items.sku"), monq.Exclude("_id"))

	printFilter(projection)
	// Output: {"email":1,"items.sku":1,"_id":0}
}

func ExampleInclude() {
	entry := monq.Include("email", "profile.name")

	printFilter(entry)
	// Output: {"email":1,"profile.name":1}
}

func ExampleExclude() {
	entry := monq.Exclude("password_hash")

	printFilter(entry)
	// Output: {"password_hash":0}
}

func ExampleSlice() {
	projection := monq.Projection(monq.Slice("comments", -5))

	printFilter(projection)
	// Output: {"comments":{"$slice":-5}}
}

func ExampleSliceFrom() {
	projection := monq.Projection(monq.SliceFrom("comments", 10, 5))

	printFilter(projection)
	// Output: {"comments":{"$slice":[10,5]}}
}

func ExampleMeta() {
	projection := monq.Projection(monq.Include("title"), monq.Meta("score", "textScore"))

	printFilter(projection)
	// Output: {"title":1,"score":{"$meta":"textScore"}}
}
