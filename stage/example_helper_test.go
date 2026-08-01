package stage_test

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// printStage prints d as relaxed extended JSON. Example tests use it instead of fmt.Println(d) whenever the stage
// contains a numeric value, since bson.D's default String() renders numbers in verbose canonical extended JSON
// (e.g. {"$numberInt":"20"}), which is accurate but not what a reader expects to see in a godoc example.
//
// It mirrors printFilter in the root package's tests. The two stay separate copies rather than one shared helper,
// since exporting it would make a test convenience part of the library's public API.
func printStage(d bson.D) {
	b, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
