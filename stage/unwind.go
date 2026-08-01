package stage

import "go.mongodb.org/mongo-driver/v2/bson"

// UnwindOption configures one optional field of an [Unwind] stage.
type UnwindOption func(*bson.D)

// IncludeArrayIndex returns an [UnwindOption] that records each element's position in a new field.
//
// The name is an output field name, not a path expression, and the field holds the element's index as a long. For
// documents that came from an empty or missing array, kept only when [PreserveNullAndEmptyArrays] is on, the field
// is null.
func IncludeArrayIndex(name string) UnwindOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "includeArrayIndex", Value: name})
	}
}

// PreserveNullAndEmptyArrays returns an [UnwindOption] that keeps documents with nothing to unwind.
//
// By default $unwind drops a document whose array field is missing, null, or empty. With this option the document
// passes through once, with the field left as it was. It is what turns an $unwind after a $lookup from an inner
// join into a left outer join.
func PreserveNullAndEmptyArrays() UnwindOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "preserveNullAndEmptyArrays", Value: true})
	}
}

// Unwind returns a stage that turns each element of an array field into its own document.
//
// $unwind emits one copy of the document per array element, with the array field replaced by that element, so a
// document with a three-element array becomes three documents. Documents whose field is missing, null, or an empty
// array are dropped unless [PreserveNullAndEmptyArrays] says otherwise, and a field that is not an array is passed
// through as if it were a single-element one.
//
// The path is an aggregation field reference and needs its "$" prefix, "$items" rather than "items", the same as
// any other expression in a pipeline. MongoDB also accepts a shorthand where the whole stage is that string, but
// Unwind always emits the document form so that a call with options and one without have the same shape.
//
// Example:
//
//	stage.Unwind("$items", stage.PreserveNullAndEmptyArrays())
//	// bson.D{{Key: "$unwind", Value: bson.D{
//	//     {Key: "path", Value: "$items"},
//	//     {Key: "preserveNullAndEmptyArrays", Value: true},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/unwind/
func Unwind(path string, opts ...UnwindOption) bson.D {
	spec := bson.D{{Key: "path", Value: path}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$unwind", Value: spec}}
}
