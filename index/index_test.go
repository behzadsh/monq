package index_test

import (
	"reflect"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/index"
)

// resolve applies a model's option setters, which is how the driver turns a builder into the options it sends.
func resolve(t *testing.T, builder *options.IndexOptionsBuilder) options.IndexOptions {
	t.Helper()

	var opts options.IndexOptions
	for _, set := range builder.List() {
		if err := set(&opts); err != nil {
			t.Fatalf("applying index options: %v", err)
		}
	}

	return opts
}

func TestNewKeys(t *testing.T) {
	tests := []struct {
		name string
		opts []index.Option
		want bson.D
	}{
		{
			name: "no options at all",
			opts: nil,
			want: bson.D{},
		},
		{
			name: "one ascending key",
			opts: []index.Option{index.Asc("email")},
			want: bson.D{{Key: "email", Value: 1}},
		},
		{
			name: "a compound key keeps the order it is given",
			opts: []index.Option{index.Asc("email"), index.Desc("created_at")},
			want: bson.D{
				{Key: "email", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			name: "keys and properties interleave",
			opts: []index.Option{index.Asc("email"), index.Unique(), index.Desc("created_at")},
			want: bson.D{
				{Key: "email", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
		{
			name: "text keys across two fields",
			opts: []index.Option{index.Text("title"), index.Text("body")},
			want: bson.D{
				{Key: "title", Value: "text"},
				{Key: "body", Value: "text"},
			},
		},
		{
			name: "hashed key",
			opts: []index.Option{index.Hashed("user_id")},
			want: bson.D{{Key: "user_id", Value: "hashed"}},
		},
		{
			name: "geospatial keys",
			opts: []index.Option{index.Geo2DSphere("loc"), index.Geo2D("legacy_loc")},
			want: bson.D{
				{Key: "loc", Value: "2dsphere"},
				{Key: "legacy_loc", Value: "2d"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := index.New(tt.opts...)

			if !reflect.DeepEqual(got.Keys, tt.want) {
				t.Fatalf("Keys = %v, want %v", got.Keys, tt.want)
			}
		})
	}
}

func TestNewOptions(t *testing.T) {
	model := index.New(
		index.Asc("email"),
		index.Unique(),
		index.Sparse(),
		index.Hidden(),
		index.Name("email_unique"),
		index.TTL(24*time.Hour),
		index.PartialFilter(monq.Exists("deleted_at", false)),
	)

	opts := resolve(t, model.Options)

	wantSet(t, "Unique", opts.Unique, true)
	wantSet(t, "Sparse", opts.Sparse, true)
	wantSet(t, "Hidden", opts.Hidden, true)
	wantSet(t, "Name", opts.Name, "email_unique")
	wantSet(t, "ExpireAfterSeconds", opts.ExpireAfterSeconds, int32(86400))

	want := monq.Exists("deleted_at", false)
	if !reflect.DeepEqual(opts.PartialFilterExpression, want) {
		t.Errorf("PartialFilterExpression = %v, want %v", opts.PartialFilterExpression, want)
	}
}

// wantSet reports whether an option was set to the expected value, treating a nil pointer as unset.
func wantSet[T comparable](t *testing.T, name string, got *T, want T) {
	t.Helper()

	if got == nil {
		t.Errorf("%s is unset, want %v", name, want)

		return
	}

	if *got != want {
		t.Errorf("%s = %v, want %v", name, *got, want)
	}
}

func TestTTLTruncatesToWholeSeconds(t *testing.T) {
	tests := []struct {
		name string
		ttl  time.Duration
		want int32
	}{
		{name: "a day", ttl: 24 * time.Hour, want: 86400},
		{name: "sub-second precision is dropped", ttl: 1500 * time.Millisecond, want: 1},
		{name: "zero expires at the stored date", ttl: 0, want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := resolve(t, index.New(index.Asc("created_at"), index.TTL(tt.ttl)).Options)

			if opts.ExpireAfterSeconds == nil || *opts.ExpireAfterSeconds != tt.want {
				t.Fatalf("ExpireAfterSeconds = %v, want %d", opts.ExpireAfterSeconds, tt.want)
			}
		})
	}
}

func TestNewWithoutOptionsCarriesAnEmptyBuilder(t *testing.T) {
	model := index.New(index.Asc("email"))

	if model.Options == nil {
		t.Fatal("Options = nil, want an empty builder")
	}

	if got := len(model.Options.List()); got != 0 {
		t.Fatalf("Options holds %d setters, want none", got)
	}
}

func ExampleNew() {
	model := index.New(index.Asc("email"), index.Unique())

	printKeys(model.Keys)
	// Output: {"email":1}
}

func ExampleAsc() {
	model := index.New(index.Asc("email"), index.Desc("created_at"))

	printKeys(model.Keys)
	// Output: {"email":1,"created_at":-1}
}

func ExampleText() {
	model := index.New(index.Text("title"), index.Text("body"))

	printKeys(model.Keys)
	// Output: {"title":"text","body":"text"}
}

func ExampleTTL() {
	model := index.New(index.Asc("created_at"), index.TTL(24*time.Hour))

	printKeys(model.Keys)
	// Output: {"created_at":1}
}
