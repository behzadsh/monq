package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// AddToSet returns an accumulator collecting the distinct values of a group into an array.
//
// $addToSet skips a value already collected, comparing with full BSON equality, and the order of the result is not
// meaningful. [Push] is the one that keeps duplicates. The update operator of the same name adds to an array of a
// stored document: monq.AddToSet(field, value) writes, this one reads across a group.
//
// Example:
//
//	expr.AddToSet(expr.Field("category"))
//	// bson.D{{Key: "$addToSet", Value: "$category"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/addToSet/
func AddToSet(value any) bson.D {
	return bson.D{{Key: "$addToSet", Value: value}}
}

// Avg returns the average of its arguments, or of one expression across a group.
//
// With a single argument $avg is an accumulator averaging that expression over every document of a group; with
// several it averages them within each document. Non-numeric values are skipped rather than counted as 0, and a
// group with no numeric values at all yields null.
//
// Example:
//
//	expr.Avg(expr.Field("score"))
//	// bson.D{{Key: "$avg", Value: "$score"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/avg/
func Avg(values ...any) bson.D {
	return bson.D{{Key: "$avg", Value: accumulatorOperand(values)}}
}

// Bottom returns an accumulator taking the output of the last document of a group in a given order.
//
// $bottom sorts the group by sortBy and keeps output from the final document, which is how the worst or newest
// entry of each group comes back alongside its aggregates. [Top] takes the first instead, and [BottomN] takes
// several. The sort document is the same shape a $sort stage takes.
//
// Example:
//
//	expr.Bottom(bson.D{{Key: "score", Value: 1}}, expr.Field("name"))
//	// bson.D{{Key: "$bottom", Value: bson.D{
//	//     {Key: "sortBy", Value: bson.D{{Key: "score", Value: 1}}},
//	//     {Key: "output", Value: "$name"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bottom/
func Bottom(sortBy, output any) bson.D {
	return bson.D{{Key: "$bottom", Value: topBottomSpec(sortBy, output)}}
}

// BottomN returns an accumulator taking the output of the last n documents of a group in a given order.
//
// $bottomN is [Bottom] over more than one document, returning an array. A group holding fewer than n documents
// yields all of them rather than padding.
//
// Example:
//
//	expr.BottomN(3, bson.D{{Key: "score", Value: 1}}, expr.Field("name"))
//	// bson.D{{Key: "$bottomN", Value: bson.D{
//	//     {Key: "n", Value: 3},
//	//     {Key: "sortBy", Value: bson.D{{Key: "score", Value: 1}}},
//	//     {Key: "output", Value: "$name"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/bottomN/
func BottomN(n, sortBy, output any) bson.D {
	spec := append(bson.D{{Key: "n", Value: n}}, topBottomSpec(sortBy, output)...)

	return bson.D{{Key: "$bottomN", Value: spec}}
}

// Count returns an accumulator counting the documents in a group.
//
// $count as an accumulator takes an empty document and belongs inside a $group or $setWindowFields, where it is
// shorter than summing 1 per document. It is not the pipeline stage of the same name: stage.Count(field) is a
// whole stage that replaces the input with a single counting document, and using this in its place is a server
// error.
//
// Example:
//
//	expr.Count()
//	// bson.D{{Key: "$count", Value: bson.D{}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/count-accumulator/
func Count() bson.D {
	return bson.D{{Key: "$count", Value: bson.D{}}}
}

// Max returns the largest of its arguments, or of one expression across a group.
//
// With a single argument $max is an accumulator taking the maximum over a group; with several it takes the
// maximum within each document. Comparison follows BSON order across types, and null and missing values are
// skipped. The update operator of the same name raises a stored field: monq.Max(field, value) writes, this one
// reads. [MaxN] takes more than one value.
//
// Example:
//
//	expr.Max(expr.Field("score"))
//	// bson.D{{Key: "$max", Value: "$score"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/max/
func Max(values ...any) bson.D {
	return bson.D{{Key: "$max", Value: accumulatorOperand(values)}}
}

// Min returns the smallest of its arguments, or of one expression across a group.
//
// With a single argument $min is an accumulator taking the minimum over a group; with several it takes the minimum
// within each document. It is the counterpart of [Max], with the same BSON comparison order and the same skipping
// of null and missing values, and it likewise shares a name with the update operator monq.Min, which lowers a
// stored field rather than reading one. [MinN] takes more than one value.
//
// Example:
//
//	expr.Min(expr.Field("score"))
//	// bson.D{{Key: "$min", Value: "$score"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/min/
func Min(values ...any) bson.D {
	return bson.D{{Key: "$min", Value: accumulatorOperand(values)}}
}

// Median returns an accumulator yielding the median of a numeric expression.
//
// $median is [Percentile] at 0.5 with a shorter spelling. The method has to be given and "approximate" is the only
// value MongoDB accepts, which is why monq asks for it rather than filling it in: an operator that takes a
// required field takes it here too.
//
// Example:
//
//	expr.Median(expr.Field("score"), "approximate")
//	// bson.D{{Key: "$median", Value: bson.D{
//	//     {Key: "input", Value: "$score"},
//	//     {Key: "method", Value: "approximate"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/median/
func Median(input, method any) bson.D {
	return bson.D{{Key: "$median", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "method", Value: method},
	}}}
}

// Percentile returns an accumulator yielding one or more percentiles of a numeric expression.
//
// $percentile takes p as an array of values between 0 and 1 and returns an array of the same length, so asking for
// the median and the 95th percentile at once is one operator. The method has to be given, with "approximate" the
// only value MongoDB accepts today.
//
// Example:
//
//	expr.Percentile(expr.Field("score"), bson.A{0.5, 0.95}, "approximate")
//	// bson.D{{Key: "$percentile", Value: bson.D{
//	//     {Key: "input", Value: "$score"},
//	//     {Key: "p", Value: bson.A{0.5, 0.95}},
//	//     {Key: "method", Value: "approximate"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/percentile/
func Percentile(input, p, method any) bson.D {
	return bson.D{{Key: "$percentile", Value: bson.D{
		{Key: "input", Value: input},
		{Key: "p", Value: p},
		{Key: "method", Value: method},
	}}}
}

// Push returns an accumulator collecting a value from every document of a group into an array.
//
// $push keeps duplicates and the order the documents arrived in, which is what separates it from [AddToSet]. A
// group large enough to exceed the aggregation's memory limit fails unless the pipeline may use disk. The update
// operator of the same name appends to a stored array: monq.Push(field, value) writes, this one reads.
//
// Example:
//
//	expr.Push(expr.Field("name"))
//	// bson.D{{Key: "$push", Value: "$name"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/push/
func Push(value any) bson.D {
	return bson.D{{Key: "$push", Value: value}}
}

// StdDevPop returns the population standard deviation of its arguments, or of one expression across a group.
//
// $stdDevPop treats the values as the whole population rather than a sample, which is the right choice when the
// group is everything there is. [StdDevSamp] is the sample version. Non-numeric values are skipped, and a single
// value yields 0.
//
// Example:
//
//	expr.StdDevPop(expr.Field("score"))
//	// bson.D{{Key: "$stdDevPop", Value: "$score"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/stdDevPop/
func StdDevPop(values ...any) bson.D {
	return bson.D{{Key: "$stdDevPop", Value: accumulatorOperand(values)}}
}

// StdDevSamp returns the sample standard deviation of its arguments, or of one expression across a group.
//
// $stdDevSamp treats the values as a sample of something larger, dividing by one less than their count. Fewer than
// two values yield null, where [StdDevPop] would yield 0.
//
// Example:
//
//	expr.StdDevSamp(expr.Field("score"))
//	// bson.D{{Key: "$stdDevSamp", Value: "$score"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/stdDevSamp/
func StdDevSamp(values ...any) bson.D {
	return bson.D{{Key: "$stdDevSamp", Value: accumulatorOperand(values)}}
}

// Sum returns the total of its arguments, or of one expression across a group.
//
// The argument count decides which of two different jobs this does. With one argument $sum is an accumulator
// adding that expression up over every document of a group, so Sum(Field("amount")) is a group total. With several
// it adds them together within each document, so Sum(Field("price"), Field("tax")) is a per-document total. Both
// are valid and they are not interchangeable.
//
// Non-numeric values are skipped rather than failing, and a group with none at all yields 0.
//
// Example:
//
//	expr.Sum(expr.Field("amount"))
//	// bson.D{{Key: "$sum", Value: "$amount"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sum/
func Sum(values ...any) bson.D {
	return bson.D{{Key: "$sum", Value: accumulatorOperand(values)}}
}

// Top returns an accumulator taking the output of the first document of a group in a given order.
//
// $top sorts the group by sortBy and keeps output from the leading document, which is how the best or newest entry
// of each group comes back alongside its aggregates. [Bottom] takes the last instead, and [TopN] takes several.
//
// Example:
//
//	expr.Top(bson.D{{Key: "score", Value: -1}}, expr.Field("name"))
//	// bson.D{{Key: "$top", Value: bson.D{
//	//     {Key: "sortBy", Value: bson.D{{Key: "score", Value: -1}}},
//	//     {Key: "output", Value: "$name"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/top/
func Top(sortBy, output any) bson.D {
	return bson.D{{Key: "$top", Value: topBottomSpec(sortBy, output)}}
}

// TopN returns an accumulator taking the output of the first n documents of a group in a given order.
//
// $topN is [Top] over more than one document, returning an array. A group holding fewer than n documents yields
// all of them.
//
// Example:
//
//	expr.TopN(3, bson.D{{Key: "score", Value: -1}}, expr.Field("name"))
//	// bson.D{{Key: "$topN", Value: bson.D{
//	//     {Key: "n", Value: 3},
//	//     {Key: "sortBy", Value: bson.D{{Key: "score", Value: -1}}},
//	//     {Key: "output", Value: "$name"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/topN/
func TopN(n, sortBy, output any) bson.D {
	spec := append(bson.D{{Key: "n", Value: n}}, topBottomSpec(sortBy, output)...)

	return bson.D{{Key: "$topN", Value: spec}}
}

// accumulatorOperand renders one argument on its own, the accumulator form these operators take inside a group,
// and several as an array, the expression form that works within a single document.
func accumulatorOperand(values []any) any {
	if len(values) == 1 {
		return values[0]
	}

	return operands(values)
}

// topBottomSpec builds the {sortBy, output} pair the $top and $bottom operators share.
func topBottomSpec(sortBy, output any) bson.D {
	return bson.D{{Key: "sortBy", Value: sortBy}, {Key: "output", Value: output}}
}
