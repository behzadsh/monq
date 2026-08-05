package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/index"
)

// runIndex covers monq/index, the one package that imports the driver's mongo package, because mongo.IndexModel
// is what it returns. It is a package of its own so that a program building only queries never compiles the
// driver's mongo package or the dependencies it carries.
func runIndex(ctx context.Context, db *mongo.Database) error {
	if err := createIndexes(ctx, db.Collection(productsColl)); err != nil {
		return err
	}

	if err := listIndexes(ctx, db.Collection(productsColl)); err != nil {
		return err
	}

	if err := uniqueIndexInAction(ctx, db.Collection(productsColl)); err != nil {
		return err
	}

	return ttlIndex(ctx, db.Collection("sessions"))
}

// createIndexes builds several index models and creates them in one call.
func createIndexes(ctx context.Context, coll *mongo.Collection) error {
	step("keys and options are one argument list, so an index reads as a single expression")

	models := []mongo.IndexModel{
		index.New(index.Asc(ProductPaths.SKU), index.Unique(), index.Name("sku_unique")),
		index.New(index.Asc(ProductPaths.Category), index.Desc(ProductPaths.Price), index.Name("category_price")),
		index.New(index.Asc(ProductPaths.Stock), index.PartialFilter(monq.Gt(ProductPaths.Stock, 0)), index.Name("stock_in_stock")),
		index.New(index.Hashed(ProductPaths.SKU), index.Name("sku_hashed"), index.Hidden()),
	}

	fmt.Println("   index.New(index.Asc(sku), index.Unique(), index.Name(\"sku_unique\"))")
	fmt.Println("   index.New(index.Asc(category), index.Desc(price), index.Name(\"category_price\"))")
	fmt.Println("   index.New(index.Asc(stock), index.PartialFilter(monq.Gt(stock, 0)), index.Name(\"stock_in_stock\"))")
	fmt.Println("   index.New(index.Hashed(sku), index.Name(\"sku_hashed\"), index.Hidden())")

	names, err := coll.Indexes().CreateMany(ctx, models)
	if err != nil {
		return fmt.Errorf("create indexes: %w", err)
	}

	fmt.Printf("   -> created: %v\n", names)

	return nil
}

// listIndexes reads back what the collection now has, including the two the seed created.
func listIndexes(ctx context.Context, coll *mongo.Collection) error {
	step("the partial filter and the text index the seed created are visible in the collection's own listing")

	cursor, err := coll.Indexes().List(ctx)
	if err != nil {
		return fmt.Errorf("list indexes: %w", err)
	}

	var specs []bson.D
	if err = cursor.All(ctx, &specs); err != nil {
		return fmt.Errorf("read index list: %w", err)
	}

	docs(specs)

	return nil
}

// uniqueIndexInAction shows the index doing its job, since an index is only interesting through the writes it
// rejects and the queries it speeds up.
func uniqueIndexInAction(ctx context.Context, coll *mongo.Collection) error {
	step("a unique index rejects the second document with the same key, which the driver reports as a write error")

	duplicate := bson.D{{Key: string(ProductPaths.SKU), Value: "cbl-004"}, {Key: string(ProductPaths.Name), Value: "Duplicate Cable"}}

	_, err := coll.InsertOne(ctx, duplicate)
	switch {
	case mongo.IsDuplicateKeyError(err):
		fmt.Println("   -> InsertOne(sku: cbl-004) rejected by sku_unique, as it should be")
	case err != nil:
		return fmt.Errorf("insert duplicate: %w", err)
	default:
		return errors.New("insert of a duplicate sku succeeded, so the unique index is missing")
	}

	step("and the compound index turns a matching query into an index scan, which explain reports as IXSCAN")

	filter := monq.And(monq.Eq(ProductPaths.Category, "accessories"), monq.Gt(ProductPaths.Price, 10))
	query("monq.And(monq.Eq(category, ...), monq.Gt(price, 10))", filter)

	return explainFind(ctx, coll, filter)
}

// explainFind asks the server how it would run the filter, which is the only way to see an index being used
// rather than assumed. The explain command takes the same filter document the query would.
func explainFind(ctx context.Context, coll *mongo.Collection, filter bson.D) error {
	command := bson.D{
		{Key: "explain", Value: bson.D{{Key: "find", Value: coll.Name()}, {Key: "filter", Value: filter}}},
		{Key: "verbosity", Value: "queryPlanner"},
	}

	explained, err := coll.Database().RunCommand(ctx, command).Raw()
	if err != nil {
		return fmt.Errorf("explain: %w", err)
	}

	plan, lookupErr := explained.LookupErr("queryPlanner", "winningPlan")
	if lookupErr != nil {
		return fmt.Errorf("read winning plan: %w", lookupErr)
	}

	fmt.Printf("   -> winning plan: %s\n", jsonOf(plan.Document()))

	return nil
}

// ttlIndex shows the one index option that changes what the collection does rather than how it is searched, on a
// collection of its own so nothing else here expires halfway through.
func ttlIndex(ctx context.Context, coll *mongo.Collection) error {
	step("a TTL index takes a Go duration, and MongoDB deletes documents once their date field is that old")

	model := index.New(index.Asc("created_at"), index.TTL(30*time.Minute), index.Name("sessions_ttl"))

	name, err := coll.Indexes().CreateOne(ctx, model)
	if err != nil {
		return fmt.Errorf("create ttl index: %w", err)
	}

	fmt.Printf("   index.New(index.Asc(\"created_at\"), index.TTL(30*time.Minute), index.Name(\"sessions_ttl\"))\n   -> created: %s\n", name)

	if _, err = coll.InsertOne(ctx, bson.D{{Key: "created_at", Value: time.Now()}, {Key: "token", Value: "abc"}}); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}

	fmt.Println("   -> one session inserted, which MongoDB will remove about half an hour from now")

	return nil
}
