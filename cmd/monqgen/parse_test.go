package main

import (
	"reflect"
	"testing"

	"golang.org/x/tools/go/packages"
)

// loadFixtures reads the fixture package once per test binary, since loading a package is the slow part.
func loadFixtures(t *testing.T) *packages.Package {
	t.Helper()

	pkg, err := load("./testdata/fixtures")
	if err != nil {
		t.Fatalf("loading fixtures: %v", err)
	}

	return pkg
}

func TestParseUserPaths(t *testing.T) {
	model, err := Parse(loadFixtures(t), "User")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := []string{
		"_id",
		"email",
		"nickname",
		"address",
		"address.street",
		"address.city",
		"billing",
		"billing.street",
		"billing.city",
		"tags",
		"items",
		"items.sku",
		"items.qty",
		"items.price",
		"avatar",
		"anything",
		"labels",
		"timestamps",
		"timestamps.created_at",
		"timestamps.updated_at",
		"version",
		"actor",
		"tree",
		"tree.label",
		"tree.children",
		"joined",
		"balance",
	}

	if got := model.Paths(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Paths() =\n%q\nwant\n%q", got, want)
	}
}

func TestParseFieldKinds(t *testing.T) {
	model, err := Parse(loadFixtures(t), "User")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	kinds := map[string]Kind{}
	recursive := map[string]bool{}

	model.Walk(
		func(path []string, field Field) {
			key := path[len(path)-1]
			if len(path) > 1 {
				key = path[0] + "." + key
			}

			kinds[key] = field.Kind
			recursive[key] = field.Recursive
		},
	)

	tests := []struct {
		path string
		want Kind
	}{
		// bson.ObjectID, time.Time, and bson.Decimal128 all marshal themselves, so they are leaves.
		{path: "_id", want: KindScalar},
		{path: "joined", want: KindScalar},
		{path: "balance", want: KindScalar},
		{path: "email", want: KindScalar},
		// A pointer to a struct is still a document.
		{path: "address", want: KindDocument},
		{path: "billing", want: KindDocument},
		// An embedded struct without inline nests under its own name.
		{path: "timestamps", want: KindDocument},
		{path: "tree", want: KindDocument},
		// Arrays of scalars, of documents, and of opaque elements are all arrays.
		{path: "tags", want: KindArray},
		{path: "items", want: KindArray},
		{path: "anything", want: KindArray},
		{path: "tree.children", want: KindArray},
		// A byte slice is binary data rather than an array, and a map has keys nothing can name ahead of time.
		{path: "avatar", want: KindScalar},
		{path: "labels", want: KindMap},
	}

	for _, tt := range tests {
		t.Run(
			tt.path, func(t *testing.T) {
				if got, ok := kinds[tt.path]; !ok || got != tt.want {
					t.Fatalf("kind of %q = %v, want %v", tt.path, got, tt.want)
				}
			},
		)
	}

	if !recursive["tree.children"] {
		t.Error("tree.children should be marked recursive, since Node holds Nodes")
	}
}

func TestParseSkipsFieldsThatNeverReachTheDocument(t *testing.T) {
	model, err := Parse(loadFixtures(t), "User")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	for _, path := range model.Paths() {
		switch path {
		case "ignored", "-":
			t.Errorf("a field tagged \"-\" produced the path %q", path)
		case "unexposed":
			t.Errorf("an unexported field produced the path %q", path)
		case "extra", "audit":
			t.Errorf("an inlined field produced its own path %q instead of merging into the parent", path)
		default:
		}
	}
}

func TestParseTreatsMarshalersAsLeaves(t *testing.T) {
	model, err := Parse(loadFixtures(t), "Custom")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	want := []string{"name", "blob", "blobs"}
	if got := model.Paths(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Paths() = %q, want %q, since Marshaled encodes itself", got, want)
	}
}

func TestParseErrors(t *testing.T) {
	pkg := loadFixtures(t)

	tests := []struct {
		name     string
		typeName string
	}{
		{name: "no such type", typeName: "Missing"},
		{name: "not a struct", typeName: "Kind"},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if _, err := Parse(pkg, tt.typeName); err == nil {
					t.Fatalf("Parse(%q) error = nil, want an error", tt.typeName)
				}
			},
		)
	}
}

func TestParseTag(t *testing.T) {
	tests := []struct {
		name   string
		goName string
		raw    string
		want   tag
	}{
		{
			name:   "no tag lowercases the whole field name",
			goName: "CreatedAt",
			raw:    "",
			want:   tag{key: "createdat"},
		},
		{
			name:   "a name of its own",
			goName: "CreatedAt",
			raw:    `bson:"created_at"`,
			want:   tag{key: "created_at"},
		},
		{
			name:   "flags after the name are not paths",
			goName: "Email",
			raw:    `bson:"email,omitempty"`,
			want:   tag{key: "email"},
		},
		{
			name:   "an empty name keeps the default",
			goName: "Email",
			raw:    `bson:",omitempty"`,
			want:   tag{key: "email"},
		},
		{
			name:   "a dash drops the field",
			goName: "Ignored",
			raw:    `bson:"-"`,
			want:   tag{key: "ignored", skip: true},
		},
		{
			name:   "inline without a name",
			goName: "Audit",
			raw:    `bson:",inline"`,
			want:   tag{key: "audit", inline: true},
		},
		{
			name:   "another tag is not the bson one",
			goName: "Email",
			raw:    `json:"email_address"`,
			want:   tag{key: "email"},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if got := parseTag(tt.goName, tt.raw); got != tt.want {
					t.Fatalf("parseTag(%q, %q) = %+v, want %+v", tt.goName, tt.raw, got, tt.want)
				}
			},
		)
	}
}

func TestKindString(t *testing.T) {
	tests := []struct {
		kind Kind
		want string
	}{
		{kind: KindScalar, want: "scalar"},
		{kind: KindDocument, want: "document"},
		{kind: KindArray, want: "array"},
		{kind: KindMap, want: "map"},
		{kind: Kind(99), want: "unknown"},
	}

	for _, tt := range tests {
		t.Run(
			tt.want, func(t *testing.T) {
				if got := tt.kind.String(); got != tt.want {
					t.Fatalf("Kind.String() = %q, want %q", got, tt.want)
				}
			},
		)
	}
}
