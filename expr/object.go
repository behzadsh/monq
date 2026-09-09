package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// GetField returns an expression reading a field of a document by name.
//
// $getField is how a field whose name would otherwise be unreadable gets read: one containing a dot, which a
// normal reference would take as a path, or one starting with a dollar sign, which a normal reference would take
// as an operator. The field name is a string or an expression producing one, and it is not a path, so dots in it
// are part of the name.
//
// Pass "$$CURRENT" as the input to read from the document being processed, which is what MongoDB's shorthand form
// of the operator means. Any other document expression reads from that instead.
//
// Example:
//
//	expr.GetField("price.usd", "$$CURRENT")
//	// bson.D{
//	//     {
//	//         Key: "$getField",
//	//         Value: bson.D{
//	//             {Key: "field", Value: "price.usd"},
//	//             {Key: "input", Value: "$$CURRENT"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/getField/
func GetField(field, input any) bson.D {
	return bson.D{
		{
			Key: "$getField",
			Value: bson.D{
				{Key: "field", Value: field},
				{Key: "input", Value: input},
			},
		},
	}
}

// MergeObjects returns an expression combining documents into one.
//
// $mergeObjects overlays its arguments left to right, so a field present in more than one takes the value from the
// last, and null or missing arguments are skipped rather than nulling the result. It is what fills in a document
// with defaults, and it is also the accumulator of the same name, merging every document of a group.
//
// Example:
//
//	expr.MergeObjects(expr.Field("defaults"), expr.Field("overrides"))
//	// bson.D{{Key: "$mergeObjects", Value: bson.A{"$defaults", "$overrides"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/mergeObjects/
func MergeObjects(documents ...any) bson.D {
	return bson.D{{Key: "$mergeObjects", Value: operands(documents)}}
}

// SetField returns an expression adding or replacing a field of a document.
//
// $setField is the counterpart of [GetField] and takes names the same way, as literal strings rather than paths,
// which makes it the way to write a field whose name holds a dot or starts with a dollar sign. It returns a new
// document; nothing is modified in place. Setting a field to the "$$REMOVE" variable deletes it, which is what
// [UnsetField] does more plainly.
//
// Example:
//
//	expr.SetField("price.usd", "$$CURRENT", 42)
//	// bson.D{
//	//     {
//	//         Key: "$setField",
//	//         Value: bson.D{
//	//             {Key: "field", Value: "price.usd"},
//	//             {Key: "input", Value: "$$CURRENT"},
//	//             {Key: "value", Value: 42},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setField/
func SetField(field, input, value any) bson.D {
	return bson.D{
		{
			Key: "$setField",
			Value: bson.D{
				{Key: "field", Value: field},
				{Key: "input", Value: input},
				{Key: "value", Value: value},
			},
		},
	}
}

// UnsetField returns an expression removing a field from a document.
//
// $unsetField returns a new document without that field, and does nothing when the field is not there. Like
// [GetField] and [SetField] it treats the name as a literal, so it reaches fields that a dotted path cannot name.
//
// Example:
//
//	expr.UnsetField("price.usd", "$$CURRENT")
//	// bson.D{
//	//     {
//	//         Key: "$unsetField",
//	//         Value: bson.D{
//	//             {Key: "field", Value: "price.usd"},
//	//             {Key: "input", Value: "$$CURRENT"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unsetField/
func UnsetField(field, input any) bson.D {
	return bson.D{
		{
			Key: "$unsetField",
			Value: bson.D{
				{Key: "field", Value: field},
				{Key: "input", Value: input},
			},
		},
	}
}
