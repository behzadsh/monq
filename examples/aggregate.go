package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
	"github.com/behzadsh/monq/stage"
)

// runAggregate covers the aggregation framework: stages from monq/stage, and the expressions they compute with
// from monq/expr. The split is what keeps the names honest, since stage.Set is the $set stage while monq.Set is
// the $set update operator, and expr.Eq compares two expressions while monq.Eq tests a field.
func runAggregate(ctx context.Context, db *mongo.Database) error {
	if err := runParts(
		ctx, db.Collection(productsColl),
		aggBasics,
		aggGroup,
		aggComputed,
		aggBuckets,
		aggWindows,
		aggReshape,
		aggOutput,
	); err != nil {
		return err
	}

	if err := runParts(
		ctx, db.Collection(ordersColl),
		aggUnwind,
		aggLookup,
		aggSeries,
	); err != nil {
		return err
	}

	if err := runParts(ctx, db.Collection(categoriesColl), aggGraph); err != nil {
		return err
	}

	return aggDocuments(ctx, db)
}

// aggBasics shows that a pipeline is a slice of stage documents, which is exactly what Aggregate takes.
func aggBasics(ctx context.Context, coll *mongo.Collection) error {
	step("Pipeline returns []bson.D, the type mongo.Pipeline is defined as, so it goes into Aggregate as it is")

	stages := stage.Pipeline(
		stage.Match(monq.Gt(ProductPaths.Price, 50)),
		stage.Sort(monq.Sort(monq.Desc(ProductPaths.Price))),
		stage.Limit(3),
		stage.Project(
			stage.Field(ProductPaths.SKU, 1),
			stage.Field(ProductPaths.Price, 1),
			stage.Field(ProductPaths.ID, 0),
		),
	)

	return showAggregate(ctx, coll, "stage.Pipeline(stage.Match(...), stage.Sort(...), stage.Limit(3), stage.Project(...))", stages)
}

// aggGroup shows grouping, where the output fields are accumulators and the accumulators are expressions.
func aggGroup(ctx context.Context, coll *mongo.Collection) error {
	step("$group takes its _id and then one Accumulator per output field")

	stages := stage.Pipeline(
		stage.Group(
			expr.Field(ProductPaths.Category),
			stage.Accumulator("products", expr.Sum(1)),
			stage.Accumulator("stock", expr.Sum(expr.Field(ProductPaths.Stock))),
			stage.Accumulator("avgPrice", expr.Avg(expr.Field(ProductPaths.Price))),
			stage.Accumulator("priciest", expr.Top(monq.Sort(monq.Desc(ProductPaths.Price)), expr.Field(ProductPaths.SKU))),
		),
		stage.Sort(monq.Sort(monq.Desc("stock"))),
	)
	if err := showAggregate(ctx, coll, "stage.Group(expr.Field(category), stage.Accumulator(...), ...)", stages); err != nil {
		return err
	}

	step("Sum reads its argument count: one argument totals a field across the group, several add them up per document")

	counted := stage.Pipeline(
		stage.SortByCount(expr.Field(ProductPaths.Category)),
		stage.Limit(3),
	)

	return showAggregate(ctx, coll, "stage.SortByCount(expr.Field(category))", counted)
}

// aggComputed shows expressions computing new fields, including the conditional forms.
func aggComputed(ctx context.Context, coll *mongo.Collection) error {
	step("expr.Field(price) is the string \"$price\": that leading $ is how the framework tells a field from a constant")

	stages := stage.Pipeline(
		stage.Match(monq.In(ProductPaths.Category, "peripherals", "displays", "computers")),
		stage.AddFields(
			stage.Field("value", expr.Round(expr.Multiply(expr.Field(ProductPaths.Price), expr.Field(ProductPaths.Stock)), 2)),
			stage.Field(
				"tier", expr.Switch(
					expr.Branch(expr.Gte(expr.Field(ProductPaths.Price), 500), "premium"),
					expr.Branch(expr.Gte(expr.Field(ProductPaths.Price), 100), "standard"),
					expr.DefaultCase("budget"),
				),
			),
			stage.Field("available", expr.Cond(expr.Gt(expr.Field(ProductPaths.Stock), 0), "yes", "no")),
		),
		stage.Project(
			stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field("value", 1), stage.Field("tier", 1),
			stage.Field("available", 1),
		),
		stage.Sort(monq.Sort(monq.Asc(ProductPaths.SKU))),
	)
	if err := showAggregate(ctx, coll, "stage.AddFields(stage.Field(\"tier\", expr.Switch(expr.Branch(...), expr.DefaultCase(...))))", stages); err != nil {
		return err
	}

	step(`array expressions reshape a field in place. One $ is a field, two are a variable: the element is "$$this"`)

	arrays := stage.Pipeline(
		stage.Match(monq.Exists(ProductPaths.Ratings.Path, true)),
		stage.Project(
			stage.Field(ProductPaths.ID, 0),
			stage.Field(ProductPaths.SKU, 1),
			stage.Field("scores", expr.Map(expr.Field(ProductPaths.Ratings.Path), "$$this.score")),
			stage.Field("top", expr.Filter(expr.Field(ProductPaths.Ratings.Path), expr.Gte("$$rating.score", 5), expr.FilterAs("rating"))),
			stage.Field("total", expr.Reduce(expr.Field(ProductPaths.Ratings.Path), 0, expr.Add("$$value", "$$this.score"))),
		),
		stage.Limit(3),
	)

	return showAggregate(ctx, coll, "stage.Project(stage.Field(\"scores\", expr.Map(...)), stage.Field(\"top\", expr.Filter(...)))", arrays)
}

// aggBuckets shows the stages that put documents into ranges, and the one that runs several pipelines at once.
func aggBuckets(ctx context.Context, coll *mongo.Collection) error {
	step("$bucket needs its boundaries and a default for whatever falls outside them")

	buckets := stage.Pipeline(
		stage.Bucket(
			expr.Field(ProductPaths.Price), []any{0, 50, 250, 1500},
			stage.BucketDefault("other"),
			stage.BucketOutput(
				stage.Accumulator("count", expr.Sum(1)),
				stage.Accumulator("skus", expr.Push(expr.Field(ProductPaths.SKU))),
			),
		),
	)
	if err := showAggregate(ctx, coll, "stage.Bucket(expr.Field(price), boundaries, stage.BucketDefault(...), stage.BucketOutput(...))", buckets); err != nil {
		return err
	}

	step("$facet runs several pipelines over the same input, each one named")

	facets := stage.Pipeline(
		stage.Facet(
			stage.FacetPipeline(
				"byCategory",
				stage.SortByCount(expr.Field(ProductPaths.Category)),
			),
			stage.FacetPipeline(
				"priceStats",
				stage.Group(
					nil,
					stage.Accumulator("min", expr.Min(expr.Field(ProductPaths.Price))),
					stage.Accumulator("max", expr.Max(expr.Field(ProductPaths.Price))),
				),
			),
			stage.FacetPipeline(
				"cheapest",
				stage.Sort(monq.Sort(monq.Asc(ProductPaths.Price))),
				stage.Limit(2),
				stage.Project(stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field(ProductPaths.Price, 1)),
			),
		),
	)

	return showAggregate(ctx, coll, "stage.Facet(stage.FacetPipeline(\"byCategory\", ...), ...)", facets)
}

// aggWindows shows $setWindowFields, where the output fields are window operators over a partition.
func aggWindows(ctx context.Context, coll *mongo.Collection) error {
	step("$setWindowFields ranks and accumulates within a partition, without collapsing the documents")

	stages := stage.Pipeline(
		stage.SetWindowFields(
			expr.Field(ProductPaths.Category), monq.Sort(monq.Desc(ProductPaths.Price)),
			stage.WindowField("rank", expr.Rank()),
			stage.WindowField(
				"runningStock", expr.Sum(expr.Field(ProductPaths.Stock)),
				stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent),
			),
		),
		stage.Project(
			stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field(ProductPaths.Category, 1),
			stage.Field("rank", 1), stage.Field("runningStock", 1),
		),
		stage.Sort(monq.Sort(monq.Asc(ProductPaths.Category), monq.Asc("rank"))),
		stage.Limit(6),
	)

	return showAggregate(ctx, coll, "stage.SetWindowFields(partition, sort, stage.WindowField(\"rank\", expr.Rank()), ...)", stages)
}

// aggUnwind runs over the orders collection, where the documents have an array worth taking apart.
func aggUnwind(ctx context.Context, coll *mongo.Collection) error {
	step("$unwind turns one document with an array into one document per element")

	stages := stage.Pipeline(
		stage.Match(monq.Eq(OrderPaths.Status, "shipped")),
		stage.Unwind(expr.Field(OrderPaths.Items.Path), stage.IncludeArrayIndex("line")),
		stage.Project(
			stage.Field(OrderPaths.ID, 0), stage.Field(OrderPaths.Customer, 1), stage.Field(OrderPaths.Items.SKU, 1),
			stage.Field(OrderPaths.Items.Quantity, 1), stage.Field("line", 1),
		),
	)
	if err := showAggregate(ctx, coll, "stage.Unwind(expr.Field(items), stage.IncludeArrayIndex(\"line\"))", stages); err != nil {
		return err
	}

	step("unwound elements group back up the usual way")

	perSKU := stage.Pipeline(
		stage.Unwind(expr.Field(OrderPaths.Items.Path)),
		stage.Group(
			expr.Field(OrderPaths.Items.SKU),
			stage.Accumulator("sold", expr.Sum(expr.Field(OrderPaths.Items.Quantity))),
			stage.Accumulator("revenue", expr.Sum(expr.Multiply(expr.Field(OrderPaths.Items.Quantity), expr.Field(OrderPaths.Items.Price)))),
		),
		stage.Sort(monq.Sort(monq.Desc("sold"))),
	)

	return showAggregate(ctx, coll, "stage.Group(expr.Field(items.sku), stage.Accumulator(\"sold\", expr.Sum(...)))", perSKU)
}

// aggLookup shows the two join stages: the field equality form and the pipeline form.
func aggLookup(ctx context.Context, coll *mongo.Collection) error {
	step("$lookup joins on one field each side and puts the matches in an array")

	stages := stage.Pipeline(
		stage.Unwind(expr.Field(OrderPaths.Items.Path)),
		stage.Lookup(productsColl, OrderPaths.Items.SKU, ProductPaths.SKU, "product"),
		stage.Unwind(expr.Field("product")),
		stage.Project(
			stage.Field(OrderPaths.ID, 0),
			stage.Field(OrderPaths.Customer, 1),
			stage.Field("sku", expr.Field(OrderPaths.Items.SKU)),
			stage.Field("name", expr.Field("product.name")),
		),
		stage.Limit(4),
	)
	if err := showAggregate(ctx, coll, `stage.Lookup("products", OrderPaths.Items.SKU, ProductPaths.SKU, "product")`, stages); err != nil {
		return err
	}

	step("the pipeline form binds variables with let and runs a pipeline of its own against the other collection")

	joined := stage.Pipeline(
		stage.Match(monq.Eq(OrderPaths.Customer, "grace")),
		stage.LookupPipeline(
			productsColl,
			bson.D{{Key: "spent", Value: expr.Field(OrderPaths.Total)}},
			stage.Pipeline(
				stage.Match(monq.Expr(expr.Gt(expr.Field(ProductPaths.Price), "$$spent"))),
				stage.Project(stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field(ProductPaths.Price, 1)),
			),
			"pricierThanThisOrder",
		),
		stage.Project(stage.Field(OrderPaths.ID, 0), stage.Field(OrderPaths.Total, 1), stage.Field("pricierThanThisOrder", 1)),
	)

	return showAggregate(ctx, coll, "stage.LookupPipeline(products, let, pipeline, \"pricierThanThisOrder\")", joined)
}

// aggReshape shows the stages that change a document's shape rather than which documents there are, plus the
// small counting and paging stages that need no explanation of their own.
func aggReshape(ctx context.Context, coll *mongo.Collection) error {
	step("$replaceWith promotes an expression to be the whole document, and $unset drops fields by path")

	stages := stage.Pipeline(
		stage.Match(monq.Eq(ProductPaths.Category, "peripherals")),
		stage.ReplaceWith(
			expr.MergeObjects(
				bson.D{{Key: "label", Value: expr.Concat(expr.Field(ProductPaths.SKU), " ", expr.Field(ProductPaths.Name))}},
				expr.Field(warehousePath),
			),
		),
		stage.Unset("type"),
	)
	if err := showAggregate(ctx, coll, "stage.ReplaceWith(expr.MergeObjects(...)), stage.Unset(\"type\")", stages); err != nil {
		return err
	}

	step("$replaceRoot is the same stage under its older name, and takes the document to promote as newRoot")

	rooted := stage.Pipeline(
		stage.Match(monq.Exists(ProductPaths.Ratings.Path, true)),
		stage.ReplaceRoot(expr.ArrayElemAt(expr.Field(ProductPaths.Ratings.Path), 0)),
		stage.Limit(3),
	)
	if err := showAggregate(ctx, coll, "stage.ReplaceRoot(expr.ArrayElemAt(expr.Field(ratings), 0))", rooted); err != nil {
		return err
	}

	step("$redact walks the document and keeps, drops, or descends into each level, deciding with an expression")

	redacted := stage.Pipeline(
		stage.Match(bySKU("mon-003")),
		stage.Redact(expr.Cond(expr.Gte(expr.IfNull(expr.Field("score"), 5), 5), "$$DESCEND", "$$PRUNE")),
		stage.Project(stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field(ProductPaths.Ratings.Path, 1)),
	)
	if err := showAggregate(ctx, coll, `stage.Redact(expr.Cond(..., "$$DESCEND", "$$PRUNE"))`, redacted); err != nil {
		return err
	}

	return aggCounting(ctx, coll)
}

// aggCounting shows the stages that count, page, and sample, and the one that reads a second collection in.
func aggCounting(ctx context.Context, coll *mongo.Collection) error {
	step("$skip, $limit, and $count are one argument each, and $count names the field it writes")

	stages := stage.Pipeline(
		stage.Sort(monq.Sort(monq.Asc(ProductPaths.SKU))),
		stage.Skip(2),
		stage.Limit(3),
		stage.Count("shown"),
	)
	if err := showAggregate(ctx, coll, `stage.Skip(2), stage.Limit(3), stage.Count("shown")`, stages); err != nil {
		return err
	}

	step("$sample picks documents at random, so this one answers differently every run")

	sampled := stage.Pipeline(
		stage.Sample(2),
		stage.Project(stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1)),
	)
	if err := showAggregate(ctx, coll, "stage.Sample(2)", sampled); err != nil {
		return err
	}

	step("$unionWith concatenates another collection's documents, optionally through a pipeline of their own")

	both := stage.Pipeline(
		stage.Match(monq.Eq(ProductPaths.Category, "furniture")),
		stage.Project(stage.Field(ProductPaths.ID, 0), stage.Field("label", expr.Field(ProductPaths.SKU))),
		stage.UnionWith(
			ordersColl,
			stage.Match(monq.Eq(OrderPaths.Status, "pending")),
			stage.Project(stage.Field(OrderPaths.ID, 0), stage.Field("label", expr.Field(OrderPaths.Customer))),
		),
	)

	return showAggregate(ctx, coll, "stage.UnionWith(ordersColl, stage.Match(...), stage.Project(...))", both)
}

// aggOutput shows the two stages that write their results back into the database, which have to come last.
func aggOutput(ctx context.Context, coll *mongo.Collection) error {
	step("$merge writes into a collection and returns nothing, deciding per document what to do about one already there")

	merged := stage.Pipeline(
		stage.Group(
			expr.Field(ProductPaths.Category),
			stage.Accumulator("stock", expr.Sum(expr.Field(ProductPaths.Stock))),
		),
		stage.Merge(
			stage.Namespace(coll.Database().Name(), "category_stock"),
			stage.MergeOn("_id"),
			stage.MergeWhenMatched("replace"),
			stage.MergeWhenNotMatched("insert"),
		),
	)
	if err := showAggregate(ctx, coll, `stage.Merge(stage.Namespace(db, "category_stock"), stage.MergeOn("_id"), ...)`, merged); err != nil {
		return err
	}

	written, err := coll.Database().Collection("category_stock").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count merged: %w", err)
	}

	fmt.Printf("   -> $merge wrote %d documents into category_stock\n", written)

	step("$out replaces a collection outright, which is the difference between the two, and returns nothing either")

	out := stage.Pipeline(
		stage.Match(monq.Gt(ProductPaths.Price, 100)),
		stage.Project(stage.Field(ProductPaths.SKU, 1), stage.Field(ProductPaths.Price, 1)),
		stage.Out("expensive"),
	)
	if err = showAggregate(ctx, coll, `stage.Out("expensive")`, out); err != nil {
		return err
	}

	replaced, err := coll.Database().Collection("expensive").CountDocuments(ctx, bson.D{})
	if err != nil {
		return fmt.Errorf("count out: %w", err)
	}

	fmt.Printf("   -> $out wrote %d documents into expensive\n", replaced)

	return nil
}

// aggSeries shows the stages for a series with gaps in it, which come as a pair: one adds the missing documents,
// the other gives them values.
func aggSeries(ctx context.Context, coll *mongo.Collection) error {
	step("$densify adds a document wherever the series skips a step, carrying only the field it stepped on")

	stages := stage.Pipeline(
		stage.Project(stage.Field(OrderPaths.ID, 0), stage.Field(OrderPaths.PlacedAt, 1), stage.Field(OrderPaths.Total, 1)),
		stage.Densify(OrderPaths.PlacedAt, stage.DensifyRange(1, stage.DensifyFull, "month")),
		stage.Sort(monq.Sort(monq.Asc(OrderPaths.PlacedAt))),
	)
	if err := showAggregate(ctx, coll, "stage.Densify(placed_at, stage.DensifyRange(1, stage.DensifyFull, \"month\"))", stages); err != nil {
		return err
	}

	step("$fill then gives those documents values, either a fixed one or one derived from their neighbors")

	filled := stage.Pipeline(
		stage.Project(
			stage.Field(OrderPaths.ID, 0), stage.Field(OrderPaths.PlacedAt, 1), stage.Field(OrderPaths.Total, 1),
			stage.Field(OrderPaths.Status, 1),
		),
		stage.Densify(OrderPaths.PlacedAt, stage.DensifyRange(1, stage.DensifyFull, "month")),
		stage.Fill(
			[]bson.D{
				stage.FillValue(OrderPaths.Status, "none"),
				stage.FillMethod(OrderPaths.Total, "locf"),
			}, stage.FillSortBy(monq.Sort(monq.Asc(OrderPaths.PlacedAt))),
		),
		stage.Sort(monq.Sort(monq.Asc(OrderPaths.PlacedAt))),
	)

	return showAggregate(ctx, coll, `stage.Fill([]bson.D{stage.FillValue(status, "none"), stage.FillMethod(total, "locf")}, ...)`, filled)
}

// aggGraph shows $graphLookup, which follows a reference from one document to another in the same collection
// until it runs out of links.
func aggGraph(ctx context.Context, coll *mongo.Collection) error {
	step("$graphLookup walks a hierarchy, and DepthField says how many hops each match took")

	stages := stage.Pipeline(
		stage.Match(monq.Eq("name", "accessories")),
		stage.GraphLookup(
			categoriesColl, expr.Field("parent"), "parent", "name", "ancestors",
			stage.DepthField("depth"),
			stage.MaxDepth(5),
		),
		stage.Project(stage.Field("_id", 0), stage.Field("name", 1), stage.Field("ancestors.name", 1), stage.Field("ancestors.depth", 1)),
	)

	return showAggregate(ctx, coll, `stage.GraphLookup(categoriesColl, expr.Field("parent"), "parent", "name", "ancestors", ...)`, stages)
}

// aggDocuments shows the one stage that brings its own input, which is why it runs against the database rather
// than a collection.
func aggDocuments(ctx context.Context, db *mongo.Database) error {
	step("$documents makes a pipeline out of literal documents, which is how an expression is tried out")

	stages := stage.Pipeline(
		stage.Documents(
			bson.D{{Key: "name", Value: "  ada  "}, {Key: "scores", Value: bson.A{5, 3, 4}}},
			bson.D{{Key: "name", Value: "grace"}, {Key: "scores", Value: bson.A{2}}},
		),
		stage.Set(
			stage.Field("name", expr.Trim(expr.Field("name"))),
			stage.Field("best", expr.Max(expr.Field("scores"))),
		),
	)

	return showDBAggregate(ctx, db, "stage.Documents(bson.D{...}, bson.D{...}), stage.Set(...)", stages)
}
