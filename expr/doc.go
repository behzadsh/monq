// Package expr builds MongoDB aggregation expressions.
//
// Aggregation expressions are what stages compute with: the value of a $group accumulator, the condition of a
// $match inside a $lookup, the fields a $project derives. Each function here is named after the operator it builds
// and returns a raw bson.D, the same as everything else in monq.
//
//	stage.Project(
//		stage.Field("name", 1),
//		stage.Field("is_adult", expr.Gte(expr.Field("age"), 18)),
//	)
//
// Expression operators compare values, not fields. That is the one thing to know before writing any of them:
// monq.Eq("status", "active") names a field and a value, while expr.Eq(a, b) takes two expressions, so
// expr.Eq("status", "active") compares two constant strings and is false for every document. The field reference
// is a string that starts with a dollar sign, "$status", which [Field] builds from a monq.FieldPath.
//
// The flip side is that any string beginning with a dollar sign is read as a field reference, so a literal string
// that starts with one has to go through [Literal]. Variables are different again and take two dollar signs:
// "$$new" and other names defined by $lookup's let, and the ones MongoDB provides such as "$$ROOT" and "$$this".
//
// Names in this package deliberately repeat names in the root package and in monq/stage, because MongoDB reuses
// them: expr.Eq is $eq the expression operator, monq.Eq is $eq the query operator, and the package qualifier is
// what tells them apart at a call site.
package expr
