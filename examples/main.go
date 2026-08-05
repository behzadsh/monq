// Command examples runs monq against a live MongoDB deployment, one demo per area of the API.
//
// Every demo prints the bson.D that monq built and then the documents MongoDB returned for it, so the same page
// shows the call, the query it produces, and what that query does to real data.
//
// It needs a MongoDB to talk to and nothing else. A throwaway one is a container away:
//
//	docker run --rm -p 27017:27017 mongo:8
//	go run ./examples
//
// The database is dropped and reseeded on every run, so nothing here depends on the state a previous run left
// behind. Name one or more demos to run only those:
//
//	go run ./examples find update
//	go run ./examples -uri mongodb://user:pass@host:27017 -db scratch aggregate
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// demo is one runnable topic, named by the argument that selects it.
type demo struct {
	name string
	what string
	run  func(ctx context.Context, db *mongo.Database) error
}

// demos are run in this order when no argument picks a subset.
var demos = []demo{
	{"find", "query operators and the Find calls they go into", runFind},
	{"project", "projections and sort documents", runProject},
	{"update", "update operators, upserts, and array updates", runUpdate},
	{"aggregate", "pipeline stages and the expressions they compute with", runAggregate},
	{"compute", "date, string, conversion, set, object, and window expressions", runCompute},
	{"geo", "geospatial queries and the geometry they take", runGeo},
	{"index", "index models, including text and TTL indexes", runIndex},
	{"paths", "generated field path constants", runPaths},
}

const (
	// Where a container started with the command in this file's doc comment listens.
	defaultURI = "mongodb://localhost:27017"
	// Created on first write, and dropped at the start of every demo.
	defaultDB = "monq_examples"
	// Bounds the whole run, seeding included.
	timeout = 2 * time.Minute
)

func main() {
	uri := flag.String("uri", envOr("MONGODB_URI", defaultURI), "MongoDB connection string")
	database := flag.String("db", defaultDB, "database to seed and query")

	flag.Usage = usage
	flag.Parse()

	if err := run(*uri, *database, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "examples: %v\n", err)
		os.Exit(1)
	}
}

// run connects, seeds the database, and runs the selected demos against it.
func run(uri, database string, names []string) error {
	selected, err := selectDemos(names)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongo.Connect(options.Client().ApplyURI(uri))
	if err != nil {
		return fmt.Errorf("connect to %s: %w", uri, err)
	}

	defer disconnect(client)

	if err = client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("ping %s (start one with: docker run --rm -p 27017:27017 mongo:8): %w", uri, err)
	}

	db := client.Database(database)

	for _, d := range selected {
		// Reseeding before each demo keeps them independent, since the update demo writes to the same documents
		// the others read.
		if err = seed(ctx, db); err != nil {
			return fmt.Errorf("seed: %w", err)
		}

		section(d.name, d.what)

		if err = d.run(ctx, db); err != nil {
			return fmt.Errorf("%s: %w", d.name, err)
		}
	}

	fmt.Printf("\nDone. The seeded data is still in the %q database if you want to poke at it.\n", database)

	return nil
}

// part is one section of a demo, run against the collection that demo is about.
type part func(ctx context.Context, coll *mongo.Collection) error

// runParts runs the sections of a demo in order, stopping at the first error.
func runParts(ctx context.Context, coll *mongo.Collection, parts ...part) error {
	for _, p := range parts {
		if err := p(ctx, coll); err != nil {
			return err
		}
	}

	return nil
}

// selectDemos resolves command line arguments to the demos they name, defaulting to all of them in order.
func selectDemos(names []string) ([]demo, error) {
	if len(names) == 0 {
		return demos, nil
	}

	selected := make([]demo, 0, len(names))

	for _, name := range names {
		i := indexOfDemo(name)
		if i < 0 {
			return nil, fmt.Errorf("no demo named %q, pick from: %s", name, strings.Join(demoNames(), " "))
		}

		selected = append(selected, demos[i])
	}

	return selected, nil
}

// indexOfDemo returns the position of the demo called name, or -1 when there is none.
func indexOfDemo(name string) int {
	for i, d := range demos {
		if d.name == name {
			return i
		}
	}

	return -1
}

// demoNames lists the demo names in the order they run.
func demoNames() []string {
	names := make([]string, len(demos))
	for i, d := range demos {
		names[i] = d.name
	}

	return names
}

// disconnect closes the client with a context of its own, since the run's context may already be done.
func disconnect(client *mongo.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil && !errors.Is(err, mongo.ErrClientDisconnected) {
		fmt.Fprintf(os.Stderr, "examples: disconnect: %v\n", err)
	}
}

// envOr returns the environment variable named key, or fallback when it is unset or empty.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return fallback
}

// usage prints the flags and the demos that can be named as arguments.
func usage() {
	fmt.Fprint(os.Stderr, "usage: go run ./examples [flags] [demo...]\n\nflags:\n")
	flag.PrintDefaults()
	fmt.Fprint(os.Stderr, "\ndemos (all of them, in this order, when none is named):\n")

	for _, d := range demos {
		fmt.Fprintf(os.Stderr, "  %-10s %s\n", d.name, d.what)
	}
}
