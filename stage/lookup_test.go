package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/stage"
)

func TestJoinStages(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Lookup on equality",
			got:  stage.Lookup("orders", "_id", "customer_id", "orders"),
			want: bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "orders"},
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "customer_id"},
				{Key: "as", Value: "orders"},
			}}},
		},
		{
			name: "Lookup with a sub-pipeline",
			got: stage.Lookup("orders", "_id", "customer_id", "orders",
				stage.SubPipeline(stage.Sort(monq.Sort(monq.Asc("placed_at")))),
			),
			want: bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "orders"},
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "customer_id"},
				{Key: "as", Value: "orders"},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$sort", Value: bson.D{{Key: "placed_at", Value: 1}}}},
				}},
			}}},
		},
		{
			name: "Lookup with variables and a sub-pipeline",
			got: stage.Lookup("orders", "_id", "customer_id", "orders",
				stage.Let(bson.D{{Key: "region", Value: "$region"}}),
				stage.SubPipeline(stage.Match(monq.Expr(bson.D{
					{Key: "$eq", Value: bson.A{"$region", "$$region"}},
				}))),
			),
			want: bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "orders"},
				{Key: "localField", Value: "_id"},
				{Key: "foreignField", Value: "customer_id"},
				{Key: "as", Value: "orders"},
				{Key: "let", Value: bson.D{{Key: "region", Value: "$region"}}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{{Key: "$expr", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$region", "$$region"}},
					}}}}},
				}},
			}}},
		},
		{
			name: "LookupPipeline with variables",
			got: stage.LookupPipeline("orders",
				bson.D{{Key: "customer", Value: "$_id"}},
				[]bson.D{stage.Match(monq.Expr(bson.D{
					{Key: "$eq", Value: bson.A{"$customer_id", "$$customer"}},
				}))},
				"orders",
			),
			want: bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "orders"},
				{Key: "let", Value: bson.D{{Key: "customer", Value: "$_id"}}},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{{Key: "$expr", Value: bson.D{
						{Key: "$eq", Value: bson.A{"$customer_id", "$$customer"}},
					}}}}},
				}},
				{Key: "as", Value: "orders"},
			}}},
		},
		{
			name: "LookupPipeline without variables leaves out let",
			got:  stage.LookupPipeline("orders", nil, []bson.D{stage.Limit(5)}, "orders"),
			want: bson.D{{Key: "$lookup", Value: bson.D{
				{Key: "from", Value: "orders"},
				{Key: "pipeline", Value: bson.A{bson.D{{Key: "$limit", Value: int64(5)}}}},
				{Key: "as", Value: "orders"},
			}}},
		},
		{
			name: "GraphLookup with its required fields",
			got:  stage.GraphLookup("employees", "$reports_to", "reports_to", "name", "chain"),
			want: bson.D{{Key: "$graphLookup", Value: bson.D{
				{Key: "from", Value: "employees"},
				{Key: "startWith", Value: "$reports_to"},
				{Key: "connectFromField", Value: "reports_to"},
				{Key: "connectToField", Value: "name"},
				{Key: "as", Value: "chain"},
			}}},
		},
		{
			name: "GraphLookup with every option",
			got: stage.GraphLookup("employees", "$reports_to", "reports_to", "name", "chain",
				stage.MaxDepth(3),
				stage.DepthField("level"),
				stage.RestrictSearchWithMatch(monq.Eq("active", true)),
			),
			want: bson.D{{Key: "$graphLookup", Value: bson.D{
				{Key: "from", Value: "employees"},
				{Key: "startWith", Value: "$reports_to"},
				{Key: "connectFromField", Value: "reports_to"},
				{Key: "connectToField", Value: "name"},
				{Key: "as", Value: "chain"},
				{Key: "maxDepth", Value: 3},
				{Key: "depthField", Value: "level"},
				{Key: "restrictSearchWithMatch", Value: bson.D{
					{Key: "active", Value: bson.D{{Key: "$eq", Value: true}}},
				}},
			}}},
		},
		{
			name: "UnionWith a whole collection",
			got:  stage.UnionWith("archived_orders"),
			want: bson.D{{Key: "$unionWith", Value: bson.D{{Key: "coll", Value: "archived_orders"}}}},
		},
		{
			name: "UnionWith a sub-pipeline",
			got:  stage.UnionWith("archived_orders", stage.Match(monq.Eq("region", "eu"))),
			want: bson.D{{Key: "$unionWith", Value: bson.D{
				{Key: "coll", Value: "archived_orders"},
				{Key: "pipeline", Value: bson.A{
					bson.D{{Key: "$match", Value: bson.D{
						{Key: "region", Value: bson.D{{Key: "$eq", Value: "eu"}}},
					}}},
				}},
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

func ExampleLookup() {
	s := stage.Lookup("orders", "_id", "customer_id", "orders",
		stage.SubPipeline(stage.Sort(monq.Sort(monq.Asc("placed_at")))),
	)

	printStage(s)
	// Output: {"$lookup":{"from":"orders","localField":"_id","foreignField":"customer_id","as":"orders","pipeline":[{"$sort":{"placed_at":1}}]}}
}

func ExampleLookupPipeline() {
	s := stage.LookupPipeline("orders",
		bson.D{{Key: "customer", Value: "$_id"}},
		[]bson.D{stage.Match(monq.Expr(bson.D{{Key: "$eq", Value: bson.A{"$customer_id", "$$customer"}}}))},
		"orders",
	)

	printStage(s)
	// Output: {"$lookup":{"from":"orders","let":{"customer":"$_id"},"pipeline":[{"$match":{"$expr":{"$eq":["$customer_id","$$customer"]}}}],"as":"orders"}}
}

func ExampleGraphLookup() {
	s := stage.GraphLookup("employees", "$reports_to", "reports_to", "name", "chain", stage.MaxDepth(3))

	printStage(s)
	// Output: {"$graphLookup":{"from":"employees","startWith":"$reports_to","connectFromField":"reports_to","connectToField":"name","as":"chain","maxDepth":3}}
}

func ExampleUnionWith() {
	s := stage.UnionWith("archived_orders", stage.Match(monq.Eq("region", "eu")))

	printStage(s)
	// Output: {"$unionWith":{"coll":"archived_orders","pipeline":[{"$match":{"region":{"$eq":"eu"}}}]}}
}

func ExampleMaxDepth() {
	s := stage.GraphLookup("employees", "$reports_to", "reports_to", "name", "chain", stage.MaxDepth(3))

	printStage(s)
	// Output: {"$graphLookup":{"from":"employees","startWith":"$reports_to","connectFromField":"reports_to","connectToField":"name","as":"chain","maxDepth":3}}
}

func ExampleDepthField() {
	s := stage.GraphLookup("staff", "$boss", "boss", "name", "chain", stage.DepthField("level"))

	printStage(s)
	// Output: {"$graphLookup":{"from":"staff","startWith":"$boss","connectFromField":"boss","connectToField":"name","as":"chain","depthField":"level"}}
}

func ExampleRestrictSearchWithMatch() {
	// Passed to stage.GraphLookup, this option contributes one field to the stage's specification. It is applied
	// on its own here only to show that field without the rest of a $graphLookup around it.
	var spec bson.D
	stage.RestrictSearchWithMatch(monq.Eq("active", true))(&spec)

	printStage(spec)
	// Output: {"restrictSearchWithMatch":{"active":{"$eq":true}}}
}
