package main

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

//go:generate go run github.com/behzadsh/monq/cmd/monqgen -type Product
//go:generate go run github.com/behzadsh/monq/cmd/monqgen -type Order

// Rating is one customer rating of a product, stored as an array element.
type Rating struct {
	User  string `bson:"user"`
	Score int    `bson:"score"`
}

// GeoPoint is a GeoJSON point, the shape a 2dsphere index understands. Coordinates are [longitude, latitude],
// which is the order monq.Point takes them in too.
type GeoPoint struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

// Product is the document the products collection stores. Ratings is tagged omitempty so that a product nobody
// rated has no ratings field at all rather than a null one, which is the difference $exists reports on.
type Product struct {
	ID        bson.ObjectID `bson:"_id"`
	SKU       string        `bson:"sku"`
	Name      string        `bson:"name"`
	Category  string        `bson:"category"`
	Price     float64       `bson:"price"`
	Stock     int           `bson:"stock"`
	Flags     int           `bson:"flags"`
	Tags      []string      `bson:"tags"`
	Ratings   []Rating      `bson:"ratings,omitempty"`
	Warehouse GeoPoint      `bson:"warehouse"`
	CreatedAt time.Time     `bson:"created_at"`
}

// OrderItem is one line of an order. It repeats the product's sku, which is what the $lookup in the aggregation
// example joins on.
type OrderItem struct {
	SKU      string  `bson:"sku"`
	Quantity int     `bson:"qty"`
	Price    float64 `bson:"price"`
}

// Order is the document the orders collection stores.
type Order struct {
	ID       bson.ObjectID `bson:"_id"`
	Customer string        `bson:"customer"`
	Status   string        `bson:"status"`
	Items    []OrderItem   `bson:"items"`
	Total    float64       `bson:"total"`
	PlacedAt time.Time     `bson:"placed_at"`
}
