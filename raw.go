package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Raw returns d unchanged. It is the escape hatch for MongoDB operators monq does not (yet) provide a named function
// for: write the bson.D by hand and drop it in anywhere a monq expression is expected, including inside [And], [Or],
// and [Nor].
//
// Raw makes partial adoption of monq viable: hand-written and monq-built query fragments compose freely.
func Raw(d bson.D) bson.D {
	return d
}
