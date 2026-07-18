package monq_test

import (
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// printFilter prints d as relaxed extended JSON. Example tests use it instead of fmt.Println(d) whenever the
// filter contains a numeric value, since bson.D's default String() renders numbers in verbose canonical extended
// JSON (e.g. {"$numberInt":"18"}), which is accurate but not what a reader expects to see in a godoc example.
func printFilter(d bson.D) {
	b, err := bson.MarshalExtJSON(d, false, false)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(b))
}
