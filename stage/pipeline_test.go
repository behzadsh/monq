package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/stage"
)

func TestPipeline(t *testing.T) {
	tests := []struct {
		name   string
		stages []bson.D
		want   []bson.D
	}{
		{
			name:   "no stages",
			stages: nil,
			want:   []bson.D{},
		},
		{
			name: "keeps the order it is given",
			stages: []bson.D{
				stage.Match(monq.Eq("status", "active")),
				stage.Sort(bson.D{{Key: "created_at", Value: -1}}),
				stage.Limit(20),
			},
			want: []bson.D{
				{{Key: "$match", Value: bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}}}},
				{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}}}},
				{{Key: "$limit", Value: int64(20)}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				got := stage.Pipeline(tt.stages...)

				if !reflect.DeepEqual(got, tt.want) {
					t.Fatalf("Pipeline() = %v, want %v", got, tt.want)
				}
			},
		)
	}
}

func TestPipelineIsAssignableToDriverPipeline(t *testing.T) {
	// mongo.Pipeline is defined as []bson.D, so the returned slice goes into Aggregate as is. The named type is
	// declared here rather than imported, to keep the test dependencies down to bson like the package itself.
	type driverPipeline []bson.D

	var pipeline driverPipeline = stage.Pipeline(stage.Limit(20))

	if len(pipeline) != 1 {
		t.Fatalf("pipeline = %v, want one stage", pipeline)
	}
}

func ExamplePipeline() {
	pipeline := stage.Pipeline(
		stage.Match(monq.Eq("status", "active")),
		stage.Limit(20),
	)

	for _, s := range pipeline {
		printStage(s)
	}
	// Output:
	// {"$match":{"status":{"$eq":"active"}}}
	// {"$limit":20}
}
