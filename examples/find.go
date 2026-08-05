package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
)

// runFind covers the query operators: what each one builds, and how the result goes into Find, FindOne, or any
// other driver call that takes a filter. Nothing here wraps the driver, the filter is a bson.D and Find takes it
// as it is.
func runFind(ctx context.Context, db *mongo.Database) error {
	return runParts(ctx, db.Collection(productsColl),
		findComparison,
		findLogical,
		findElement,
		findEvaluation,
		findArray,
		findBitwise,
		findRawAndCounts,
	)
}

// brief trims the printed results down to a few fields in a fixed order, so the demo output stays about the
// filter rather than the fixture data. Both the projection and the sort are monq documents.
func brief() *options.FindOptionsBuilder {
	return options.Find().
		SetProjection(monq.Projection(monq.Include(ProductPaths.SKU, ProductPaths.Price, ProductPaths.Stock), monq.Exclude(ProductPaths.ID))).
		SetSort(monq.Sort(monq.Asc(ProductPaths.SKU)))
}

// findComparison shows the operators that compare a field against a value.
func findComparison(ctx context.Context, coll *mongo.Collection) error {
	step("comparison: every operator emits the explicit {field: {$op: value}} form, $eq included")

	if err := showFind(ctx, coll, `monq.Eq(ProductPaths.SKU, "kbd-001")`, bySKU("kbd-001"), brief()); err != nil {
		return err
	}

	step("a range is two operators on one field, which is what And is for")

	between := monq.And(
		monq.Gte(ProductPaths.Price, 20),
		monq.Lt(ProductPaths.Price, 100),
	)
	if err := showFind(ctx, coll, "monq.And(monq.Gte(price, 20), monq.Lt(price, 100))", between, brief()); err != nil {
		return err
	}

	step("$in is variadic: In(field, values) with an un-spread slice matches the slice itself, so spread it")

	categories := []any{"peripherals", "displays"}
	if err := showFind(ctx, coll, `monq.In(ProductPaths.Category, categories...)`, monq.In(ProductPaths.Category, categories...), brief()); err != nil {
		return err
	}

	step("$ne and $nin also match documents missing the field, which is a MongoDB rule rather than a monq one")

	if err := showFind(ctx, coll, `monq.Ne(ProductPaths.Category, "accessories")`, monq.Ne(ProductPaths.Category, "accessories"), brief()); err != nil {
		return err
	}

	excluded := monq.Nin(ProductPaths.Category, "accessories", "peripherals")
	if err := showFind(ctx, coll, `monq.Nin(ProductPaths.Category, "accessories", "peripherals")`, excluded, brief()); err != nil {
		return err
	}

	step("$lte and $gt bound the other two corners of a range")

	return showFind(ctx, coll, "monq.Lte(ProductPaths.Stock, 5)", monq.Lte(ProductPaths.Stock, 5), brief())
}

// findLogical shows how filters nest, and what Not does with the explicit operator form.
func findLogical(ctx context.Context, coll *mongo.Collection) error {
	step("filters compose by nesting: no builder object, no method chain")

	filter := monq.And(
		monq.Eq(ProductPaths.Category, "peripherals"),
		monq.Or(
			monq.Gt(ProductPaths.Price, 80),
			monq.Eq(ProductPaths.Stock, 0),
		),
	)
	if err := showFind(ctx, coll, "monq.And(monq.Eq(category, ...), monq.Or(monq.Gt(price, 80), monq.Eq(stock, 0)))", filter, brief()); err != nil {
		return err
	}

	step("$nor matches documents failing every condition")

	nor := monq.Nor(
		monq.Lt(ProductPaths.Price, 100),
		monq.Gt(ProductPaths.Stock, 5),
	)
	if err := showFind(ctx, coll, "monq.Nor(monq.Lt(price, 100), monq.Gt(stock, 5))", nor, brief()); err != nil {
		return err
	}

	step("Not takes a built filter and turns its operator inside out, which only works because Eq emits {$eq: v}")

	return showFind(ctx, coll, "monq.Not(monq.Gte(price, 60))", monq.Not(monq.Gte(ProductPaths.Price, 60)), brief())
}

// findElement shows the operators that ask about a field's presence and BSON type.
func findElement(ctx context.Context, coll *mongo.Collection) error {
	step("$exists asks whether the field is there at all: two products were seeded with no ratings")

	if err := showFind(ctx, coll, "monq.Exists(ProductPaths.Ratings.Path, false)", monq.Exists(ProductPaths.Ratings.Path, false), brief()); err != nil {
		return err
	}

	step("$type takes either the type names or their numbers, and several of them at once")

	return showFind(ctx, coll, `monq.Type(ProductPaths.Price, "double")`, monq.Type(ProductPaths.Price, "double"), brief())
}

// findEvaluation shows the operators that compute something: regexes, arithmetic, aggregation expressions, and
// full text search.
func findEvaluation(ctx context.Context, coll *mongo.Collection) error {
	step("$regex takes the pattern and its options as two strings, not a compiled regexp")

	if err := showFind(ctx, coll, `monq.Regex(ProductPaths.Name, "^wireless", "i")`, monq.Regex(ProductPaths.Name, "^wireless", "i"), brief()); err != nil {
		return err
	}

	step("$mod picks documents whose field divides a given way: an even stock count here")

	if err := showFind(ctx, coll, "monq.Mod(ProductPaths.Stock, 2, 0)", monq.Mod(ProductPaths.Stock, 2, 0), brief()); err != nil {
		return err
	}

	step("$expr compares two fields of the same document, which a plain filter cannot do")

	price := expr.Field(ProductPaths.Price)
	stock := expr.Field(ProductPaths.Stock)

	if err := showFind(ctx, coll, "monq.Expr(expr.Gt(expr.Field(price), expr.Multiply(expr.Field(stock), 10)))",
		monq.Expr(expr.Gt(price, expr.Multiply(stock, 10))), brief()); err != nil {
		return err
	}

	step("$text searches the text index the seed created, and its options are typed to the operator")

	if err := showFind(ctx, coll, `monq.Text("wireless", monq.Language("english"))`, monq.Text("wireless", monq.Language("english")), brief()); err != nil {
		return err
	}

	cased := monq.Text("Wireless", monq.CaseSensitive(), monq.DiacriticSensitive())
	if err := showFind(ctx, coll, "monq.Text(\"Wireless\", monq.CaseSensitive(), monq.DiacriticSensitive())", cased, brief()); err != nil {
		return err
	}

	return findSchemaAndSample(ctx, coll)
}

// findSchemaAndSample shows the two evaluation operators that judge a document as a whole rather than a field.
func findSchemaAndSample(ctx context.Context, coll *mongo.Collection) error {
	step("$jsonSchema validates the whole document, and the schema goes in as a bson.D since it is JSON Schema, not monq")

	schema := monq.JSONSchema(bson.D{
		{Key: "required", Value: bson.A{string(ProductPaths.SKU), string(ProductPaths.Ratings.Path)}},
		{Key: "properties", Value: bson.D{
			{Key: string(ProductPaths.Price), Value: bson.D{{Key: "bsonType", Value: "double"}, {Key: "maximum", Value: 100}}},
		}},
	})
	if err := showFind(ctx, coll, "monq.JSONSchema(bson.D{...})", schema, brief()); err != nil {
		return err
	}

	step("$sampleRate keeps roughly that fraction of the documents, so this one answers differently every run")

	return showFind(ctx, coll, "monq.SampleRate(0.5)", monq.SampleRate(0.5), brief())
}

// findArray shows the operators that ask about arrays, and the difference between conditions on one element and
// conditions spread across a whole array.
func findArray(ctx context.Context, coll *mongo.Collection) error {
	step("$all wants every value present, in any order")

	if err := showFind(ctx, coll, `monq.All(ProductPaths.Tags.Path, "usb-c", "powered")`,
		monq.All(ProductPaths.Tags.Path, "usb-c", "powered"), brief()); err != nil {
		return err
	}

	step("$size is an exact length, with no comparison form: $gt inside it is not a thing MongoDB has")

	if err := showFind(ctx, coll, "monq.Size(ProductPaths.Ratings.Path, 3)", monq.Size(ProductPaths.Ratings.Path, 3), brief()); err != nil {
		return err
	}

	step("a dotted path matches if any one element matches, so these two conditions can land on different ratings")

	loose := monq.And(
		monq.Eq(ProductPaths.Ratings.User, "ada"),
		monq.Eq(ProductPaths.Ratings.Score, 5),
	)
	if err := showFind(ctx, coll, `monq.And(monq.Eq("ratings.user", "ada"), monq.Eq("ratings.score", 5))`, loose, brief()); err != nil {
		return err
	}

	step("$elemMatch is how one element is made to satisfy all of them at once, and its paths are element relative")

	tight := monq.ElemMatch(ProductPaths.Ratings.Path,
		monq.Eq("user", "ada"),
		monq.Eq("score", 5),
	)

	return showFind(ctx, coll, `monq.ElemMatch("ratings", monq.Eq("user", "ada"), monq.Eq("score", 5))`, tight, brief())
}

// findBitwise shows the bitwise query operators, which read a number as a set of flags.
func findBitwise(ctx context.Context, coll *mongo.Collection) error {
	step("$bitsAllSet takes the mask as a number, a bit position list, or binary data")

	if err := showFind(ctx, coll, "monq.BitsAllSet(ProductPaths.Flags, 1)", monq.BitsAllSet(ProductPaths.Flags, 1), brief()); err != nil {
		return err
	}

	step("$bitsAnyClear is the same mask read the other way round")

	if err := showFind(ctx, coll, "monq.BitsAnyClear(ProductPaths.Flags, 8)", monq.BitsAnyClear(ProductPaths.Flags, 8), brief()); err != nil {
		return err
	}

	step("the other two spell out the remaining combinations, and a mask can be a list of bit positions")

	if err := showFind(
		ctx,
		coll,
		"monq.BitsAllClear(ProductPaths.Flags, []any{1, 2})",
		monq.BitsAllClear(ProductPaths.Flags, []any{1, 2}),
		brief(),
	); err != nil {
		return err
	}

	return showFind(ctx, coll, "monq.BitsAnySet(ProductPaths.Flags, 12)", monq.BitsAnySet(ProductPaths.Flags, 12), brief())
}

// findRawAndCounts shows the escape hatch for operators monq has no function for, and that a monq filter is just a
// filter: every driver call taking one accepts it.
func findRawAndCounts(ctx context.Context, coll *mongo.Collection) error {
	step("Raw drops a hand-written document into a monq filter, so adoption can be partial")

	mixed := monq.And(
		monq.Eq(ProductPaths.Category, "accessories"),
		monq.Raw(bson.D{{Key: string(ProductPaths.Stock), Value: bson.D{{Key: "$gt", Value: 100}}}}),
	)
	if err := showFind(ctx, coll, "monq.And(monq.Eq(category, ...), monq.Raw(bson.D{...}))", mixed, brief()); err != nil {
		return err
	}

	step("the same filter goes into CountDocuments, Distinct, DeleteMany, and the rest unchanged")

	inStock := monq.Gt(ProductPaths.Stock, 0)
	query("monq.Gt(ProductPaths.Stock, 0)", inStock)

	count, err := coll.CountDocuments(ctx, inStock)
	if err != nil {
		return fmt.Errorf("count: %w", err)
	}

	var categories []string
	if err = coll.Distinct(ctx, string(ProductPaths.Category), inStock).Decode(&categories); err != nil {
		return fmt.Errorf("distinct: %w", err)
	}

	deleted, err := coll.DeleteMany(ctx, monq.Eq(ProductPaths.Stock, 0))
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	fmt.Printf("   -> CountDocuments: %d\n   -> Distinct(category): %v\n   -> DeleteMany(stock $eq 0): %d removed\n", count, categories, deleted.DeletedCount)
	fmt.Println("   The next demo reseeds, so the deleted products come back.")

	return nil
}
