package main

import (
	"context"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// The helpers here print what a demo built and what MongoDB answered, so every example shows both halves: the
// bson.D monq produced, and the documents the driver returned for it. Documents are rendered as relaxed extended
// JSON rather than with bson.D's own String method, which spells numbers out in full ({"$numberInt":"12"}).

// section prints the banner that starts one demo.
func section(name, what string) {
	fmt.Printf("\n%s\n== %s: %s\n%s\n", strings.Repeat("=", 100), name, what, strings.Repeat("=", 100))
}

// step prints the one-line description of a single example inside a demo.
func step(format string, args ...any) {
	fmt.Printf("\n-- "+format+"\n", args...)
}

// jsonOf renders a document as relaxed extended JSON.
func jsonOf(d any) string {
	b, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		return fmt.Sprintf("<unmarshalable: %v>", err)
	}

	return string(b)
}

// query prints a document monq built, labeled with the call that produced it.
func query(label string, d bson.D) {
	fmt.Printf("   %s\n   %s\n", label, jsonOf(d))
}

// pipeline prints an aggregation pipeline, one stage per line.
func pipeline(label string, stages []bson.D) {
	fmt.Printf("   %s\n", label)

	for _, s := range stages {
		fmt.Printf("   %s\n", jsonOf(s))
	}
}

// docs prints the documents a query returned, or says so when it returned none.
func docs(returned []bson.D) {
	if len(returned) == 0 {
		fmt.Println("   -> no documents")

		return
	}

	for _, d := range returned {
		fmt.Printf("   -> %s\n", jsonOf(d))
	}
}

// showFind runs a Find with the filter it just printed and prints the documents it returned. Cursor.All closes the
// cursor itself, which is why no example here defers a Close.
func showFind(ctx context.Context, coll *mongo.Collection, label string, filter bson.D, opts ...options.Lister[options.FindOptions]) error {
	query(label, filter)

	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return fmt.Errorf("find: %w", err)
	}

	var found []bson.D
	if err := cursor.All(ctx, &found); err != nil {
		return fmt.Errorf("read find results: %w", err)
	}

	docs(found)

	return nil
}

// showAggregate runs a pipeline it just printed and prints the documents it returned.
func showAggregate(ctx context.Context, coll *mongo.Collection, label string, stages []bson.D) error {
	pipeline(label, stages)

	cursor, err := coll.Aggregate(ctx, stages)
	if err != nil {
		return fmt.Errorf("aggregate: %w", err)
	}

	var found []bson.D
	if err := cursor.All(ctx, &found); err != nil {
		return fmt.Errorf("read aggregate results: %w", err)
	}

	docs(found)

	return nil
}

// showDBAggregate runs a pipeline against the database rather than a collection, which is what a pipeline
// starting with $documents needs, since it brings its own input.
func showDBAggregate(ctx context.Context, db *mongo.Database, label string, stages []bson.D) error {
	pipeline(label, stages)

	cursor, err := db.Aggregate(ctx, stages)
	if err != nil {
		return fmt.Errorf("aggregate: %w", err)
	}

	var found []bson.D
	if err := cursor.All(ctx, &found); err != nil {
		return fmt.Errorf("read aggregate results: %w", err)
	}

	docs(found)

	return nil
}

// showUpdate prints an update document, applies it to every matching document, and reports the counts the driver
// returned.
func showUpdate(
	ctx context.Context, coll *mongo.Collection, label string, filter, update bson.D,
	opts ...options.Lister[options.UpdateManyOptions],
) error {
	query(label, update)

	result, err := coll.UpdateMany(ctx, filter, update, opts...)
	if err != nil {
		return fmt.Errorf("update: %w", err)
	}

	fmt.Printf("   -> matched %d, modified %d, upserted %v\n", result.MatchedCount, result.ModifiedCount, result.UpsertedID)

	return nil
}

// showDoc prints the single document a filter matches, projected down to the fields the demo is about.
func showDoc(ctx context.Context, coll *mongo.Collection, label string, filter, projection bson.D) error {
	var found bson.D

	err := coll.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&found)
	if err != nil {
		return fmt.Errorf("find one: %w", err)
	}

	fmt.Printf("   %s\n   -> %s\n", label, jsonOf(found))

	return nil
}
