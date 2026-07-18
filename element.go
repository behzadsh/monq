package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Exists returns a filter that matches documents based on whether field is present.
//
// $exists matches documents that contain the field (including fields whose value is null) when exists is true, and
// documents that do not contain the field at all when exists is false. It tests only the field's presence, not its
// value.
//
// Example:
//
//	monq.Exists("verified_at", true)
//	// bson.D{{Key: "verified_at", Value: bson.D{{Key: "$exists", Value: true}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/exists/
func Exists(field FieldPath, exists bool) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$exists", Value: exists}}}}
}
