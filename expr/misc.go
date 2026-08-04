package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// BSONSize returns an expression yielding the size in bytes of a document.
//
// $bsonSize measures the BSON encoding of a document, which is what the 16 MB document limit is measured against,
// and takes null for a null input. Pass "$$ROOT" to measure the whole document being processed. It works on
// documents only; [BinarySize] is the one for strings and binary data.
//
// Example:
//
//	expr.BSONSize("$$ROOT")
//	// bson.D{{Key: "$bsonSize", Value: "$$ROOT"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bsonSize/
func BSONSize(document any) bson.D {
	return bson.D{{Key: "$bsonSize", Value: document}}
}

// BinarySize returns an expression yielding the size in bytes of a string or binary value.
//
// $binarySize counts encoded bytes for a string, so it agrees with [StrLenBytes] rather than [StrLenCP], and
// counts the data length for binData. [BSONSize] is the one for whole documents.
//
// Example:
//
//	expr.BinarySize(expr.Field("thumbnail"))
//	// bson.D{{Key: "$binarySize", Value: "$thumbnail"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/binarySize/
func BinarySize(value any) bson.D {
	return bson.D{{Key: "$binarySize", Value: value}}
}

// Let returns an expression that binds variables for the duration of another expression.
//
// $let names intermediate results so a long expression stays readable and a subexpression is computed once rather
// than repeated. Inside in, a variable declared as total is written "$$total", the same two-dollar form [Filter]
// and [Reduce] use for the variables they bind. The bindings last only for that in expression, unlike the let of a
// $lookup, which spans a whole sub-pipeline.
//
// The variables are a plain document rather than option constructors, since their names are the caller's to choose
// and there is nothing fixed to name a constructor after.
//
// Example:
//
//	expr.Let(
//		bson.D{{Key: "total", Value: expr.Add(expr.Field("price"), expr.Field("tax"))}},
//		expr.Multiply("$$total", 0.9),
//	)
//	// bson.D{{Key: "$let", Value: bson.D{
//	//     {Key: "vars", Value: bson.D{{Key: "total", Value: ...}}},
//	//     {Key: "in", Value: bson.D{{Key: "$multiply", Value: bson.A{"$$total", 0.9}}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/let/
func Let(vars bson.D, in any) bson.D {
	return bson.D{{Key: "$let", Value: bson.D{
		{Key: "vars", Value: vars},
		{Key: "in", Value: in},
	}}}
}

// Rand returns an expression yielding a random double between 0 and 1.
//
// $rand takes no arguments and is evaluated afresh every time it appears, so two references in one expression give
// two different numbers; bind it with [Let] when the same draw is needed twice. Nothing about it is stable across
// runs, which makes it useful for sampling and useless for anything that has to be reproducible. The query-side
// counterpart for keeping a fraction of the documents is monq.SampleRate.
//
// Example:
//
//	expr.Rand()
//	// bson.D{{Key: "$rand", Value: bson.D{}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/rand/
func Rand() bson.D {
	return bson.D{{Key: "$rand", Value: bson.D{}}}
}

// TsIncrement returns an expression yielding the ordinal part of a BSON timestamp.
//
// A BSON timestamp pairs a seconds value with an increment that orders the operations recorded within the same
// second. $tsIncrement returns that second part as a long. It works on timestamps only, not on dates, and
// [TsSecond] returns the other half.
//
// Example:
//
//	expr.TsIncrement(expr.Field("oplog_ts"))
//	// bson.D{{Key: "$tsIncrement", Value: "$oplog_ts"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/tsIncrement/
func TsIncrement(timestamp any) bson.D {
	return bson.D{{Key: "$tsIncrement", Value: timestamp}}
}

// TsSecond returns an expression yielding the seconds part of a BSON timestamp.
//
// $tsSecond returns the seconds since the epoch held in a BSON timestamp, as a long. Timestamps are an internal
// type used by replication rather than a way to store application dates, so this mostly turns up when reading the
// oplog or a change stream. [TsIncrement] returns the ordinal that goes with it.
//
// Example:
//
//	expr.TsSecond(expr.Field("oplog_ts"))
//	// bson.D{{Key: "$tsSecond", Value: "$oplog_ts"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/tsSecond/
func TsSecond(timestamp any) bson.D {
	return bson.D{{Key: "$tsSecond", Value: timestamp}}
}
