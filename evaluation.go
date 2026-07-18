package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Regex returns a filter that matches documents where field matches the regular expression pattern.
//
// $regex matches string values against pattern, applying options as $options flags: "i" for case-insensitivity,
// "m" for multiline (^ and $ match line boundaries), "x" to ignore unescaped whitespace and # comments in pattern,
// and "s" so "." matches newline characters too. An empty options string means no flags, not an error. Unless
// pattern is anchored to the start of the string (e.g. "^prefix") and field is indexed, $regex cannot use an index
// efficiently and scans every document.
//
// Example:
//
//	monq.Regex("email", "^alice", "i")
//	// bson.D{{Key: "email", Value: bson.D{{Key: "$regex", Value: "^alice"}, {Key: "$options", Value: "i"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/regex/
func Regex(field FieldPath, pattern, options string) bson.D {
	return bson.D{
		{Key: string(field), Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: options}}},
	}
}
