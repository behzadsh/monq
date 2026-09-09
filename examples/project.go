package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
)

// runProject covers the two documents that shape a Find's answer rather than choose it: the projection saying
// which fields come back, and the sort saying in what order. Both are built from entries that keep the order they
// were given in, which matters for the sort and reads better in the projection.
func runProject(ctx context.Context, db *mongo.Database) error {
	return runParts(
		ctx, db.Collection(productsColl),
		projectFields,
		projectArrays,
		projectSort,
		projectTextScore,
	)
}

// projectFields shows the two ways a projection can go: naming what to keep, or naming what to drop.
func projectFields(ctx context.Context, coll *mongo.Collection) error {
	step("a projection is one document, built from entries, and goes into the driver's own options builder")

	keep := monq.Projection(
		monq.Include(ProductPaths.SKU, ProductPaths.Name, ProductPaths.Price),
		monq.Exclude(ProductPaths.ID),
	)
	query("monq.Projection(monq.Include(sku, name, price), monq.Exclude(_id))", keep)

	if err := showFind(
		ctx, coll, "products under 30, projected", monq.Lt(ProductPaths.Price, 30),
		options.Find().SetProjection(keep).SetSort(monq.Sort(monq.Asc(ProductPaths.SKU))),
	); err != nil {
		return err
	}

	step("the other direction: everything except a few fields, which cannot be mixed with including them")

	drop := monq.Projection(monq.Exclude(ProductPaths.ID, ProductPaths.Ratings.Path, ProductPaths.CreatedAt, warehousePath))
	query("monq.Projection(monq.Exclude(_id, ratings, created_at, warehouse))", drop)

	return showFind(ctx, coll, "one product, most of it", bySKU("mon-003"), options.Find().SetProjection(drop))
}

// projectArrays shows the projection entries that cut an array down instead of dropping it.
func projectArrays(ctx context.Context, coll *mongo.Collection) error {
	step("$slice takes from one end of an array: a positive count from the front, a negative one from the back")

	firstTwo := monq.Projection(
		monq.Include(ProductPaths.SKU),
		monq.Exclude(ProductPaths.ID),
		monq.Slice(ProductPaths.Ratings.Path, 2),
	)
	query("monq.Slice(ProductPaths.Ratings.Path, 2)", firstTwo)

	if err := showFind(ctx, coll, "the first two ratings of one product", bySKU("mon-003"), options.Find().SetProjection(firstTwo)); err != nil {
		return err
	}

	step("SliceFrom is the skip and limit form, which is how a long array is paged through")

	second := monq.Projection(
		monq.Include(ProductPaths.SKU),
		monq.Exclude(ProductPaths.ID),
		monq.SliceFrom(ProductPaths.Ratings.Path, 1, 1),
	)
	query("monq.SliceFrom(ProductPaths.Ratings.Path, 1, 1)", second)

	return showFind(ctx, coll, "the second rating alone", bySKU("mon-003"), options.Find().SetProjection(second))
}

// projectSort shows that sort entries keep their order, which is the order MongoDB applies them in.
func projectSort(ctx context.Context, coll *mongo.Collection) error {
	step("Sort keeps the entries in the order given, so category groups the results and price orders each group")

	order := monq.Sort(monq.Asc(ProductPaths.Category), monq.Desc(ProductPaths.Price))
	query("monq.Sort(monq.Asc(category), monq.Desc(price))", order)

	opts := options.Find().
		SetSort(order).
		SetLimit(5).
		SetProjection(monq.Projection(monq.Include(ProductPaths.Category, ProductPaths.Price), monq.Exclude(ProductPaths.ID)))

	return showFind(ctx, coll, "the five most expensive by category", monq.Exists(ProductPaths.Category, true), opts)
}

// projectTextScore shows the one projection entry that computes something, and the sort that goes with it.
func projectTextScore(ctx context.Context, coll *mongo.Collection) error {
	step("$meta pulls the text search score out into a field of its own, and sorts on it the same way")

	scored := monq.Projection(
		monq.Include(ProductPaths.SKU, ProductPaths.Name),
		monq.Exclude(ProductPaths.ID),
		monq.Meta("score", "textScore"),
	)
	query(`monq.Meta("score", "textScore")`, scored)
	query(`monq.Sort(monq.TextScore("score"))`, monq.Sort(monq.TextScore("score")))

	opts := options.Find().SetProjection(scored).SetSort(monq.Sort(monq.TextScore("score")))

	return showFind(ctx, coll, `monq.Text("usb wireless")`, monq.Text("usb wireless"), opts)
}
