// Package stage builds MongoDB aggregation pipeline stages.
//
// Each function is named after the stage it builds and returns a raw bson.D, exactly as the filter and update
// operators in the root monq package do. Stages go into a pipeline in the order they run:
//
//	pipeline := stage.Pipeline(
//		stage.Match(monq.Eq("status", "active")),
//		stage.Sort(bson.D{{Key: "created_at", Value: -1}}),
//		stage.Limit(20),
//	)
//
//	cursor, err := collection.Aggregate(ctx, pipeline)
//
// Some stage names are also operator names, which is why they live in their own package: stage.Set is the $set
// stage that adds fields to documents, while monq.Set is the $set update operator. The package qualifier says
// which one is meant, and neither name has to be bent out of shape.
//
// Aggregation expressions inside stages refer to fields with the "$field" string form rather than a bare field
// name, since that is how the aggregation framework tells a field reference from a literal string. Parameter types
// follow that split: a parameter naming a path into a document is a monq.FieldPath ([Lookup]'s localField,
// [Unset]'s fields), while an output field name ([Count], [Accumulator]) or an expression that happens to be a
// string ([Unwind]'s path, [Group]'s id) is a plain string. Nothing here adds or strips a "$".
//
// monq does not track what a pipeline does to the shape of its documents. A stage that renames or regroups fields
// is described in its doc comment, not in the type system, which is a deliberate limit: pipelines reshape data far
// too freely for a Go type to follow along.
package stage
