package monq

// FieldPath is a document field name or dotted path, as accepted by MongoDB query and update operators (for example
// "status" or "stats.followers").
//
// It is a defined type rather than a plain string for two reasons. A signature then says which of its arguments
// names a field, which is most of what makes these functions readable. And it separates the parameters that address
// a path in the document from the ones that name a new output field, which stay plain strings: compare
// stage.Field(path, value), which may point at something that already exists, with stage.Count(name), which invents
// a field that does not.
//
// Untyped string literals and constants satisfy a FieldPath with no conversion, so handwritten calls stay short. A
// value already typed as string does need an explicit conversion, monq.FieldPath(s). Generated path constants are
// typed this way, so they pass wherever a path belongs and fail to compile where one does not.
//
// Nothing is enforced beyond that. A misspelled field name is still a valid FieldPath.
//
// No validation is performed on FieldPath values, including the empty string: an invalid or empty path is passed
// through untouched and rejected by MongoDB at query time, not by monq.
type FieldPath string
