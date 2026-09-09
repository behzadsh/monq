package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// AddToSet returns an update that appends value to the array at field unless it is already there.
//
// $addToSet appends only when no existing element equals value, which is what makes it the set-like counterpart of
// [Push]. Equality is full BSON equality, so documents match only when their fields have the same values in the
// same order. A missing field becomes a new array holding value, and a field that is not an array is a server
// error. To add several elements use [AddToSetEach]: passing a slice here appends the slice itself as one nested
// element. Combine it with other operators through [Update]; two operator documents concatenated by hand keep two
// separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.AddToSet("tags", "go")
//	// bson.D{{Key: "$addToSet", Value: bson.D{{Key: "tags", Value: "go"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/addToSet/
func AddToSet(field FieldPath, value any) bson.D {
	return bson.D{{Key: "$addToSet", Value: bson.D{{Key: string(field), Value: value}}}}
}

// AddToSetEach returns an update that appends each of values to the array at field, skipping duplicates.
//
// It is the $each form of $addToSet, which adds the elements one by one instead of adding the array as a single
// nested element. Every value is checked against the array separately, so some may be added while others are
// skipped. Values is a plain slice rather than a variadic parameter, matching [PushEach] and [PullAll], which
// keeps the whole family uniform and sidesteps the un-spread-slice gotcha that [In] and [All] carry. Combine it
// with other operators through [Update]; two operator documents concatenated by hand keep two separate keys, which
// MongoDB does not merge.
//
// Example:
//
//	monq.AddToSetEach("tags", []any{"go", "mongodb"})
//	// bson.D{
//	//     {
//	//         Key: "$addToSet",
//	//         Value: bson.D{
//	//             {Key: "tags", Value: bson.D{{Key: "$each", Value: bson.A{"go", "mongodb"}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/addToSet/#each-modifier
func AddToSetEach(field FieldPath, values []any) bson.D {
	each := bson.D{{Key: "$each", Value: toArray(values)}}

	return bson.D{{Key: "$addToSet", Value: bson.D{{Key: string(field), Value: each}}}}
}

// PopFirst returns an update that removes the first element of the array at field.
//
// $pop with -1 drops the leading element and shortens the array, unlike [Unset], which would leave a null behind.
// Popping from an empty array or a missing field does nothing; popping a field that is not an array is a server
// error. Combine it with other operators through [Update]; two operator documents concatenated by hand keep two
// separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.PopFirst("queue")
//	// bson.D{{Key: "$pop", Value: bson.D{{Key: "queue", Value: -1}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/pop/
func PopFirst(field FieldPath) bson.D {
	return bson.D{{Key: "$pop", Value: bson.D{{Key: string(field), Value: -1}}}}
}

// PopLast returns an update that removes the last element of the array at field.
//
// $pop with 1 drops the trailing element. It is a separate function from [PopFirst] because the two differ only by
// the operand -1 or 1, which a boolean parameter would hide at the call site. Combine it with other operators
// through [Update]; two operator documents concatenated by hand keep two separate keys, which MongoDB does not
// merge.
//
// Example:
//
//	monq.PopLast("history")
//	// bson.D{{Key: "$pop", Value: bson.D{{Key: "history", Value: 1}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/pop/
func PopLast(field FieldPath) bson.D {
	return bson.D{{Key: "$pop", Value: bson.D{{Key: string(field), Value: 1}}}}
}

// Pull returns an update that removes every element of the array at field matching condition.
//
// $pull deletes all matching elements at once and shortens the array. Condition is either a value to match exactly
// or a condition document, and which one to write depends on what the array holds. For arrays of documents a monq
// filter works directly: Pull("items", Eq("sku", "abc")) removes the elements whose sku is "abc". For arrays of
// scalars MongoDB wants a bare operator expression with no field name, which no monq comparison operator emits;
// build those with [Raw], e.g. Pull("scores", Raw(bson.D{{Key: "$gte", Value: 80}})). Pull takes a single
// condition, so several criteria go into one filter rather than several arguments. Combine it with other operators
// through [Update]; two operator documents concatenated by hand keep two separate keys, which MongoDB does not
// merge.
//
// Example:
//
//	monq.Pull("items", monq.Eq("sku", "abc"))
//	// bson.D{
//	//     {
//	//         Key: "$pull",
//	//         Value: bson.D{
//	//             {Key: "items", Value: bson.D{{Key: "sku", Value: bson.D{{Key: "$eq", Value: "abc"}}}}},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/pull/
func Pull(field FieldPath, condition any) bson.D {
	return bson.D{{Key: "$pull", Value: bson.D{{Key: string(field), Value: condition}}}}
}

// PullAll returns an update that removes every element of the array at field equal to one of values.
//
// $pullAll matches by exact value only, where [Pull] also accepts conditions, so it is the direct way to strip a
// known set of entries. Values is a plain slice rather than a variadic parameter, matching [PushEach] and
// [AddToSetEach]. Elements are compared with full BSON equality, so a nested array or document must match
// element for element. Combine it with other operators through [Update]; two operator documents concatenated by
// hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.PullAll("tags", []any{"draft", "wip"})
//	// bson.D{{Key: "$pullAll", Value: bson.D{{Key: "tags", Value: bson.A{"draft", "wip"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/pullAll/
func PullAll(field FieldPath, values []any) bson.D {
	return bson.D{{Key: "$pullAll", Value: bson.D{{Key: string(field), Value: toArray(values)}}}}
}

// Push returns an update that appends value to the array at field.
//
// $push appends unconditionally, duplicates included, and creates the array when the field is missing. Pushing a
// slice appends it as one nested element; to append several elements use [PushEach], which also carries the
// $position, $slice, and $sort modifiers. Combine it with other operators through [Update]; two operator documents
// concatenated by hand keep two separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.Push("tags", "go")
//	// bson.D{{Key: "$push", Value: bson.D{{Key: "tags", Value: "go"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/push/
func Push(field FieldPath, value any) bson.D {
	return bson.D{{Key: "$push", Value: bson.D{{Key: string(field), Value: value}}}}
}

// PushOption configures one modifier of a [PushEach] update.
type PushOption func(*bson.D)

// PushPosition returns a [PushOption] that inserts the new elements at the given index instead of appending them.
//
// A negative index counts back from the end of the array, and an index past the end appends. The option is named
// for its family rather than for the bare $position field, because the modifiers of $push are the only place these
// words mean anything.
func PushPosition(index int) PushOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$position", Value: index})
	}
}

// PushSlice returns a [PushOption] that trims the array to n elements after the new ones are added.
//
// A positive n keeps the first n elements, a negative n the last n, and 0 empties the array. Together with
// [PushSort] this is how a capped "latest N" list is maintained in one update.
func PushSlice(n int) PushOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$slice", Value: n})
	}
}

// PushSort returns a [PushOption] that sorts the array after the new elements are added.
//
// The spec is 1 or -1 for an array of scalars, or a sort document such as bson.D{{Key: "score", Value: -1}} for an
// array of documents. It is called PushSort rather than Sort because the root package keeps that name for building
// sort documents.
func PushSort(spec any) PushOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$sort", Value: spec})
	}
}

// PushEach returns an update that appends each of values to the array at field.
//
// It is the $each form of $push, which appends the elements one by one instead of appending the array as a single
// nested element, and the only form that carries the $position, $slice, and $sort modifiers ([PushPosition],
// [PushSlice], [PushSort]). MongoDB applies those in a fixed order, $position then $sort then $slice, whatever
// order they appear in the document, so a sorted and capped list keeps its top entries rather than the ones that
// happened to land first. Values is a plain slice rather than a variadic parameter, since the modifiers take that
// spot. Combine it with other operators through [Update]; two operator documents concatenated by hand keep two
// separate keys, which MongoDB does not merge.
//
// Example:
//
//	monq.PushEach("scores", []any{90, 80}, monq.PushSort(-1), monq.PushSlice(3))
//	// bson.D{
//	//     {
//	//         Key: "$push",
//	//         Value: bson.D{
//	//             {
//	//                 Key: "scores",
//	//                 Value: bson.D{
//	//                     {Key: "$each", Value: bson.A{90, 80}},
//	//                     {Key: "$sort", Value: -1},
//	//                     {Key: "$slice", Value: 3},
//	//                 },
//	//             },
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/push/#each-modifier
func PushEach(field FieldPath, values []any, opts ...PushOption) bson.D {
	modifiers := bson.D{{Key: "$each", Value: toArray(values)}}
	for _, opt := range opts {
		opt(&modifiers)
	}

	return bson.D{{Key: "$push", Value: bson.D{{Key: string(field), Value: modifiers}}}}
}

// toArray copies values into the bson.A the driver marshals as a BSON array.
func toArray(values []any) bson.A {
	arr := make(bson.A, len(values))
	copy(arr, values)

	return arr
}
