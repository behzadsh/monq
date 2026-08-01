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

// Type returns a filter that matches documents where field's BSON type is one of types.
//
// $type matches on the type the value is stored as, named either by string alias ("string", "objectId", "double")
// or by numeric type code; the alias "number" covers every numeric type at once. For an array field it matches when
// the array itself has the given type or when any of its elements does. One type is emitted as a scalar and several
// as an array, the two forms MongoDB accepts. Types is variadic with the same un-spread-slice gotcha as [In]: spread
// it, Type(field, wanted...), rather than passing the slice directly.
//
// Example:
//
//	monq.Type("legacyId", "string", "objectId")
//	// bson.D{{Key: "legacyId", Value: bson.D{{Key: "$type", Value: bson.A{"string", "objectId"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/type/
func Type(field FieldPath, types ...any) bson.D {
	var value any
	if len(types) == 1 {
		value = types[0]
	} else {
		arr := make(bson.A, len(types))
		copy(arr, types)
		value = arr
	}

	return bson.D{{Key: string(field), Value: bson.D{{Key: "$type", Value: value}}}}
}
