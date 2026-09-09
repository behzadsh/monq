package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// ArrayElemAt returns an expression yielding the element of an array at the given index.
//
// $arrayElemAt counts from 0, and a negative index counts back from the end, so -1 is the last element. An index
// past either end yields missing rather than null, which means a $project field built from it is left out of the
// document entirely.
//
// Example:
//
//	expr.ArrayElemAt(expr.Field("scores"), 0)
//	// bson.D{{Key: "$arrayElemAt", Value: bson.A{"$scores", 0}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/arrayElemAt/
func ArrayElemAt(array, idx any) bson.D {
	return bson.D{{Key: "$arrayElemAt", Value: bson.A{array, idx}}}
}

// ArrayToObject returns an expression turning an array of pairs into a document.
//
// $arrayToObject takes either an array of two-element arrays or an array of {k, v} documents, and is the inverse
// of [ObjectToArray]. Duplicate keys keep the last value. A key that is not a string, or a pair of the wrong
// shape, fails the aggregation.
//
// Example:
//
//	expr.ArrayToObject(expr.Field("settings"))
//	// bson.D{{Key: "$arrayToObject", Value: "$settings"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/arrayToObject/
func ArrayToObject(array any) bson.D {
	return bson.D{{Key: "$arrayToObject", Value: array}}
}

// ConcatArrays returns an expression joining arrays end to end.
//
// $concatArrays keeps duplicates and order, unlike [SetUnion], which treats its arguments as sets. A null or
// missing argument makes the whole result null.
//
// Example:
//
//	expr.ConcatArrays(expr.Field("tags"), expr.Field("extra_tags"))
//	// bson.D{{Key: "$concatArrays", Value: bson.A{"$tags", "$extra_tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/concatArrays/
func ConcatArrays(arrays ...any) bson.D {
	return bson.D{{Key: "$concatArrays", Value: operands(arrays)}}
}

// Filter returns an expression keeping the elements of an array that satisfy a condition.
//
// $filter walks the input array and keeps the elements for which cond is true, preserving their order. The
// condition refers to the element under test as the variable "$$this", not as a field: the two dollar signs are
// what tell a variable from a field reference. [FilterAs] renames it, and [FilterLimit] stops after a number of
// matches.
//
// Example:
//
//	expr.Filter(expr.Field("scores"), expr.Gte("$$this", 80))
//	// bson.D{
//	//     {
//	//         Key: "$filter",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$scores"},
//	//             {Key: "cond", Value: bson.D{{Key: "$gte", Value: bson.A{"$$this", 80}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/filter/
func Filter(input, cond any, opts ...FilterOption) bson.D {
	spec := bson.D{{Key: "input", Value: input}, {Key: "cond", Value: cond}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$filter", Value: spec}}
}

// FilterAs returns a [FilterOption] renaming the variable that holds the element under test.
//
// Without it the element is "$$this". Naming it "score" makes it "$$score" inside the condition, which is worth
// doing when one [Filter] sits inside another and two elements are in scope at once.
func FilterAs(name string) FilterOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "as", Value: name})
	}
}

// FilterLimit returns a [FilterOption] capping how many elements come back.
//
// The filter stops once that many elements have matched, so it is the cheap way to ask whether a few matches
// exist. A limit of 0 or less fails the aggregation.
func FilterLimit(n any) FilterOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "limit", Value: n})
	}
}

// FilterOption configures the optional fields of a [Filter] expression.
type FilterOption func(*bson.D)

// First returns an expression yielding the first element of an array.
//
// $first is a shorthand for [ArrayElemAt] with an index of 0, and yields missing for an empty array. The same
// operator is the accumulator that takes the first document of a group, with the same shape, so this one function
// covers both uses: inside a [github.com/behzadsh/monq/stage.Group] it accumulates, elsewhere it reads an array.
//
// Example:
//
//	expr.First(expr.Field("scores"))
//	// bson.D{{Key: "$first", Value: "$scores"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/first/
func First(array any) bson.D {
	return bson.D{{Key: "$first", Value: array}}
}

// FirstN returns an expression yielding the first n elements of an array.
//
// $firstN returns fewer elements than asked for rather than padding when the array is shorter, and an n of 0 or
// less fails the aggregation. Like [First] it doubles as an accumulator with the same shape, collecting from the
// first n documents of a group.
//
// Example:
//
//	expr.FirstN(expr.Field("scores"), 3)
//	// bson.D{{Key: "$firstN", Value: bson.D{{Key: "input", Value: "$scores"}, {Key: "n", Value: 3}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/firstN/
func FirstN(input, n any) bson.D {
	return bson.D{{Key: "$firstN", Value: inputN(input, n)}}
}

// In returns an expression that is true when a value is an element of an array.
//
// $in compares with full BSON equality and takes exactly two arguments, the value first and the array second.
// The query operator of the same name works the other way around: monq.In names a field and is variadic over the
// candidate values, while this one compares two expressions.
//
// Example:
//
//	expr.In("draft", expr.Field("tags"))
//	// bson.D{{Key: "$in", Value: bson.A{"draft", "$tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/in/
func In(value, array any) bson.D {
	return bson.D{{Key: "$in", Value: bson.A{value, array}}}
}

// IndexOfArray returns an expression yielding the position of a value within an array.
//
// $indexOfArray compares with full BSON equality and yields -1 when the value is absent, or null when the array
// argument is null or missing. The optional trailing arguments narrow the search the same way they do in
// [IndexOfBytes]: none searches the whole array, one gives a start index, two give a start and an end.
//
// Example:
//
//	expr.IndexOfArray(expr.Field("tags"), "draft")
//	// bson.D{{Key: "$indexOfArray", Value: bson.A{"$tags", "draft"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/indexOfArray/
func IndexOfArray(array, value any, startEnd ...any) bson.D {
	return bson.D{{Key: "$indexOfArray", Value: indexOfArgs(array, value, startEnd)}}
}

// IsArray returns an expression that is true when a value is an array.
//
// $isArray takes its argument wrapped in an array, which is a quirk of the operator rather than a choice here, and
// yields a boolean for any input at all, including null and missing.
//
// Example:
//
//	expr.IsArray(expr.Field("tags"))
//	// bson.D{{Key: "$isArray", Value: bson.A{"$tags"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isArray/
func IsArray(value any) bson.D {
	return bson.D{{Key: "$isArray", Value: bson.A{value}}}
}

// Last returns an expression yielding the last element of an array.
//
// $last is a shorthand for [ArrayElemAt] with an index of -1, and yields missing for an empty array. Like [First]
// it is also the accumulator of the same name, taking the last document of a group.
//
// Example:
//
//	expr.Last(expr.Field("scores"))
//	// bson.D{{Key: "$last", Value: "$scores"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/last/
func Last(array any) bson.D {
	return bson.D{{Key: "$last", Value: array}}
}

// LastN returns an expression yielding the last n elements of an array.
//
// $lastN returns fewer elements than asked for when the array is shorter, and doubles as the accumulator that
// collects from the last n documents of a group.
//
// Example:
//
//	expr.LastN(expr.Field("scores"), 3)
//	// bson.D{{Key: "$lastN", Value: bson.D{{Key: "input", Value: "$scores"}, {Key: "n", Value: 3}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/lastN/
func LastN(input, n any) bson.D {
	return bson.D{{Key: "$lastN", Value: inputN(input, n)}}
}

// Map returns an expression applying a sub-expression to every element of an array.
//
// $map produces an array of the same length, with each element replaced by what in evaluates to. The element is
// bound to the variable "$$this" unless [MapAs] renames it, and again the two dollar signs mark a variable rather
// than a field.
//
// Example:
//
//	expr.Map(expr.Field("prices"), expr.Multiply("$$this", 1.1))
//	// bson.D{
//	//     {
//	//         Key: "$map",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$prices"},
//	//             {Key: "in", Value: bson.D{{Key: "$multiply", Value: bson.A{"$$this", 1.1}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/map/
func Map(input, in any, opts ...MapOption) bson.D {
	spec := make(bson.D, 0, len(opts)+2)
	spec = append(spec, bson.E{Key: "input", Value: input})

	for _, opt := range opts {
		opt(&spec)
	}

	spec = append(spec, bson.E{Key: "in", Value: in})

	return bson.D{{Key: "$map", Value: spec}}
}

// MapAs returns a [MapOption] renaming the variable that holds the current element.
//
// Without it the element is "$$this". A name of its own keeps things readable when one [Map] is nested inside
// another.
func MapAs(name string) MapOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "as", Value: name})
	}
}

// MapOption configures the optional fields of a [Map] expression.
type MapOption func(*bson.D)

// MaxN returns an expression yielding the n largest elements of an array.
//
// $maxN sorts by BSON comparison order, skips null and missing elements, and returns fewer than n when the array
// holds fewer. It is also the accumulator that takes the n largest values across a group.
//
// Example:
//
//	expr.MaxN(expr.Field("scores"), 3)
//	// bson.D{{Key: "$maxN", Value: bson.D{{Key: "input", Value: "$scores"}, {Key: "n", Value: 3}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/maxN/
func MaxN(input, n any) bson.D {
	return bson.D{{Key: "$maxN", Value: inputN(input, n)}}
}

// MinN returns an expression yielding the n smallest elements of an array.
//
// $minN is the counterpart of [MaxN], with the same handling of nulls and short arrays, and likewise doubles as an
// accumulator.
//
// Example:
//
//	expr.MinN(expr.Field("scores"), 3)
//	// bson.D{{Key: "$minN", Value: bson.D{{Key: "input", Value: "$scores"}, {Key: "n", Value: 3}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/minN/
func MinN(input, n any) bson.D {
	return bson.D{{Key: "$minN", Value: inputN(input, n)}}
}

// ObjectToArray returns an expression turning a document into an array of {k, v} pairs.
//
// $objectToArray is the inverse of [ArrayToObject] and keeps the document's field order. It is how a pipeline
// works over field names it does not know in advance, usually with a [Filter] or [Map] over the result.
//
// Example:
//
//	expr.ObjectToArray(expr.Field("settings"))
//	// bson.D{{Key: "$objectToArray", Value: "$settings"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/objectToArray/
func ObjectToArray(object any) bson.D {
	return bson.D{{Key: "$objectToArray", Value: object}}
}

// Range returns an expression building an array of numbers.
//
// $range counts from start up to but not including end. The optional third argument is the step, which defaults to
// 1 and may be negative to count down; a step of 0 fails the aggregation, and a range that never reaches its end
// yields an empty array. All three arguments have to be integers.
//
// Example:
//
//	expr.Range(0, 10, 2)
//	// bson.D{{Key: "$range", Value: bson.A{0, 10, 2}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/range/
func Range(start, end any, step ...any) bson.D {
	args := make(bson.A, 0, len(step)+2)
	args = append(args, start, end)

	return bson.D{{Key: "$range", Value: append(args, step...)}}
}

// Reduce returns an expression folding an array down to a single value.
//
// $reduce walks the input array once, and the in expression sees two variables: "$$value", the result so far,
// starting at initialValue, and "$$this", the current element. Whatever in evaluates to becomes the next
// "$$value". An empty array yields initialValue untouched, and a null input yields null.
//
// Example:
//
//	expr.Reduce(expr.Field("scores"), 0, expr.Add("$$value", "$$this"))
//	// bson.D{
//	//     {
//	//         Key: "$reduce",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$scores"},
//	//             {Key: "initialValue", Value: 0},
//	//             {Key: "in", Value: bson.D{{Key: "$add", Value: bson.A{"$$value", "$$this"}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/reduce/
func Reduce(input, initialValue, in any) bson.D {
	return bson.D{
		{
			Key: "$reduce",
			Value: bson.D{
				{Key: "input", Value: input},
				{Key: "initialValue", Value: initialValue},
				{Key: "in", Value: in},
			},
		},
	}
}

// ReverseArray returns an expression yielding an array in reverse order.
//
// $reverseArray leaves the elements themselves alone, so a nested array is not reversed in turn. A null or missing
// input yields null.
//
// Example:
//
//	expr.ReverseArray(expr.Field("history"))
//	// bson.D{{Key: "$reverseArray", Value: "$history"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/reverseArray/
func ReverseArray(array any) bson.D {
	return bson.D{{Key: "$reverseArray", Value: array}}
}

// Size returns an expression yielding the number of elements in an array.
//
// $size computes a length, where the query operator of the same name tests one: monq.Size(field, size) matches
// documents whose array has exactly that many elements. This takes an array expression and gives back a number,
// and fails the aggregation on anything that is not an array, missing fields included.
//
// Example:
//
//	expr.Size(expr.Field("tags"))
//	// bson.D{{Key: "$size", Value: "$tags"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/size/
func Size(array any) bson.D {
	return bson.D{{Key: "$size", Value: array}}
}

// Slice returns an expression taking the first or last n elements of an array.
//
// $slice with two arguments counts from the start for a positive n and from the end for a negative one, so -3 is
// the last three elements. [SliceFrom] is the three-argument form, which starts at a position instead; they are
// separate functions because the second argument means something different in each.
//
// Example:
//
//	expr.Slice(expr.Field("history"), 5)
//	// bson.D{{Key: "$slice", Value: bson.A{"$history", 5}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/slice/
func Slice(array, n any) bson.D {
	return bson.D{{Key: "$slice", Value: bson.A{array, n}}}
}

// SliceFrom returns an expression taking n elements of an array starting at a position.
//
// It is the three-argument $slice. The position counts from 0, or from the end when negative, and n has to be
// positive here, unlike in [Slice]. A position past the end yields an empty array.
//
// Example:
//
//	expr.SliceFrom(expr.Field("history"), 10, 5)
//	// bson.D{{Key: "$slice", Value: bson.A{"$history", 10, 5}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/slice/
func SliceFrom(array, position, n any) bson.D {
	return bson.D{{Key: "$slice", Value: bson.A{array, position, n}}}
}

// SortArray returns an expression sorting the elements of an array.
//
// $sortArray takes a sort document of field-to-direction pairs for arrays of documents, 1 for ascending and -1 for
// descending, or a bare 1 or -1 for an array of scalars. Sorting is stable and uses BSON comparison order across
// types.
//
// Example:
//
//	expr.SortArray(expr.Field("items"), bson.D{{Key: "price", Value: -1}})
//	// bson.D{
//	//     {
//	//         Key: "$sortArray",
//	//         Value: bson.D{
//	//             {Key: "input", Value: "$items"},
//	//             {Key: "sortBy", Value: bson.D{{Key: "price", Value: -1}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sortArray/
func SortArray(input, sortBy any) bson.D {
	return bson.D{
		{
			Key: "$sortArray", Value: bson.D{
				{Key: "input", Value: input},
				{Key: "sortBy", Value: sortBy},
			},
		},
	}
}

// Zip returns an expression stitching several arrays together element by element.
//
// $zip pairs up the first elements, then the second, and so on, producing an array of arrays. It stops at the
// shortest input unless [ZipUseLongestLength] says otherwise, in which case the gaps are filled with null or with
// what [ZipDefaults] provides.
//
// Example:
//
//	expr.Zip([]any{expr.Field("names"), expr.Field("scores")})
//	// bson.D{
//	//     {
//	//         Key: "$zip",
//	//         Value: bson.D{
//	//             {Key: "inputs", Value: bson.A{"$names", "$scores"}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/zip/
func Zip(inputs []any, opts ...ZipOption) bson.D {
	spec := bson.D{{Key: "inputs", Value: operands(inputs)}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$zip", Value: spec}}
}

// ZipDefaults returns a [ZipOption] giving the values that fill in for missing elements.
//
// The defaults are an array with one entry per input, used when [ZipUseLongestLength] is on and an input has run
// out. Without them the gaps are null, and using this option without ZipUseLongestLength fails the aggregation.
func ZipDefaults(defaults any) ZipOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "defaults", Value: defaults})
	}
}

// ZipOption configures the optional fields of a [Zip] expression.
type ZipOption func(*bson.D)

// ZipUseLongestLength returns a [ZipOption] that keeps zipping to the length of the longest input.
//
// Without it the result stops at the shortest input, dropping whatever the longer ones still held.
func ZipUseLongestLength() ZipOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "useLongestLength", Value: true})
	}
}

// inputN builds the {input, n} document the $firstN family of operators takes.
func inputN(input, n any) bson.D {
	return bson.D{{Key: "input", Value: input}, {Key: "n", Value: n}}
}
