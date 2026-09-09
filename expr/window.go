package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// WindowOption configures the optional fields of a window operator.
//
// [TimeUnit] scales [Derivative] and [Integral], and [ShiftDefault] fills the gap [Shift] leaves at the edge of a
// partition. The type is shared because both operators are window operators; passing one to the wrong operator
// compiles and produces a document the server rejects.
type WindowOption func(*bson.D)

// ShiftDefault returns a [WindowOption] giving the value [Shift] yields when it points outside the partition.
//
// Without it the result is null for the documents at the edge, where there is no document that far away.
func ShiftDefault(value any) WindowOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "default", Value: value})
	}
}

// TimeUnit returns a [WindowOption] setting the time unit of [Derivative] or [Integral].
//
// The unit is one of "week", "day", "hour", "minute", "second", or "millisecond", and it turns a rate per
// millisecond, which is what a date-sorted window gives by default, into a rate per something a reader
// understands. It has no meaning when the sort key is a plain number.
func TimeUnit(unit any) WindowOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "unit", Value: unit})
	}
}

// CovariancePop returns a window operator yielding the population covariance of two expressions.
//
// $covariancePop measures how two values move together across the documents of a window. It treats them as the
// whole population; [CovarianceSamp] treats them as a sample. Non-numeric pairs are skipped.
//
// It is only valid inside a $setWindowFields stage.
//
// Example:
//
//	expr.CovariancePop(expr.Field("x"), expr.Field("y"))
//	// bson.D{{Key: "$covariancePop", Value: bson.A{"$x", "$y"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/covariancePop/
func CovariancePop(a, b any) bson.D {
	return bson.D{{Key: "$covariancePop", Value: bson.A{a, b}}}
}

// CovarianceSamp returns a window operator yielding the sample covariance of two expressions.
//
// $covarianceSamp divides by one less than the count, the sample convention, where [CovariancePop] divides by the
// count itself. A window holding fewer than two documents yields null.
//
// It is only valid inside a $setWindowFields stage.
//
// Example:
//
//	expr.CovarianceSamp(expr.Field("x"), expr.Field("y"))
//	// bson.D{{Key: "$covarianceSamp", Value: bson.A{"$x", "$y"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/covarianceSamp/
func CovarianceSamp(a, b any) bson.D {
	return bson.D{{Key: "$covarianceSamp", Value: bson.A{a, b}}}
}

// DenseRank returns a window operator yielding the rank of a document, leaving no gaps after a tie.
//
// $denseRank numbers the documents by their sort key, giving tied documents the same number and continuing from
// the next one: 1, 2, 2, 3. [Rank] leaves a gap instead, and [DocumentNumber] never ties at all. It takes no
// arguments.
//
// It is only valid inside a $setWindowFields stage, which has to sort by exactly one key, and it cannot be given a
// window.
//
// Example:
//
//	expr.DenseRank()
//	// bson.D{{Key: "$denseRank", Value: bson.D{}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/denseRank/
func DenseRank() bson.D {
	return bson.D{{Key: "$denseRank", Value: bson.D{}}}
}

// Derivative returns a window operator yielding the rate of change across a window.
//
// $derivative divides the change in the input by the change in the sort key between the first and last documents
// of the window, so it answers how fast something is moving. With a date sort key the rate is per millisecond
// unless [TimeUnit] says otherwise.
//
// It is only valid inside a $setWindowFields stage, needs a sortBy, and a window is worth giving: without one it
// measures across the whole partition.
//
// Example:
//
//	expr.Derivative(expr.Field("odometer"), expr.TimeUnit("hour"))
//	// bson.D{
//	//     {
//	//         Key: "$derivative",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$odometer"},
//	//             {Key: "unit", Value: "hour"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/derivative/
func Derivative(input any, opts ...WindowOption) bson.D {
	return bson.D{{Key: "$derivative", Value: windowInput(input, opts)}}
}

// DocumentNumber returns a window operator yielding the position of a document in its partition.
//
// $documentNumber numbers documents from 1 and never repeats a number, so tied documents still come out in some
// order: 1, 2, 3, 4. [Rank] and [DenseRank] are the ones that share a number between ties. It takes no arguments.
//
// It is only valid inside a $setWindowFields stage, needs a sortBy, and cannot be given a window.
//
// Example:
//
//	expr.DocumentNumber()
//	// bson.D{{Key: "$documentNumber", Value: bson.D{}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/documentNumber/
func DocumentNumber() bson.D {
	return bson.D{{Key: "$documentNumber", Value: bson.D{}}}
}

// ExpMovingAvgAlpha returns a window operator yielding an exponential moving average with a given smoothing factor.
//
// The alpha is a number between 0 and 1 that says how much weight the current document carries, so a larger one
// follows recent values more closely. [ExpMovingAvgN] is the same operator expressed as a number of documents;
// MongoDB accepts one or the other and never both, which is why they are separate functions here.
//
// It is only valid inside a $setWindowFields stage, needs a sortBy, and takes no window.
//
// Example:
//
//	expr.ExpMovingAvgAlpha(expr.Field("price"), 0.3)
//	// bson.D{
//	//     {
//	//         Key: "$expMovingAvg",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$price"},
//	//             {Key: "alpha", Value: 0.3},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/expMovingAvg/
func ExpMovingAvgAlpha(input any, alpha float64) bson.D {
	return bson.D{
		{
			Key: "$expMovingAvg",
			Value: bson.D{
				{Key: "input", Value: input},
				{Key: "alpha", Value: alpha},
			},
		},
	}
}

// ExpMovingAvgN returns a window operator yielding an exponential moving average over n documents.
//
// The n is how many documents the average leans on, which MongoDB turns into a smoothing factor of 2/(n+1).
// [ExpMovingAvgAlpha] takes that factor directly instead. Unlike a plain moving average, this one never forgets a
// document completely; it only weighs the older ones less.
//
// It is only valid inside a $setWindowFields stage, needs a sortBy, and takes no window.
//
// Example:
//
//	expr.ExpMovingAvgN(expr.Field("price"), 5)
//	// bson.D{
//	//     {
//	//         Key: "$expMovingAvg",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$price"},
//	//             {Key: "N", Value: 5},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/expMovingAvg/
func ExpMovingAvgN(input any, n int) bson.D {
	return bson.D{
		{
			Key: "$expMovingAvg",
			Value: bson.D{
				{Key: "input", Value: input},
				{Key: "N", Value: n},
			},
		},
	}
}

// Integral returns a window operator yielding the area under the curve of a value over a window.
//
// $integral is the counterpart of [Derivative]: where that one measures a rate, this one accumulates a quantity,
// which is what turns a series of readings into a total. It uses the trapezoidal rule between consecutive
// documents. With a date sort key the result is per millisecond unless [TimeUnit] says otherwise.
//
// It is only valid inside a $setWindowFields stage and needs a sortBy.
//
// Example:
//
//	expr.Integral(expr.Field("power"), expr.TimeUnit("hour"))
//	// bson.D{
//	//     {
//	//         Key: "$integral",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$power"},
//	//             {Key: "unit", Value: "hour"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/integral/
func Integral(input any, opts ...WindowOption) bson.D {
	return bson.D{{Key: "$integral", Value: windowInput(input, opts)}}
}

// LinearFill returns a window operator filling null gaps by interpolating between the values around them.
//
// $linearFill looks at the last value before a gap and the first one after it and spreads the difference evenly
// across the documents between, in proportion to the sort key. A gap at the start or the end of a partition, with
// a value on one side only, stays null. [Locf] is the blunter alternative that repeats the previous value.
//
// It is only valid inside a $setWindowFields stage and needs a sortBy.
//
// Example:
//
//	expr.LinearFill(expr.Field("price"))
//	// bson.D{{Key: "$linearFill", Value: "$price"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/linearFill/
func LinearFill(input any) bson.D {
	return bson.D{{Key: "$linearFill", Value: input}}
}

// Locf returns a window operator filling null gaps with the last value seen.
//
// The name is short for last observation carried forward: where the input is null or missing, the result is
// whatever it was in the nearest earlier document of the partition. A gap before any value at all stays null.
// [LinearFill] interpolates instead of repeating.
//
// It is only valid inside a $setWindowFields stage and needs a sortBy.
//
// Example:
//
//	expr.Locf(expr.Field("price"))
//	// bson.D{{Key: "$locf", Value: "$price"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/locf/
func Locf(input any) bson.D {
	return bson.D{{Key: "$locf", Value: input}}
}

// Rank returns a window operator yielding the rank of a document, leaving a gap after a tie.
//
// $rank numbers documents by their sort key, giving tied documents the same number and then skipping ahead: 1, 2,
// 2, 4. [DenseRank] closes those gaps and [DocumentNumber] never ties. It takes no arguments.
//
// It is only valid inside a $setWindowFields stage, which has to sort by exactly one key, and it cannot be given a
// window.
//
// Example:
//
//	expr.Rank()
//	// bson.D{{Key: "$rank", Value: bson.D{}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/rank/
func Rank() bson.D {
	return bson.D{{Key: "$rank", Value: bson.D{}}}
}

// Shift returns a window operator yielding a value from another document of the partition.
//
// $shift reaches by, a number of documents, away from the current one along the sort order, which is how a row is
// compared with the one before or after it: a by of -1 is the previous document and 1 the next. Where it points
// outside the partition the result is null unless [ShiftDefault] gives something else.
//
// It is only valid inside a $setWindowFields stage, needs a sortBy, and takes no window: the shift is its own
// notion of position. [ShiftDefault] is the only option that means anything here.
//
// Example:
//
//	expr.Shift(expr.Field("price"), -1, expr.ShiftDefault(0))
//	// bson.D{
//	//     {
//	//         Key: "$shift",
//	//         Value: bson.D{
//	//             {Key: "output", Value: "$price"},
//	//             {Key: "by", Value: -1},
//	//             {Key: "default", Value: 0},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/shift/
func Shift(output any, by int, opts ...WindowOption) bson.D {
	spec := bson.D{{Key: "output", Value: output}, {Key: "by", Value: by}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$shift", Value: spec}}
}

// windowInput builds the {input, unit} specification $derivative and $integral share.
func windowInput(input any, opts []WindowOption) bson.D {
	spec := bson.D{{Key: "input", Value: input}}
	for _, opt := range opts {
		opt(&spec)
	}

	return spec
}
