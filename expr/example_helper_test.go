package expr_test

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// printExpr prints d as relaxed extended JSON. Example tests use it instead of fmt.Println(d) whenever the
// expression contains a numeric value, since bson.D's default String() renders numbers in verbose canonical
// extended JSON (e.g. {"$numberInt":"18"}), which is accurate but not what a reader expects to see in a godoc
// example.
//
// It mirrors printFilter in the root package's tests and printStage in the stage package's. The copies stay
// separate rather than becoming one shared helper, since exporting it would make a test convenience part of the
// library's public API.
func printExpr(d bson.D) {
	b, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
