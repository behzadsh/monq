package index_test

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
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
