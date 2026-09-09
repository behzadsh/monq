package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// Cond returns an expression that evaluates one of two branches depending on a condition.
//
// $cond is the ternary of the aggregation language: when ifExpr is true the result is thenExpr, otherwise elseExpr.
// Both branches have to be given, and only the taken one is evaluated. MongoDB accepts an array form as well,
// [if, then, else]; Cond emits the named form, which says which argument is which when the query is printed.
//
// Example:
//
//	expr.Cond(expr.Gte(expr.Field("score"), 60), "pass", "fail")
//	// bson.D{
//	//     {
//	//         Key: "$cond",
//	//         Value: bson.D{
//	//             {Key: "if", Value: bson.D{{Key: "$gte", Value: bson.A{"$score", 60}}}},
//	//             {Key: "then", Value: "pass"},
//	//             {Key: "else", Value: "fail"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/cond/
func Cond(ifExpr, thenExpr, elseExpr any) bson.D {
	return bson.D{
		{
			Key: "$cond",
			Value: bson.D{
				{Key: "if", Value: ifExpr},
				{Key: "then", Value: thenExpr},
				{Key: "else", Value: elseExpr},
			},
		},
	}
}

// IfNull returns the first of expressions that is neither null nor missing.
//
// $ifNull walks its arguments in order and returns the first one that resolves to something, which makes it the
// way to supply a default for an optional field. The last argument is the fallback and is returned as is even if
// it is null. MongoDB accepted exactly two arguments before 5.0 and any number from 5.0 on.
//
// Example:
//
//	expr.IfNull(expr.Field("nickname"), expr.Field("name"), "anonymous")
//	// bson.D{{Key: "$ifNull", Value: bson.A{"$nickname", "$name", "anonymous"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ifNull/
func IfNull(expressions ...any) bson.D {
	return bson.D{{Key: "$ifNull", Value: operands(expressions)}}
}

// SwitchOption adds a branch or the default result to a [Switch] expression.
type SwitchOption func(*bson.D)

// Branch returns a [SwitchOption] holding one case of a [Switch] and the value it produces.
//
// Branches are tested in the order they are given, and the first case that evaluates to true wins. A case that is
// not a boolean follows the usual rule: false, null, 0, and missing count as false.
func Branch(caseExpr, thenExpr any) SwitchOption {
	return func(d *bson.D) {
		branch := bson.D{{Key: "case", Value: caseExpr}, {Key: "then", Value: thenExpr}}

		i := indexOfKey(*d, "branches")
		if i < 0 {
			return
		}

		branches, ok := (*d)[i].Value.(bson.A)
		if !ok {
			return
		}

		(*d)[i].Value = append(branches, branch)
	}
}

// DefaultCase returns a [SwitchOption] holding the value a [Switch] produces when no branch matches.
//
// Without it, a document that matches no branch makes the whole aggregation fail rather than yielding null, which
// is the usual surprise with $switch.
func DefaultCase(value any) SwitchOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "default", Value: value})
	}
}

// Switch returns an expression that picks the value of the first matching branch.
//
// $switch is the chain of else-ifs the aggregation language gives in place of nesting [Cond] inside itself.
// Branches come from [Branch] and are tested in order; [DefaultCase] gives the value for a document that matches
// none of them, and without it such a document fails the aggregation. A $switch with no branches at all is an
// error the server reports, not something checked here.
//
// Example:
//
//	expr.Switch(
//		expr.Branch(expr.Gte(expr.Field("score"), 90), "A"),
//		expr.Branch(expr.Gte(expr.Field("score"), 80), "B"),
//		expr.DefaultCase("F"),
//	)
//	// bson.D{
//	//     {
//	//         Key: "$switch",
//	//         Value: bson.D{
//	//             {
//	//                 Key: "branches",
//	//                 Value: bson.A{
//	//                     bson.D{{Key: "case", Value: ...}, {Key: "then", Value: "A"}},
//	//                     bson.D{{Key: "case", Value: ...}, {Key: "then", Value: "B"}},
//	//                 },
//	//             },
//	//             {Key: "default", Value: "F"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/switch/
func Switch(opts ...SwitchOption) bson.D {
	spec := bson.D{{Key: "branches", Value: bson.A{}}}
	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$switch", Value: spec}}
}

// indexOfKey returns the position of key in d, or -1 when d has no such key.
func indexOfKey(d bson.D, key string) int {
	for i, e := range d {
		if e.Key == key {
			return i
		}
	}

	return -1
}
