package index_test

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// printKeys prints an index model's keys as relaxed extended JSON. Example tests use it instead of
// fmt.Println(keys), since bson.D's default String() renders the directions in verbose canonical extended JSON
// (e.g. {"$numberInt":"1"}), which is accurate but not what a reader expects to see in a godoc example.
//
// It mirrors printFilter in the root package's tests, printStage in the stage package's, and printExpr in the expr
// package's. The copies stay separate rather than becoming one shared helper, since exporting it would make a test
// convenience part of the library's public API. This one takes an any, because mongo.IndexModel's Keys field does.
func printKeys(keys any) {
	b, err := bson.MarshalExtJSON(keys, false, false)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}

// printOptions prints the properties an index model carries, one per line and only the ones that were set. Index
// options are a builder of setter functions rather than a document, so an example that demonstrates an option has
// nothing to marshal; this resolves the builder the way the driver does and reports the result.
func printOptions(builder *options.IndexOptionsBuilder) {
	var opts options.IndexOptions
	for _, set := range builder.List() {
		if err := set(&opts); err != nil {
			panic(err)
		}
	}

	line := func(name string, value any) {
		fmt.Printf("%s: %v\n", name, value)
	}

	if opts.Name != nil {
		line("name", *opts.Name)
	}

	if opts.Unique != nil {
		line("unique", *opts.Unique)
	}

	if opts.Sparse != nil {
		line("sparse", *opts.Sparse)
	}

	if opts.Hidden != nil {
		line("hidden", *opts.Hidden)
	}

	if opts.ExpireAfterSeconds != nil {
		line("expireAfterSeconds", *opts.ExpireAfterSeconds)
	}

	if opts.PartialFilterExpression != nil {
		b, err := bson.MarshalExtJSON(opts.PartialFilterExpression, false, false)
		if err != nil {
			panic(err)
		}

		line("partialFilterExpression", string(b))
	}
}
