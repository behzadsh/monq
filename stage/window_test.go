package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
	"github.com/behzadsh/monq/stage"
)

func TestSetWindowFields(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "partition, sort, and one running total",
			got: stage.SetWindowFields(
				"$account", bson.D{{Key: "date", Value: 1}},
				stage.WindowField(
					"running_total", expr.Sum(expr.Field("amount")),
					stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent),
				),
			),
			want: bson.D{
				{
					Key: "$setWindowFields",
					Value: bson.D{
						{Key: "partitionBy", Value: "$account"},
						{Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
						{
							Key: "output",
							Value: bson.D{
								{
									Key: "running_total",
									Value: bson.D{
										{Key: "$sum", Value: "$amount"},
										{
											Key: "window",
											Value: bson.D{
												{Key: "documents", Value: bson.A{"unbounded", "current"}},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "a nil partition leaves the key out",
			got: stage.SetWindowFields(
				nil, bson.D{{Key: "date", Value: 1}},
				stage.WindowField("rank", expr.Rank()),
			),
			want: bson.D{
				{
					Key: "$setWindowFields",
					Value: bson.D{
						{Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
						{
							Key: "output",
							Value: bson.D{
								{
									Key: "rank",
									Value: bson.D{
										{Key: "$rank", Value: bson.D{}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "a nil sort leaves the key out",
			got:  stage.SetWindowFields("$account", nil, stage.WindowField("total", expr.Sum(expr.Field("amount")))),
			want: bson.D{
				{
					Key: "$setWindowFields",
					Value: bson.D{
						{Key: "partitionBy", Value: "$account"},
						{
							Key: "output",
							Value: bson.D{
								{
									Key: "total",
									Value: bson.D{
										{Key: "$sum", Value: "$amount"},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "several output fields merge in order",
			got: stage.SetWindowFields(
				"$account", bson.D{{Key: "date", Value: 1}},
				stage.WindowField("rank", expr.Rank()),
				stage.WindowField("previous", expr.Shift(expr.Field("amount"), -1)),
			),
			want: bson.D{
				{
					Key: "$setWindowFields",
					Value: bson.D{
						{Key: "partitionBy", Value: "$account"},
						{Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
						{
							Key: "output", Value: bson.D{
								{Key: "rank", Value: bson.D{{Key: "$rank", Value: bson.D{}}}},
								{
									Key: "previous", Value: bson.D{
										{
											Key: "$shift", Value: bson.D{
												{Key: "output", Value: "$amount"},
												{Key: "by", Value: -1},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "no output fields at all",
			got:  stage.SetWindowFields(nil, nil),
			want: bson.D{{Key: "$setWindowFields", Value: bson.D{{Key: "output", Value: bson.D{}}}}},
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

func TestWindowField(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "no window at all",
			got:  stage.WindowField("total", expr.Sum(expr.Field("amount"))),
			want: bson.D{{Key: "total", Value: bson.D{{Key: "$sum", Value: "$amount"}}}},
		},
		{
			name: "a moving window of four documents",
			got:  stage.WindowField("moving", expr.Avg(expr.Field("price")), stage.WindowDocuments(-3, 0)),
			want: bson.D{
				{
					Key: "moving", Value: bson.D{
						{Key: "$avg", Value: "$price"},
						{Key: "window", Value: bson.D{{Key: "documents", Value: bson.A{-3, 0}}}},
					},
				},
			},
		},
		{
			name: "a range window with a time unit shares one window document",
			got: stage.WindowField(
				"last_hour", expr.Sum(expr.Field("amount")),
				stage.WindowRange(-1, 0), stage.WindowUnit("hour"),
			),
			want: bson.D{
				{
					Key: "last_hour", Value: bson.D{
						{Key: "$sum", Value: "$amount"},
						{
							Key: "window", Value: bson.D{
								{Key: "range", Value: bson.A{-1, 0}},
								{Key: "unit", Value: "hour"},
							},
						},
					},
				},
			},
		},
		{
			name: "a dotted output path",
			got:  stage.WindowField("stats.total", expr.Sum(expr.Field("amount"))),
			want: bson.D{{Key: "stats.total", Value: bson.D{{Key: "$sum", Value: "$amount"}}}},
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

func TestWindowFieldLeavesTheOperatorUntouched(t *testing.T) {
	operator := expr.Sum(expr.Field("amount"))

	stage.WindowField("a", operator, stage.WindowDocuments(-3, 0))
	stage.WindowField("b", operator, stage.WindowRange(-1, 0))

	want := bson.D{{Key: "$sum", Value: "$amount"}}
	if !reflect.DeepEqual(operator, want) {
		t.Fatalf("operator = %v, want %v unchanged", operator, want)
	}
}

func ExampleSetWindowFields() {
	s := stage.SetWindowFields(
		"$acct", monq.Sort(monq.Asc("date")),
		stage.WindowField(
			"total", expr.Sum(expr.Field("amt")),
			stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent),
		),
	)

	printStage(s)
	// Output: {"$setWindowFields":{"partitionBy":"$acct","sortBy":{"date":1},"output":{"total":{"$sum":"$amt","window":{"documents":["unbounded","current"]}}}}}
}

func ExampleWindowField() {
	field := stage.WindowField("moving_average", expr.Avg(expr.Field("price")), stage.WindowDocuments(-3, 0))

	printStage(field)
	// Output: {"moving_average":{"$avg":"$price","window":{"documents":[-3,0]}}}
}

func ExampleWindowDocuments() {
	field := stage.WindowField(
		"running_total", expr.Sum(expr.Field("amount")),
		stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent),
	)

	printStage(field)
	// Output: {"running_total":{"$sum":"$amount","window":{"documents":["unbounded","current"]}}}
}

func ExampleWindowRange() {
	field := stage.WindowField("nearby", expr.Sum(expr.Field("amount")), stage.WindowRange(-10, 10))

	printStage(field)
	// Output: {"nearby":{"$sum":"$amount","window":{"range":[-10,10]}}}
}

func ExampleWindowUnit() {
	field := stage.WindowField(
		"last_hour", expr.Sum(expr.Field("amount")),
		stage.WindowRange(-1, 0), stage.WindowUnit("hour"),
	)

	printStage(field)
	// Output: {"last_hour":{"$sum":"$amount","window":{"range":[-1,0],"unit":"hour"}}}
}
