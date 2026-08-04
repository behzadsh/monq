package monq

import "strconv"

// ArrayPath is the path of an array field, and the starting point for naming positions inside it.
//
// MongoDB addresses array elements with a segment between the array's path and whatever comes after: a number for
// a known position, "$" for the element a query matched, "$[]" for every element, and "$[name]" for the elements an
// update's arrayFilters pick out. ArrayPath spells those four out as methods so the punctuation does not have to be
// remembered.
//
// The field itself is the array's own path, which is what operators such as [Size] and [All] take.
//
// Example:
//
//	tags := monq.ArrayPath{Path: "tags"}
//	monq.Size(tags.Path, 3)     // {"tags": {"$size": 3}}
//	monq.Set(tags.At(0), "go")  // {"$set": {"tags.0": "go"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update-array/
type ArrayPath struct {
	// Path is the array field itself.
	Path FieldPath
}

// All returns the path of every element of the array, the "$[]" all-positional operator.
//
// An update through this path applies to each element, which is how a field is changed across a whole array in one
// operation. It needs no query condition on the array, unlike [ArrayPath.Positional].
//
// Example:
//
//	monq.ArrayPath{Path: "items"}.All()
//	// monq.FieldPath("items.$[]")
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/positional-all/
func (a ArrayPath) All() FieldPath {
	return a.element("$[]")
}

// At returns the path of the element at index i, counting from 0.
//
// A fixed position is the one form that needs nothing from the query, since the index is written into the path
// itself. Reading or writing past the end of the array is not an error: MongoDB pads with nulls on a write.
//
// Example:
//
//	monq.ArrayPath{Path: "tags"}.At(0)
//	// monq.FieldPath("tags.0")
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/document/#array-elements
func (a ArrayPath) At(i int) FieldPath {
	return a.element(strconv.Itoa(i))
}

// Filtered returns the path of the elements an array filter selects, the "$[identifier]" operator.
//
// The identifier names a condition supplied with the update rather than here: the driver takes it through
// options.UpdateOne().SetArrayFilters(...), and the update fails if nothing there matches the name. That is what
// makes this the way to change only the elements meeting a condition of their own.
//
// Example:
//
//	monq.ArrayPath{Path: "items"}.Filtered("cheap")
//	// monq.FieldPath("items.$[cheap]")
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/positional-filtered/
func (a ArrayPath) Filtered(identifier string) FieldPath {
	return a.element("$[" + identifier + "]")
}

// Positional returns the path of the first element the query matched, the "$" positional operator.
//
// It only means something in an update whose filter contained a condition on this array, since "$" stands for
// whichever element that condition matched, and only the first match at that. An update using it without such a
// condition is an error, and [ArrayPath.All] or [ArrayPath.Filtered] is what covers more than one element.
//
// Example:
//
//	monq.ArrayPath{Path: "items"}.Positional()
//	// monq.FieldPath("items.$")
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/positional/
func (a ArrayPath) Positional() FieldPath {
	return a.element("$")
}

// element appends one segment to the array's path.
func (a ArrayPath) element(segment string) FieldPath {
	return FieldPath(string(a.Path) + "." + segment)
}
