package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
	"github.com/behzadsh/monq/stage"
)

// runPaths covers monqgen, the optional code generator. Field paths are strings, so a renamed field turns into a
// query that quietly matches nothing; the generated constants make that a compile error instead. Everything else
// in this program already uses them, so this demo is about what they are rather than what they do.
func runPaths(ctx context.Context, db *mongo.Database) error {
	return runParts(ctx, db.Collection(productsColl),
		pathConstants,
		pathPositions,
		pathsInQueries,
	)
}

// pathConstants shows what monqgen wrote for the structs in model.go.
func pathConstants(_ context.Context, _ *mongo.Collection) error {
	step("the go:generate line in model.go writes product_paths.go, and the constants are monq.FieldPath values")

	fmt.Printf("   ProductPaths.SKU                 = %q\n", ProductPaths.SKU)
	fmt.Printf("   ProductPaths.Warehouse.Type      = %q   // a nested document nests its paths too\n", ProductPaths.Warehouse.Type)
	fmt.Printf("   ProductPaths.Ratings.Path        = %q   // the array itself, for $size and $all\n", ProductPaths.Ratings.Path)
	fmt.Printf("   ProductPaths.Ratings.Score       = %q   // the dotted form, matching any element\n", ProductPaths.Ratings.Score)
	fmt.Printf("   OrderPaths.Items.Quantity        = %q   // the bson tag wins over the Go field name\n", OrderPaths.Items.Quantity)

	fmt.Println("\n   The same tree without generating anything:")
	fmt.Println("   $ go run github.com/behzadsh/monq/cmd/monqgen -type Product -print ./examples")

	return nil
}

// pathPositions shows the four ways MongoDB names a position inside an array, which the generated array types and
// monq.ArrayPath both spell out.
func pathPositions(_ context.Context, _ *mongo.Collection) error {
	step("an array carries the four positional forms, so the punctuation never has to be remembered")

	fmt.Printf("   ProductPaths.Ratings.At(2).Score          = %q\n", ProductPaths.Ratings.At(2).Score)
	fmt.Printf("   ProductPaths.Ratings.Positional().Score   = %q\n", ProductPaths.Ratings.Positional().Score)
	fmt.Printf("   ProductPaths.Ratings.All().Score          = %q\n", ProductPaths.Ratings.All().Score)
	fmt.Printf("   ProductPaths.Ratings.Filtered(\"low\").Score = %q\n", ProductPaths.Ratings.Filtered("low").Score)
	fmt.Printf("   ProductPaths.Tags.At(0)                   = %q   // an array of scalars has no fields under it\n", ProductPaths.Tags.At(0))

	step("without codegen the same forms come from monq.ArrayPath, and hand-written paths work exactly as well")

	items := monq.ArrayPath{Path: "items"}
	fmt.Printf("   monq.ArrayPath{Path: \"items\"}.Filtered(\"cheap\") = %q\n", items.Filtered("cheap"))
	fmt.Printf("   monq.Size(items.Path, 3)                      = %s\n", jsonOf(monq.Size(items.Path, 3)))

	return nil
}

// pathsInQueries shows the constants going into the same operators a string literal would, in all three packages.
func pathsInQueries(ctx context.Context, coll *mongo.Collection) error {
	step("a generated path goes anywhere a path belongs, and fails to compile where one does not")

	filter := monq.And(
		monq.Eq(ProductPaths.Category, "accessories"),
		monq.Gte(ProductPaths.Ratings.Score, 1),
	)
	if err := showFind(ctx, coll, "monq.And(monq.Eq(ProductPaths.Category, ...), monq.Gte(ProductPaths.Ratings.Score, 1))", filter,
		options.Find().SetProjection(monq.Projection(monq.Include(ProductPaths.SKU, ProductPaths.Ratings.Path), monq.Exclude(ProductPaths.ID)))); err != nil {
		return err
	}

	step("stage and expr take them too, since a path is a path whichever package is asking")

	stages := stage.Pipeline(
		stage.Match(monq.Exists(ProductPaths.Ratings.Path, true)),
		stage.Group(expr.Field(ProductPaths.Category),
			stage.Accumulator("ratings", expr.Sum(expr.Size(expr.Field(ProductPaths.Ratings.Path)))),
		),
		stage.Sort(monq.Sort(monq.Desc("ratings"))),
		stage.Limit(3),
	)
	if err := showAggregate(ctx, coll, "expr.Field(ProductPaths.Category), expr.Size(expr.Field(ProductPaths.Ratings.Path))", stages); err != nil {
		return err
	}

	fmt.Println("\n   Rename the bson tag on Product.Ratings, rerun go generate, and every line above stops compiling,")
	fmt.Println("   which is the whole point: a string literal would have gone on matching nothing.")

	return nil
}
