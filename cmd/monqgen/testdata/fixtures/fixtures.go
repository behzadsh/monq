// Package fixtures holds the struct shapes the parser is tested against. Every field in it stands for one rule of
// the driver's bson tag handling, or one shape the path tree has to cope with.
package fixtures

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Timestamps is embedded by [User] twice over, once plainly and once inlined, to show that an embedded struct is
// only flattened when the tag says so.
type Timestamps struct {
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// Address is a plain nested document.
type Address struct {
	Street string `bson:"street"`
	City   string `bson:"city"`
}

// Item is the element of an array of documents.
type Item struct {
	SKU      string  `bson:"sku"`
	Quantity int     `bson:"qty"`
	Price    float64 `bson:"price"`
}

// Node refers to itself, which the walk has to stop rather than follow forever.
type Node struct {
	Label    string  `bson:"label"`
	Children []*Node `bson:"children"`
}

// Audit is inlined into [User], so its fields belong to the parent document.
type Audit struct {
	Version int    `bson:"version"`
	Actor   string `bson:"actor"`
}

// User covers the whole rule set in one struct.
type User struct {
	ID        bson.ObjectID     `bson:"_id"`
	Email     string            `bson:"email"`
	Nickname  string            // no tag: the key is the field name lowercased
	Ignored   string            `bson:"-"`
	unexposed string            //nolint:unused // unexported fields never reach the document
	Address   Address           `bson:"address"`
	Billing   *Address          `bson:"billing"`
	Tags      []string          `bson:"tags"`
	Items     []Item            `bson:"items"`
	Avatar    []byte            `bson:"avatar"`
	Anything  []any             `bson:"anything"`
	Labels    map[string]string `bson:"labels"`
	Timestamps                  // embedded and not inlined: nests under "timestamps"
	Audit     `bson:",inline"`  // inlined: version and actor belong to User
	Extra     bson.M            `bson:",inline"`
	Tree      Node              `bson:"tree"`
	Joined    time.Time         `bson:"joined"`
	Balance   bson.Decimal128   `bson:"balance"`
}

// Marshaled has a marshaler of its own, so the document it produces has nothing to do with its fields and the walk
// must treat it as a leaf.
type Marshaled struct {
	Hidden string `bson:"hidden"`
}

// MarshalBSON makes [Marshaled] responsible for its own encoding.
func (Marshaled) MarshalBSON() ([]byte, error) { return nil, nil }

// Custom holds a field whose type marshals itself.
type Custom struct {
	Name  string    `bson:"name"`
	Blob  Marshaled `bson:"blob"`
	Blobs []Marshaled
}
