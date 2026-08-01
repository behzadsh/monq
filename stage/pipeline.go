package stage

import "go.mongodb.org/mongo-driver/v2/bson"

// Pipeline collects stages into an aggregation pipeline.
//
// A pipeline is a list of stage documents that run in order, each one feeding the next. The driver's
// mongo.Pipeline is defined as []bson.D, so the returned slice is assignable to it and goes straight into
// Aggregate; the return type stays []bson.D to keep this package's dependencies down to bson alone. A plain slice
// literal works exactly as well, and Pipeline is there to keep a pipeline reading as a call like the stages inside
// it. Stages are taken in the order given, with no reordering and no validation, so a pipeline that has to start
// with $geoNear or $match is the caller's responsibility to write that way.
//
// Example:
//
//	stage.Pipeline(
//		stage.Match(monq.Eq("status", "active")),
//		stage.Limit(20),
//	)
//	// []bson.D{
//	//     {{Key: "$match", Value: bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}}}},
//	//     {{Key: "$limit", Value: int64(20)}},
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/aggregation-pipeline/
func Pipeline(stages ...bson.D) []bson.D {
	pipeline := make([]bson.D, len(stages))
	copy(pipeline, stages)

	return pipeline
}
