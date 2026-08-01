package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// CurrentDate returns an update that sets field to the current date.
//
// $currentDate writes the server's current time into the field, creating it if it is missing. The true form used
// here stores a BSON date; [CurrentDateTimestamp] stores a BSON timestamp instead. Combine it with other operators
// through [Update]; two operator documents concatenated by hand keep two separate keys, which MongoDB does not
// merge.
//
// Example:
//
//	monq.CurrentDate("updated_at")
//	// bson.D{{Key: "$currentDate", Value: bson.D{{Key: "updated_at", Value: true}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/currentDate/
func CurrentDate(field FieldPath) bson.D {
	return bson.D{{Key: "$currentDate", Value: bson.D{{Key: string(field), Value: true}}}}
}

// CurrentDateTimestamp returns an update that sets field to the current BSON timestamp.
//
// It is the {$type: "timestamp"} form of $currentDate, a separate function rather than an option on [CurrentDate]
// because it changes the shape of the operator document instead of adding to it. BSON timestamps are meant for
// internal replication use; store a date with [CurrentDate] unless a timestamp is what the application really
// wants.
//
// Example:
//
//	monq.CurrentDateTimestamp("synced_at")
//	// bson.D{{Key: "$currentDate", Value: bson.D{
//	//     {Key: "synced_at", Value: bson.D{{Key: "$type", Value: "timestamp"}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/currentDate/
func CurrentDateTimestamp(field FieldPath) bson.D {
	value := bson.D{{Key: "$type", Value: "timestamp"}}

	return bson.D{{Key: "$currentDate", Value: bson.D{{Key: string(field), Value: value}}}}
}

// Inc returns an update that increases field by amount.
//
// $inc adds amount to the field's current value, treating a missing field as 0 so the field is created with
// amount. A negative amount decreases the value, since MongoDB has no $dec. The field must hold a number:
// incrementing a string or null is a server error. Combine it with other operators through [Update]; two operator
// documents concatenated by hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.Inc("logins", 1)
//	// bson.D{{Key: "$inc", Value: bson.D{{Key: "logins", Value: 1}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/inc/
func Inc(field FieldPath, amount any) bson.D {
	return bson.D{{Key: "$inc", Value: bson.D{{Key: string(field), Value: amount}}}}
}

// Max returns an update that raises field to value when value is the larger of the two.
//
// $max compares the field's current value with value using BSON comparison order, which covers dates and strings
// as well as numbers, and writes value only when it is greater. A missing field is simply set. Combine it with
// other operators through [Update]; two operator documents concatenated by hand keep two separate keys, which
// MongoDB does not merge.
//
// Example:
//
//	monq.Max("high_score", 250)
//	// bson.D{{Key: "$max", Value: bson.D{{Key: "high_score", Value: 250}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/max/
func Max(field FieldPath, value any) bson.D {
	return bson.D{{Key: "$max", Value: bson.D{{Key: string(field), Value: value}}}}
}

// Min returns an update that lowers field to value when value is the smaller of the two.
//
// $min is the counterpart of [Max]: it compares with BSON comparison order and writes value only when it is less
// than the current one, setting a missing field outright. Combine it with other operators through [Update]; two
// operator documents concatenated by hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.Min("best_time", 42)
//	// bson.D{{Key: "$min", Value: bson.D{{Key: "best_time", Value: 42}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/min/
func Min(field FieldPath, value any) bson.D {
	return bson.D{{Key: "$min", Value: bson.D{{Key: string(field), Value: value}}}}
}

// Mul returns an update that multiplies field by factor.
//
// $mul multiplies the field's current value by factor and creates a missing field as 0 of the same numeric type as
// factor, not as factor itself. The result's type follows MongoDB's numeric promotion rules, so multiplying an int
// by a double gives a double. Combine it with other operators through [Update]; two operator documents
// concatenated by hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.Mul("price", 1.1)
//	// bson.D{{Key: "$mul", Value: bson.D{{Key: "price", Value: 1.1}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/mul/
func Mul(field FieldPath, factor any) bson.D {
	return bson.D{{Key: "$mul", Value: bson.D{{Key: string(field), Value: factor}}}}
}

// Rename returns an update that renames field to newName.
//
// $rename removes the field and re-adds it under the new name, keeping the value. Both names are paths from the
// document root, and neither may point through an array: renaming across array elements is not supported. An
// existing field at newName is overwritten, and a missing source field makes the operator a no-op. Combine it with
// other operators through [Update]; two operator documents concatenated by hand keep two separate keys, which
// MongoDB does not merge.
//
// Example:
//
//	monq.Rename("nickname", "profile.display_name")
//	// bson.D{{Key: "$rename", Value: bson.D{{Key: "nickname", Value: "profile.display_name"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/rename/
func Rename(field, newName FieldPath) bson.D {
	return bson.D{{Key: "$rename", Value: bson.D{{Key: string(field), Value: string(newName)}}}}
}

// Set returns an update that writes value to field.
//
// $set replaces the field's value, creating the field, and any missing parent documents along a dotted path, when
// it is not there. It is the workhorse of update documents.
//
// Every monq update operator returns a complete update document for one field, so a single [Set] can be handed to
// UpdateOne as is. Two of them cannot be concatenated: bson.D{{"$set", ...}, {"$set", ...}} has a duplicate key
// that MongoDB does not merge, and the second one is what survives. Pass them to [Update] instead, which merges
// operators sharing a key.
//
// Example:
//
//	monq.Set("status", "active")
//	// bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/set/
func Set(field FieldPath, value any) bson.D {
	return bson.D{{Key: "$set", Value: bson.D{{Key: string(field), Value: value}}}}
}

// SetOnInsert returns an update that writes value to field only when the operation inserts a new document.
//
// $setOnInsert applies to upserts alone: if the upsert matches an existing document the operator does nothing, and
// on a plain update it is ignored entirely. It is how creation-time defaults are attached to an upsert without
// overwriting them on later updates. Combine it with other operators through [Update]; two operator documents
// concatenated by hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.SetOnInsert("created_at", time.Now())
//	// bson.D{{Key: "$setOnInsert", Value: bson.D{{Key: "created_at", Value: time.Now()}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/setOnInsert/
func SetOnInsert(field FieldPath, value any) bson.D {
	return bson.D{{Key: "$setOnInsert", Value: bson.D{{Key: string(field), Value: value}}}}
}

// Unset returns an update that removes field from the document.
//
// $unset deletes the field outright rather than setting it to null, and does nothing when the field is not there.
// The operand is the empty string only because MongoDB requires some value and ignores whichever one it gets.
// Removing an array element with $unset leaves a null hole in the array instead of shortening it, which is what
// the $pull family is for. Combine it with other operators through [Update]; two operator documents concatenated
// by hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.Unset("deleted_at")
//	// bson.D{{Key: "$unset", Value: bson.D{{Key: "deleted_at", Value: ""}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/unset/
func Unset(field FieldPath) bson.D {
	return bson.D{{Key: "$unset", Value: bson.D{{Key: string(field), Value: ""}}}}
}
