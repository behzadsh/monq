package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
	"github.com/behzadsh/monq/stage"
)

func TestMiscStages(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Redact takes an expression",
			got:  stage.Redact(expr.Cond(expr.Eq(expr.Field("level"), "public"), "$$DESCEND", "$$PRUNE")),
			want: bson.D{
				{
					Key: "$redact",
					Value: bson.D{
						{
							Key: "$cond",
							Value: bson.D{
								{Key: "if", Value: bson.D{{Key: "$eq", Value: bson.A{"$level", "public"}}}},
								{Key: "then", Value: "$$DESCEND"},
								{Key: "else", Value: "$$PRUNE"},
							},
						},
					},
				},
			},
		},
		{
			name: "DensifyRange without a unit",
			got:  stage.DensifyRange(1, stage.DensifyFull),
			want: bson.D{{Key: "step", Value: 1}, {Key: "bounds", Value: "full"}},
		},
		{
			name: "DensifyRange with a unit",
			got:  stage.DensifyRange(1, stage.DensifyPartition, "hour"),
			want: bson.D{
				{Key: "step", Value: 1},
				{Key: "bounds", Value: "partition"},
				{Key: "unit", Value: "hour"},
			},
		},
		{
			name: "DensifyRange with explicit bounds",
			got:  stage.DensifyRange(5, bson.A{0, 100}),
			want: bson.D{{Key: "step", Value: 5}, {Key: "bounds", Value: bson.A{0, 100}}},
		},
		{
			name: "Densify over one series",
			got:  stage.Densify("timestamp", stage.DensifyRange(1, stage.DensifyFull, "hour")),
			want: bson.D{
				{
					Key: "$densify",
					Value: bson.D{
						{Key: "field", Value: "timestamp"},
						{
							Key: "range", Value: bson.D{
								{Key: "step", Value: 1},
								{Key: "bounds", Value: "full"},
								{Key: "unit", Value: "hour"},
							},
						},
					},
				},
			},
		},
		{
			name: "Densify per partition",
			got: stage.Densify(
				"timestamp", stage.DensifyRange(1, stage.DensifyPartition, "hour"),
				stage.DensifyPartitionByFields("sensor_id"),
			),
			want: bson.D{
				{
					Key: "$densify", Value: bson.D{
						{Key: "field", Value: "timestamp"},
						{Key: "partitionByFields", Value: "sensor_id"},
						{
							Key: "range", Value: bson.D{
								{Key: "step", Value: 1},
								{Key: "bounds", Value: "partition"},
								{Key: "unit", Value: "hour"},
							},
						},
					},
				},
			},
		},
		{
			name: "Densify partitioned by several fields",
			got: stage.Densify(
				"timestamp", stage.DensifyRange(1, stage.DensifyPartition),
				stage.DensifyPartitionByFields("sensor_id", "site"),
			),
			want: bson.D{
				{
					Key: "$densify", Value: bson.D{
						{Key: "field", Value: "timestamp"},
						{Key: "partitionByFields", Value: bson.A{"sensor_id", "site"}},
						{
							Key: "range", Value: bson.D{
								{Key: "step", Value: 1},
								{Key: "bounds", Value: "partition"},
							},
						},
					},
				},
			},
		},
		{
			name: "FillValue",
			got:  stage.FillValue("price", 0),
			want: bson.D{{Key: "price", Value: bson.D{{Key: "value", Value: 0}}}},
		},
		{
			name: "FillMethod",
			got:  stage.FillMethod("price", "locf"),
			want: bson.D{{Key: "price", Value: bson.D{{Key: "method", Value: "locf"}}}},
		},
		{
			name: "Fill with a constant and no options",
			got:  stage.Fill([]bson.D{stage.FillValue("price", 0)}),
			want: bson.D{
				{
					Key: "$fill", Value: bson.D{
						{Key: "output", Value: bson.D{{Key: "price", Value: bson.D{{Key: "value", Value: 0}}}}},
					},
				},
			},
		},
		{
			name: "Fill by carrying the last value forward",
			got: stage.Fill(
				[]bson.D{stage.FillMethod("price", "locf")},
				stage.FillSortBy(monq.Sort(monq.Asc("date"))),
			),
			want: bson.D{
				{
					Key: "$fill", Value: bson.D{
						{Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
						{Key: "output", Value: bson.D{{Key: "price", Value: bson.D{{Key: "method", Value: "locf"}}}}},
					},
				},
			},
		},
		{
			name: "Fill within partitions",
			got: stage.Fill(
				[]bson.D{stage.FillMethod("price", "linear")},
				stage.FillPartitionByFields("symbol"),
				stage.FillSortBy(monq.Sort(monq.Asc("date"))),
			),
			want: bson.D{
				{
					Key: "$fill", Value: bson.D{
						{Key: "partitionByFields", Value: "symbol"},
						{Key: "sortBy", Value: bson.D{{Key: "date", Value: 1}}},
						{Key: "output", Value: bson.D{{Key: "price", Value: bson.D{{Key: "method", Value: "linear"}}}}},
					},
				},
			},
		},
		{
			name: "Fill partitioned by an expression",
			got: stage.Fill(
				[]bson.D{stage.FillValue("price", 0)},
				stage.FillPartitionBy(expr.Field("symbol")),
			),
			want: bson.D{
				{
					Key: "$fill", Value: bson.D{
						{Key: "partitionBy", Value: "$symbol"},
						{Key: "output", Value: bson.D{{Key: "price", Value: bson.D{{Key: "value", Value: 0}}}}},
					},
				},
			},
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

func ExampleRedact() {
	s := stage.Redact(expr.Cond(expr.Eq(expr.Field("level"), "public"), "$$DESCEND", "$$PRUNE"))

	printStage(s)
	// Output: {"$redact":{"$cond":{"if":{"$eq":["$level","public"]},"then":"$$DESCEND","else":"$$PRUNE"}}}
}

func ExampleDensify() {
	s := stage.Densify("timestamp", stage.DensifyRange(1, stage.DensifyFull, "hour"))

	printStage(s)
	// Output: {"$densify":{"field":"timestamp","range":{"step":1,"bounds":"full","unit":"hour"}}}
}

func ExampleDensifyRange() {
	spec := stage.DensifyRange(5, bson.A{0, 100})

	printStage(spec)
	// Output: {"step":5,"bounds":[0,100]}
}

func ExampleDensifyPartitionByFields() {
	s := stage.Densify(
		"timestamp", stage.DensifyRange(1, stage.DensifyPartition, "hour"),
		stage.DensifyPartitionByFields("sensor_id"),
	)

	printStage(s)
	// Output: {"$densify":{"field":"timestamp","partitionByFields":"sensor_id","range":{"step":1,"bounds":"partition","unit":"hour"}}}
}

func ExampleFill() {
	s := stage.Fill([]bson.D{stage.FillMethod("price", "locf")}, stage.FillSortBy(monq.Sort(monq.Asc("date"))))

	printStage(s)
	// Output: {"$fill":{"sortBy":{"date":1},"output":{"price":{"method":"locf"}}}}
}

func ExampleFillValue() {
	s := stage.Fill([]bson.D{stage.FillValue("price", 0)})

	printStage(s)
	// Output: {"$fill":{"output":{"price":{"value":0}}}}
}

func ExampleFillMethod() {
	s := stage.Fill([]bson.D{stage.FillMethod("price", "linear")}, stage.FillSortBy(monq.Sort(monq.Asc("date"))))

	printStage(s)
	// Output: {"$fill":{"sortBy":{"date":1},"output":{"price":{"method":"linear"}}}}
}

func ExampleFillSortBy() {
	s := stage.Fill([]bson.D{stage.FillMethod("price", "locf")}, stage.FillSortBy(monq.Sort(monq.Asc("date"))))

	printStage(s)
	// Output: {"$fill":{"sortBy":{"date":1},"output":{"price":{"method":"locf"}}}}
}

func ExampleFillPartitionBy() {
	s := stage.Fill([]bson.D{stage.FillValue("price", 0)}, stage.FillPartitionBy(expr.Field("symbol")))

	printStage(s)
	// Output: {"$fill":{"partitionBy":"$symbol","output":{"price":{"value":0}}}}
}

func ExampleFillPartitionByFields() {
	s := stage.Fill([]bson.D{stage.FillValue("price", 0)}, stage.FillPartitionByFields("symbol", "exchange"))

	printStage(s)
	// Output: {"$fill":{"partitionByFields":["symbol","exchange"],"output":{"price":{"value":0}}}}
}
