package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestConditionalExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Cond emits the named form",
			got:  expr.Cond(expr.Gte(expr.Field("score"), 60), "pass", "fail"),
			want: bson.D{{Key: "$cond", Value: bson.D{
				{Key: "if", Value: bson.D{{Key: "$gte", Value: bson.A{"$score", 60}}}},
				{Key: "then", Value: "pass"},
				{Key: "else", Value: "fail"},
			}}},
		},
		{
			name: "IfNull with a single fallback",
			got:  expr.IfNull(expr.Field("nickname"), "anonymous"),
			want: bson.D{{Key: "$ifNull", Value: bson.A{"$nickname", "anonymous"}}},
		},
		{
			name: "IfNull with several candidates",
			got:  expr.IfNull(expr.Field("nickname"), expr.Field("name"), "anonymous"),
			want: bson.D{{Key: "$ifNull", Value: bson.A{"$nickname", "$name", "anonymous"}}},
		},
		{
			name: "Switch with branches and a default",
			got: expr.Switch(
				expr.Branch(expr.Gte(expr.Field("score"), 90), "A"),
				expr.Branch(expr.Gte(expr.Field("score"), 80), "B"),
				expr.DefaultCase("F"),
			),
			want: bson.D{{Key: "$switch", Value: bson.D{
				{Key: "branches", Value: bson.A{
					bson.D{
						{Key: "case", Value: bson.D{{Key: "$gte", Value: bson.A{"$score", 90}}}},
						{Key: "then", Value: "A"},
					},
					bson.D{
						{Key: "case", Value: bson.D{{Key: "$gte", Value: bson.A{"$score", 80}}}},
						{Key: "then", Value: "B"},
					},
				}},
				{Key: "default", Value: "F"},
			}}},
		},
		{
			name: "Switch without a default",
			got:  expr.Switch(expr.Branch(expr.Field("staff"), "internal")),
			want: bson.D{{Key: "$switch", Value: bson.D{
				{Key: "branches", Value: bson.A{
					bson.D{{Key: "case", Value: "$staff"}, {Key: "then", Value: "internal"}},
				}},
			}}},
		},
		{
			name: "Switch with no branches at all",
			got:  expr.Switch(),
			want: bson.D{{Key: "$switch", Value: bson.D{{Key: "branches", Value: bson.A{}}}}},
		},
		{
			name: "Switch with only a default",
			got:  expr.Switch(expr.DefaultCase("F")),
			want: bson.D{{Key: "$switch", Value: bson.D{
				{Key: "branches", Value: bson.A{}},
				{Key: "default", Value: "F"},
			}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func ExampleCond() {
	e := expr.Cond(expr.Gte(expr.Field("score"), 60), "pass", "fail")

	printExpr(e)
	// Output: {"$cond":{"if":{"$gte":["$score",60]},"then":"pass","else":"fail"}}
}

func ExampleIfNull() {
	e := expr.IfNull(expr.Field("nickname"), expr.Field("name"), "anonymous")

	printExpr(e)
	// Output: {"$ifNull":["$nickname","$name","anonymous"]}
}

func ExampleSwitch() {
	e := expr.Switch(
		expr.Branch(expr.Gte(expr.Field("score"), 90), "A"),
		expr.Branch(expr.Gte(expr.Field("score"), 80), "B"),
		expr.DefaultCase("F"),
	)

	printExpr(e)
	// Output: {"$switch":{"branches":[{"case":{"$gte":["$score",90]},"then":"A"},{"case":{"$gte":["$score",80]},"then":"B"}],"default":"F"}}
}

func ExampleBranch() {
	e := expr.Switch(expr.Branch(expr.Gte(expr.Field("score"), 90), "A"))

	printExpr(e)
	// Output: {"$switch":{"branches":[{"case":{"$gte":["$score",90]},"then":"A"}]}}
}

func ExampleDefaultCase() {
	e := expr.Switch(expr.Branch(expr.Gte(expr.Field("score"), 90), "A"), expr.DefaultCase("F"))

	printExpr(e)
	// Output: {"$switch":{"branches":[{"case":{"$gte":["$score",90]},"then":"A"}],"default":"F"}}
}
