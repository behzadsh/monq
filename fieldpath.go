package monq

// FieldPath is a document field name or dotted path, as accepted by MongoDB query and update operators (for example
// "status" or "stats.followers").
//
// FieldPath is defined as a string so untyped string literals and plain string variables convert to it implicitly at
// call sites, so hand-written code pays no conversion tax. A future codegen layer may generate typed FieldPath
// constants from struct bson tags, but that is not required to use monq.
//
// No validation is performed on FieldPath values, including the empty string: an invalid or empty path is passed
// through untouched and rejected by MongoDB at query time, not by monq.
type FieldPath string
