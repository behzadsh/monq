package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// Window bounds that name a position rather than a number of documents or a span of values.
//
// WindowUnbounded is the start or the end of the partition, and WindowCurrent is the document being computed. A
// window of {WindowUnbounded, WindowCurrent} is the running total from the beginning of the partition up to here,
// which is the most common window there is.
const (
	WindowUnbounded = "unbounded"
	WindowCurrent   = "current"
)

// WindowOption configures the window an output field of [SetWindowFields] is computed over.
//
// A window narrows an accumulator to part of the partition instead of all of it. [WindowDocuments] counts in
// documents and [WindowRange] measures in values of the sort key, with [WindowUnit] making that a span of time.
// Operators that carry their own ordering, [github.com/behzadsh/monq/expr.Rank] and its neighbors among them,
// reject a window outright.
type WindowOption func(*bson.D)

// WindowDocuments returns a [WindowOption] bounding the window by a count of documents.
//
// The bounds are positions relative to the current document: a negative number counts back, a positive one counts
// forward, and [WindowUnbounded] and [WindowCurrent] name the edges. WindowDocuments(-3, 0) is the current
// document and the three before it, a four-document moving window, and
// WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent) is a running total.
func WindowDocuments(lower, upper any) WindowOption {
	return func(d *bson.D) {
		setWindowBound(d, "documents", bson.A{lower, upper})
	}
}

// WindowRange returns a [WindowOption] bounding the window by the value of the sort key.
//
// Where [WindowDocuments] counts documents, this compares values, so it takes in however many documents fall
// within that span of the sort key. That makes it the right choice over irregular data, where three documents back
// might be a minute or a week. With [WindowUnit] the bounds are a span of time rather than plain numbers.
func WindowRange(lower, upper any) WindowOption {
	return func(d *bson.D) {
		setWindowBound(d, "range", bson.A{lower, upper})
	}
}

// WindowUnit returns a [WindowOption] making a [WindowRange] a span of time.
//
// The unit is one of "year", "quarter", "month", "week", "day", "hour", "minute", "second", or "millisecond", and
// the sort key has to be a date for it to mean anything. It has no effect without [WindowRange]; a window bounded
// by documents counts documents whatever this says.
func WindowUnit(unit any) WindowOption {
	return func(d *bson.D) {
		setWindowBound(d, "unit", unit)
	}
}

// SetWindowFields returns a stage that computes values over a window of related documents.
//
// $setWindowFields is what MongoDB has in place of SQL's window functions. It splits the documents into partitions
// by the partitionBy expression, orders each partition by sortBy, and adds the output fields to every document,
// computing each one over a window of its partition rather than collapsing the partition the way a $group would.
// Every input document comes out, with its new fields attached.
//
// Pass nil for partitionBy to treat the whole collection as one partition, and nil for sortBy when nothing needs
// ordering. Plenty needs ordering though: [github.com/behzadsh/monq/expr.Rank], Shift, Locf, LinearFill,
// Derivative, and Integral all fail without a sortBy, and the two ranking operators want exactly one sort key.
//
// Output fields come from [WindowField]. Each pairs an accumulator with an optional window, which is the one place
// in the aggregation language where an accumulator carries a sibling key.
//
// Example:
//
//	stage.SetWindowFields("$account", bson.D{{Key: "date", Value: 1}},
//		stage.WindowField("running_total", expr.Sum(expr.Field("amount")),
//			stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent)),
//	)
//	// bson.D{{Key: "$setWindowFields", Value: bson.D{
//	//     {Key: "partitionBy", Value: "$account"},
//	//     {Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
//	//     {Key: "output", Value: bson.D{{Key: "running_total", Value: bson.D{
//	//         {Key: "$sum", Value: "$amount"},
//	//         {Key: "window", Value: bson.D{{Key: "documents", Value: bson.A{"unbounded", "current"}}}},
//	//     }}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setWindowFields/
func SetWindowFields(partitionBy any, sortBy bson.D, output ...bson.D) bson.D {
	spec := bson.D{}
	if partitionBy != nil {
		spec = append(spec, bson.E{Key: "partitionBy", Value: partitionBy})
	}

	if sortBy != nil {
		spec = append(spec, bson.E{Key: "sortBy", Value: sortBy})
	}

	spec = append(spec, bson.E{Key: "output", Value: mergeFields(bson.D{}, output)})

	return bson.D{{Key: "$setWindowFields", Value: spec}}
}

// WindowField returns one output field of a [SetWindowFields] stage.
//
// The operator is an accumulator such as [github.com/behzadsh/monq/expr.Sum], or one of the operators that only
// exist inside this stage, such as [github.com/behzadsh/monq/expr.Rank]. Options add the window it is computed
// over; without one the accumulator sees the whole partition. The operator document is copied rather than written
// into, so the same expression can be reused across several fields.
//
// The name is a path like any other field this library writes, so it may reach into a subdocument and it overwrites
// whatever was there.
//
// Example:
//
//	stage.WindowField("running_total", expr.Sum(expr.Field("amount")),
//		stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent))
//	// bson.D{{Key: "running_total", Value: bson.D{
//	//     {Key: "$sum", Value: "$amount"},
//	//     {Key: "window", Value: bson.D{{Key: "documents", Value: bson.A{"unbounded", "current"}}}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/setWindowFields/
func WindowField(name monq.FieldPath, operator bson.D, opts ...WindowOption) bson.D {
	spec := make(bson.D, len(operator), len(operator)+1)
	copy(spec, operator)

	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: string(name), Value: spec}}
}

// setWindowBound adds one key to the window document of an output field, creating that document the first time.
// The field specification is seeded by [WindowField], so the window is the only collection an option extends here.
func setWindowBound(d *bson.D, key string, value any) {
	i := indexOfKey(*d, "window")
	if i < 0 {
		*d = append(*d, bson.E{Key: "window", Value: bson.D{{Key: key, Value: value}}})

		return
	}

	window, ok := (*d)[i].Value.(bson.D)
	if !ok {
		return
	}

	(*d)[i].Value = append(window, bson.E{Key: key, Value: value})
}
