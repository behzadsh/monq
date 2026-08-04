// Package example is a worked example of monqgen: a struct, the paths generated from it, and tests using those
// paths against the monq API.
//
// The generated file next to this one is committed rather than produced at build time, so it is compiled by every
// build and read by anyone wondering what the tool writes. A test regenerates it and fails if the two have drifted
// apart, which is what keeps the committed copy honest.
package example

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

//go:generate go run github.com/behzadsh/monq/cmd/monqgen -type User

// Address is a nested document.
type Address struct {
	Street string `bson:"street"`
	City   string `bson:"city"`
}

// Item is one element of the array of line items.
type Item struct {
	SKU      string  `bson:"sku"`
	Quantity int     `bson:"qty"`
	Price    float64 `bson:"price"`
}

// User is the document this example stores.
type User struct {
	ID        bson.ObjectID `bson:"_id"`
	Email     string        `bson:"email"`
	Address   Address       `bson:"address"`
	Tags      []string      `bson:"tags"`
	Items     []Item        `bson:"items"`
	CreatedAt time.Time     `bson:"created_at"`
}
