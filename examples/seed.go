package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/index"
)

// Collection names, so a typo is a compile error in one place rather than an empty result set somewhere else.
const (
	productsColl   = "products"
	ordersColl     = "orders"
	categoriesColl = "categories"
)

// The warehouse field is the embedded document holding a product's GeoJSON point. Generated paths name the
// fields inside it (ProductPaths.Warehouse.Type and .Coordinates) but not the document itself, which is what a
// 2dsphere index and a $near query are on.
const warehousePath monq.FieldPath = "warehouse"

// seedTime is the instant the fixture dates are measured from, so a document's age is the same on every run.
var seedTime = time.Date(2026, time.March, 2, 9, 0, 0, 0, time.UTC)

// seed drops the database and refills it, so every run starts from the same documents no matter what the update
// and aggregation demos did last time.
func seed(ctx context.Context, db *mongo.Database) error {
	if err := db.Drop(ctx); err != nil {
		return fmt.Errorf("drop database: %w", err)
	}

	if _, err := db.Collection(productsColl).InsertMany(ctx, products()); err != nil {
		return fmt.Errorf("insert products: %w", err)
	}

	if _, err := db.Collection(ordersColl).InsertMany(ctx, orders()); err != nil {
		return fmt.Errorf("insert orders: %w", err)
	}

	if _, err := db.Collection(categoriesColl).InsertMany(ctx, categories()); err != nil {
		return fmt.Errorf("insert categories: %w", err)
	}

	return seedIndexes(ctx, db)
}

// seedIndexes creates the two indexes the demos query through: MongoDB rejects a $near or a $text search outright
// when the collection has no index supporting it.
func seedIndexes(ctx context.Context, db *mongo.Database) error {
	models := []mongo.IndexModel{
		index.New(index.Geo2DSphere(warehousePath), index.Name("warehouse_2dsphere")),
		index.New(index.Text(ProductPaths.Name), index.Text(ProductPaths.Tags.Path), index.Name("search_text")),
	}

	if _, err := db.Collection(productsColl).Indexes().CreateMany(ctx, models); err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}

	return nil
}

// products is the fixture set the demos query. The fields are chosen for the operators: some products have no
// ratings, one is out of stock, prices and dates spread out, and every warehouse is somewhere real.
func products() []any {
	return []any{
		Product{
			ID: bson.NewObjectID(), SKU: "kbd-001", Name: "Wireless Keyboard", Category: "peripherals",
			Price: 89.99, Stock: 12, Flags: 5, Tags: []string{"wireless", "rgb"},
			Ratings:   []Rating{{User: "ada", Score: 5}, {User: "grace", Score: 4}},
			Warehouse: point(-73.97, 40.77), CreatedAt: seedTime.AddDate(0, -10, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "mse-002", Name: "Wireless Mouse", Category: "peripherals",
			Price: 39.50, Stock: 0, Flags: 3, Tags: []string{"wireless"},
			Ratings:   []Rating{{User: "ada", Score: 3}},
			Warehouse: point(-73.97, 40.77), CreatedAt: seedTime.AddDate(0, -8, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "mon-003", Name: "4K Monitor", Category: "displays",
			Price: 249.00, Stock: 5, Flags: 6, Tags: []string{"4k", "hdr"},
			Ratings:   []Rating{{User: "linus", Score: 5}, {User: "ada", Score: 4}, {User: "grace", Score: 5}},
			Warehouse: point(-87.63, 41.88), CreatedAt: seedTime.AddDate(0, -6, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "cbl-004", Name: "USB-C Cable", Category: "accessories",
			Price: 9.99, Stock: 250, Flags: 1, Tags: []string{"usb-c"},
			Ratings:   []Rating{{User: "grace", Score: 2}},
			Warehouse: point(-87.63, 41.88), CreatedAt: seedTime.AddDate(0, -4, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "hub-005", Name: "USB-C Hub", Category: "accessories",
			Price: 24.95, Stock: 40, Flags: 9, Tags: []string{"usb-c", "powered"},
			Warehouse: point(-122.33, 47.61), CreatedAt: seedTime.AddDate(0, -3, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "lap-006", Name: "Developer Laptop", Category: "computers",
			Price: 1299.00, Stock: 3, Flags: 12, Tags: []string{"16gb", "ssd"},
			Ratings:   []Rating{{User: "linus", Score: 4}},
			Warehouse: point(-122.33, 47.61), CreatedAt: seedTime.AddDate(0, -2, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "cam-007", Name: "1080p Webcam", Category: "peripherals",
			Price: 59.00, Stock: 0, Flags: 2, Tags: []string{"1080p"},
			Warehouse: point(-73.97, 40.77), CreatedAt: seedTime.AddDate(0, -1, 0),
		},
		Product{
			ID: bson.NewObjectID(), SKU: "dsk-008", Name: "Standing Desk", Category: "furniture",
			Price: 399.00, Stock: 7, Flags: 8, Tags: []string{"standing", "oak"},
			Ratings:   []Rating{{User: "grace", Score: 5}},
			Warehouse: point(-97.74, 30.27), CreatedAt: seedTime.AddDate(0, 0, -10),
		},
	}
}

// orders is the second collection, there so the aggregation demo has something to join against.
func orders() []any {
	return []any{
		Order{
			ID: bson.NewObjectID(), Customer: "ada", Status: "shipped", Total: 129.49,
			Items:    []OrderItem{{SKU: "kbd-001", Quantity: 1, Price: 89.99}, {SKU: "mse-002", Quantity: 1, Price: 39.50}},
			PlacedAt: seedTime.AddDate(0, -2, 0),
		},
		Order{
			ID: bson.NewObjectID(), Customer: "grace", Status: "shipped", Total: 259.98,
			Items:    []OrderItem{{SKU: "cbl-004", Quantity: 1, Price: 9.99}, {SKU: "mon-003", Quantity: 1, Price: 249.00}},
			PlacedAt: seedTime.AddDate(0, -1, -5),
		},
		Order{
			ID: bson.NewObjectID(), Customer: "ada", Status: "pending", Total: 1299.00,
			Items:    []OrderItem{{SKU: "lap-006", Quantity: 1, Price: 1299.00}},
			PlacedAt: seedTime.AddDate(0, 0, -20),
		},
		Order{
			ID: bson.NewObjectID(), Customer: "linus", Status: "canceled", Total: 49.90,
			Items:    []OrderItem{{SKU: "cbl-004", Quantity: 5, Price: 9.98}},
			PlacedAt: seedTime.AddDate(0, 0, -12),
		},
		Order{
			ID: bson.NewObjectID(), Customer: "grace", Status: "shipped", Total: 448.95,
			Items:    []OrderItem{{SKU: "dsk-008", Quantity: 1, Price: 399.00}, {SKU: "hub-005", Quantity: 2, Price: 24.95}},
			PlacedAt: seedTime.AddDate(0, 0, -3),
		},
	}
}

// categories is a small tree, there so the $graphLookup example has something to walk. It is written as plain
// documents rather than a struct, since one field of it points at another document in the same collection.
func categories() []any {
	return []any{
		bson.D{{Key: "name", Value: "electronics"}, {Key: "parent", Value: nil}},
		bson.D{{Key: "name", Value: "peripherals"}, {Key: "parent", Value: "electronics"}},
		bson.D{{Key: "name", Value: "displays"}, {Key: "parent", Value: "electronics"}},
		bson.D{{Key: "name", Value: "computers"}, {Key: "parent", Value: "electronics"}},
		bson.D{{Key: "name", Value: "accessories"}, {Key: "parent", Value: "peripherals"}},
		bson.D{{Key: "name", Value: "furniture"}, {Key: "parent", Value: nil}},
	}
}

// point builds the GeoJSON point a struct field stores. It is the same document monq.Point returns, which is what
// the geo demo passes to $near, so the two ends of a geospatial query agree on the shape without either one
// hand-writing it.
func point(lon, lat float64) GeoPoint {
	return GeoPoint{Type: "Point", Coordinates: []float64{lon, lat}}
}

// bySKU is the filter naming a single product, which several demos need before they show anything else.
func bySKU(sku string) bson.D {
	return monq.Eq(ProductPaths.SKU, sku)
}
