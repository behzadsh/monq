package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// Bounds accepted by [DensifyRange] in place of an explicit pair of values.
//
// DensifyFull spans from the smallest to the largest value across every document the stage sees, so the same
// series is generated for every partition. DensifyPartition spans each partition's own range instead, which leaves
// partitions with different extents ending at different points.
const (
	DensifyFull      = "full"
	DensifyPartition = "partition"
)

// DensifyOption configures the optional fields of a [Densify] stage.
type DensifyOption func(*bson.D)

// DensifyPartitionByFields returns a [DensifyOption] naming the fields that separate one series from another.
//
// Without it every document belongs to one series, so a gap is filled once across the whole collection. With it,
// each combination of values gets its own series and its own generated documents, which is what densifying a
// reading per sensor or a price per symbol needs.
func DensifyPartitionByFields(fields ...monq.FieldPath) DensifyOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "partitionByFields", Value: fieldNames(fields)})
	}
}

// Densify returns a stage that generates the documents missing from a series.
//
// $densify fills gaps in a sequence, adding documents where the field skips a step, so a series with nothing
// recorded on a Tuesday comes out with a Tuesday in it. Generated documents carry only the field itself and
// whatever the partition is keyed on; every other field is absent, which is why a [Fill] usually follows to give
// them values.
//
// The range comes from [DensifyRange]. Input documents pass through untouched, and the output is not sorted.
//
// Example:
//
//	stage.Densify("timestamp", stage.DensifyRange(1, stage.DensifyFull, "hour"))
//	// bson.D{
//	//     {
//	//         Key: "$densify",
//	//         Value: bson.D{
//	//             {Key: "field", Value: "timestamp"},
//	//             {
//	//                 Key: "range",
//	//                 Value: bson.D{
//	//                     {Key: "step", Value: 1},
//	//                     {Key: "bounds", Value: "full"},
//	//                     {Key: "unit", Value: "hour"},
//	//                 },
//	//             },
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/densify/
func Densify(field monq.FieldPath, rangeSpec bson.D, opts ...DensifyOption) bson.D {
	spec := make(bson.D, 0, len(opts)+2)
	spec = append(spec, bson.E{Key: "field", Value: string(field)})

	for _, opt := range opts {
		opt(&spec)
	}

	spec = append(spec, bson.E{Key: "range", Value: rangeSpec})

	return bson.D{{Key: "$densify", Value: spec}}
}

// DensifyRange returns the range specification a [Densify] stage fills in.
//
// The step is how far apart the generated values are. The bounds are [DensifyFull], [DensifyPartition], or an
// explicit bson.A of a lower and an upper value, where the lower one is included and the upper one is not. The
// optional unit turns the step into a span of time, one of "year", "quarter", "month", "week", "day", "hour",
// "minute", "second", or "millisecond", and is required when the field holds dates.
//
// Example:
//
//	stage.DensifyRange(1, stage.DensifyFull, "hour")
//	// bson.D{{Key: "step", Value: 1}, {Key: "bounds", Value: "full"}, {Key: "unit", Value: "hour"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/densify/
func DensifyRange(step, bounds any, unit ...any) bson.D {
	spec := make(bson.D, 0, len(unit)+2)
	spec = append(spec, bson.E{Key: "step", Value: step}, bson.E{Key: "bounds", Value: bounds})

	for _, u := range unit {
		spec = append(spec, bson.E{Key: "unit", Value: u})
	}

	return spec
}

// Fill returns a stage that gives values to fields that are missing or null.
//
// $fill replaces a null or absent field with something usable, which is what makes a sparse series safe to chart
// or to sum. Each output entry comes from [FillValue], which supplies a value of your own, or [FillMethod], which
// derives one from the documents around it. A field takes one or the other, never both.
//
// [FillSortBy] is required by the "linear" and "locf" methods, since both depend on the order of the documents,
// and [FillPartitionBy] or [FillPartitionByFields] keeps separate series from filling from each other. Those two
// are alternatives to one another and cannot both be given.
//
// Example:
//
//	stage.Fill([]bson.D{stage.FillMethod("price", "locf")}, stage.FillSortBy(monq.Sort(monq.Asc("date"))))
//	// bson.D{
//	//     {
//	//         Key: "$fill",
//	//         Value: bson.D{
//	//             {Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
//	//             {Key: "output", Value: bson.D{{Key: "price", Value: bson.D{{Key: "method", Value: "locf"}}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/fill/
func Fill(output []bson.D, opts ...FillOption) bson.D {
	spec := make(bson.D, 0, len(opts)+1)
	for _, opt := range opts {
		opt(&spec)
	}

	spec = append(spec, bson.E{Key: "output", Value: mergeFields(bson.D{}, output)})

	return bson.D{{Key: "$fill", Value: spec}}
}

// FillMethod returns one output entry of a [Fill] stage, filled from the surrounding documents.
//
// The method is "locf", which repeats the last value seen, or "linear", which interpolates between the values on
// either side of the gap. Both need the stage to have a [FillSortBy], and both leave a gap null when there is
// nothing to work from: "locf" at the start of a partition, "linear" at either end.
func FillMethod(name monq.FieldPath, method string) bson.D {
	return bson.D{{Key: string(name), Value: bson.D{{Key: "method", Value: method}}}}
}

// FillOption configures the optional fields of a [Fill] stage.
type FillOption func(*bson.D)

// FillPartitionBy returns a [FillOption] grouping documents by an expression before filling.
//
// Filling never crosses a partition, so this is what stops one series from borrowing a value from another. Use
// [FillPartitionByFields] to partition by field names instead; MongoDB accepts one or the other, not both.
func FillPartitionBy(expression any) FillOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "partitionBy", Value: expression})
	}
}

// FillPartitionByFields returns a [FillOption] grouping documents by the values of the given fields.
//
// It is the field-name form of [FillPartitionBy], and the two cannot both be given.
func FillPartitionByFields(fields ...monq.FieldPath) FillOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "partitionByFields", Value: fieldNames(fields)})
	}
}

// FillSortBy returns a [FillOption] ordering the documents within each partition.
//
// The "locf" and "linear" methods of [FillMethod] both depend on this order and fail without it. A sort document
// built by monq.Sort drops straight in.
func FillSortBy(sortBy bson.D) FillOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "sortBy", Value: sortBy})
	}
}

// FillValue returns one output entry of a [Fill] stage, filled with a value of your own.
//
// The value is an aggregation expression, so it may be a constant or something computed from the document. It is
// used wherever the field is null or missing, and needs no sort order, unlike [FillMethod].
func FillValue(name monq.FieldPath, value any) bson.D {
	return bson.D{{Key: string(name), Value: bson.D{{Key: "value", Value: value}}}}
}

// Redact returns a stage that keeps, prunes, or descends into each part of a document.
//
// $redact walks a document from the top down, evaluating the expression at every level, including inside embedded
// documents, and acts on what it returns: the "$$KEEP" variable takes that level and everything under it, "$$PRUNE"
// drops it, and "$$DESCEND" keeps the level but goes on to test the documents nested inside. That is how a
// per-document access rule gets applied at every depth, usually with a [github.com/behzadsh/monq/expr.Cond]
// choosing between the three.
//
// It works on embedded documents, not on array elements directly, and the check runs on every level, so it costs
// more than a [Match] that can use an index.
//
// Example:
//
//	stage.Redact(expr.Cond(expr.Eq(expr.Field("level"), "public"), "$$DESCEND", "$$PRUNE"))
//	// bson.D{{Key: "$redact", Value: bson.D{{Key: "$cond", Value: ...}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/redact/
func Redact(expression any) bson.D {
	return bson.D{{Key: "$redact", Value: expression}}
}
