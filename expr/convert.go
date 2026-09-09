package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// ConvertOption configures the optional fields of a [Convert] expression.
type ConvertOption func(*bson.D)

// ConvertOnError returns a [ConvertOption] giving the value to use when a conversion fails.
//
// Without it a value that cannot be converted, such as "abc" to an int, fails the whole aggregation. With it the
// document survives carrying the fallback instead, which is what makes [Convert] usable over data that is not
// uniform.
func ConvertOnError(value any) ConvertOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "onError", Value: value})
	}
}

// ConvertOnNull returns a [ConvertOption] giving the value to use when the input is null or missing.
//
// Without it a null input converts to null.
func ConvertOnNull(value any) ConvertOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "onNull", Value: value})
	}
}

// Convert returns an expression converting a value to another BSON type.
//
// $convert is the general form the shorthand operators such as [ToInt] and [ToDate] stand for, and the only one
// that takes [ConvertOnError] and [ConvertOnNull]. The target type is a string such as "int", "long", "double",
// "decimal", "string", "bool", "date", or "objectId", or the numeric type code.
//
// Conversions that lose information are rejected rather than rounded: a double with a fractional part does not
// become an int. Without a fallback, one bad value fails the aggregation for every document.
//
// Example:
//
//	expr.Convert(expr.Field("legacy_id"), "objectId", expr.ConvertOnError(nil))
//	// bson.D{
//	//     {
//	//         Key: "$convert",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$legacy_id"},
//	//             {Key: "to", Value: "objectId"},
//	//             {Key: "onError", Value: nil},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/convert/
func Convert(input, to any, opts ...ConvertOption) bson.D {
	spec := bson.D{{Key: "input", Value: input}, {Key: "to", Value: to}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$convert", Value: spec}}
}

// IsNumber returns an expression that is true when a value is any numeric type.
//
// $isNumber covers int, long, double, and decimal at once, which is what makes it easier than testing [Type]
// against four names. It yields false rather than failing for null, missing, and every other type.
//
// Example:
//
//	expr.IsNumber(expr.Field("score"))
//	// bson.D{{Key: "$isNumber", Value: "$score"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isNumber/
func IsNumber(value any) bson.D {
	return bson.D{{Key: "$isNumber", Value: value}}
}

// ToBool returns an expression converting a value to a boolean.
//
// $toBool treats 0 as false and every other number as true, and any string at all, "false" included, as true. A
// null or missing input yields null. It is [Convert] to "bool" without the fallbacks.
//
// Example:
//
//	expr.ToBool(expr.Field("flag"))
//	// bson.D{{Key: "$toBool", Value: "$flag"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toBool/
func ToBool(value any) bson.D {
	return bson.D{{Key: "$toBool", Value: value}}
}

// ToDate returns an expression converting a value to a date.
//
// $toDate reads an ISO 8601 string, a number as milliseconds since the epoch, or an ObjectID by its embedded
// timestamp. A string it cannot parse fails the aggregation, which is when [Convert] with [ConvertOnError] is
// worth the extra words. [DateFromString] is the one that takes a format.
//
// Example:
//
//	expr.ToDate(expr.Field("created_on"))
//	// bson.D{{Key: "$toDate", Value: "$created_on"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toDate/
func ToDate(value any) bson.D {
	return bson.D{{Key: "$toDate", Value: value}}
}

// ToDecimal returns an expression converting a value to a decimal.
//
// $toDecimal keeps the precision that a double would lose, which is why money belongs in this type. Strings have
// to look like numbers, and booleans become 1 or 0.
//
// Example:
//
//	expr.ToDecimal(expr.Field("price"))
//	// bson.D{{Key: "$toDecimal", Value: "$price"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toDecimal/
func ToDecimal(value any) bson.D {
	return bson.D{{Key: "$toDecimal", Value: value}}
}

// ToDouble returns an expression converting a value to a double.
//
// $toDouble takes numbers, numeric strings, booleans as 1 or 0, and dates as milliseconds since the epoch. A
// decimal too large for a double fails the aggregation.
//
// Example:
//
//	expr.ToDouble(expr.Field("weight"))
//	// bson.D{{Key: "$toDouble", Value: "$weight"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toDouble/
func ToDouble(value any) bson.D {
	return bson.D{{Key: "$toDouble", Value: value}}
}

// ToInt returns an expression converting a value to a 32-bit integer.
//
// $toInt refuses anything it cannot represent exactly: a double with a fractional part, or a value beyond the
// 32-bit range, fails the aggregation rather than being truncated. [Trunc] or [Round] first is the way to accept
// the loss on purpose.
//
// Example:
//
//	expr.ToInt(expr.Field("quantity"))
//	// bson.D{{Key: "$toInt", Value: "$quantity"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toInt/
func ToInt(value any) bson.D {
	return bson.D{{Key: "$toInt", Value: value}}
}

// ToLong returns an expression converting a value to a 64-bit integer.
//
// $toLong behaves as [ToInt] does over a wider range, and reads a date as milliseconds since the epoch.
//
// Example:
//
//	expr.ToLong(expr.Field("quantity"))
//	// bson.D{{Key: "$toLong", Value: "$quantity"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toLong/
func ToLong(value any) bson.D {
	return bson.D{{Key: "$toLong", Value: value}}
}

// ToObjectID returns an expression converting a value to an ObjectID.
//
// $toObjectId takes a 24-character hexadecimal string and nothing else, so a string of any other length or with a
// non-hexadecimal character fails the aggregation. It is how a collection storing ids as strings is joined against
// one storing them properly.
//
// Example:
//
//	expr.ToObjectID(expr.Field("legacy_id"))
//	// bson.D{{Key: "$toObjectId", Value: "$legacy_id"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toObjectId/
func ToObjectID(value any) bson.D {
	return bson.D{{Key: "$toObjectId", Value: value}}
}

// ToString returns an expression converting a value to a string.
//
// $toString renders numbers, booleans as "true" or "false", dates as ISO 8601 strings, and ObjectIDs as their
// hexadecimal form. [DateToString] is the one that takes a format for dates.
//
// Example:
//
//	expr.ToString(expr.Field("quantity"))
//	// bson.D{{Key: "$toString", Value: "$quantity"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toString/
func ToString(value any) bson.D {
	return bson.D{{Key: "$toString", Value: value}}
}

// Type returns an expression yielding the BSON type of a value as a string.
//
// $type reports a type where the query operator of the same name tests one. The query operator monq.Type matches
// documents whose field has one of the types it is given, while this returns the name, such as "string", "double",
// or "missing" for a field that is not there. [IsNumber] is the shorter way to ask about numbers.
//
// Example:
//
//	expr.Type(expr.Field("legacy_id"))
//	// bson.D{{Key: "$type", Value: "$legacy_id"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/type/
func Type(value any) bson.D {
	return bson.D{{Key: "$type", Value: value}}
}
