package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestMiscExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Let binds a variable for one expression",
			got: expr.Let(
				bson.D{{Key: "total", Value: expr.Add(expr.Field("price"), expr.Field("tax"))}},
				expr.Multiply("$$total", 0.9),
			),
			want: bson.D{
				{
					Key: "$let",
					Value: bson.D{
						{
							Key: "vars",
							Value: bson.D{
								{Key: "total", Value: bson.D{{Key: "$add", Value: bson.A{"$price", "$tax"}}}},
							},
						},
						{Key: "in", Value: bson.D{{Key: "$multiply", Value: bson.A{"$$total", 0.9}}}},
					},
				},
			},
		},
		{
			name: "Let with several variables",
			got: expr.Let(
				bson.D{{Key: "a", Value: 1}, {Key: "b", Value: 2}},
				expr.Add("$$a", "$$b"),
			),
			want: bson.D{
				{
					Key: "$let",
					Value: bson.D{
						{Key: "vars", Value: bson.D{{Key: "a", Value: 1}, {Key: "b", Value: 2}}},
						{Key: "in", Value: bson.D{{Key: "$add", Value: bson.A{"$$a", "$$b"}}}},
					},
				},
			},
		},
		{
			name: "Rand takes an empty document",
			got:  expr.Rand(),
			want: bson.D{{Key: "$rand", Value: bson.D{}}},
		},
		{
			name: "BinarySize",
			got:  expr.BinarySize(expr.Field("thumbnail")),
			want: bson.D{{Key: "$binarySize", Value: "$thumbnail"}},
		},
		{
			name: "BSONSize over the whole document",
			got:  expr.BSONSize("$$ROOT"),
			want: bson.D{{Key: "$bsonSize", Value: "$$ROOT"}},
		},
		{
			name: "TsSecond",
			got:  expr.TsSecond(expr.Field("oplog_ts")),
			want: bson.D{{Key: "$tsSecond", Value: "$oplog_ts"}},
		},
		{
			name: "TsIncrement",
			got:  expr.TsIncrement(expr.Field("oplog_ts")),
			want: bson.D{{Key: "$tsIncrement", Value: "$oplog_ts"}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if !reflect.DeepEqual(tt.got, tt.want) {
					t.Fatalf("got %v, want %v", tt.got, tt.want)
				}
			},
		)
	}
}

func ExampleLet() {
	e := expr.Let(
		bson.D{{Key: "total", Value: expr.Add(expr.Field("price"), expr.Field("tax"))}},
		expr.Multiply("$$total", 0.9),
	)

	printExpr(e)
	// Output: {"$let":{"vars":{"total":{"$add":["$price","$tax"]}},"in":{"$multiply":["$$total",0.9]}}}
}

func ExampleRand() {
	e := expr.Rand()

	printExpr(e)
	// Output: {"$rand":{}}
}

func ExampleBinarySize() {
	e := expr.BinarySize(expr.Field("thumbnail"))

	printExpr(e)
	// Output: {"$binarySize":"$thumbnail"}
}

func ExampleBSONSize() {
	e := expr.BSONSize("$$ROOT")

	printExpr(e)
	// Output: {"$bsonSize":"$$ROOT"}
}

func ExampleTsSecond() {
	e := expr.TsSecond(expr.Field("oplog_ts"))

	printExpr(e)
	// Output: {"$tsSecond":"$oplog_ts"}
}

func ExampleTsIncrement() {
	e := expr.TsIncrement(expr.Field("oplog_ts"))

	printExpr(e)
	// Output: {"$tsIncrement":"$oplog_ts"}
}
