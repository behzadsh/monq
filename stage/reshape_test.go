package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/stage"
)

func TestReshapeStages(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Field builds one pair",
			got:  stage.Field("name", "$profile.display_name"),
			want: bson.D{{Key: "name", Value: "$profile.display_name"}},
		},
		{
			name: "Project with inclusion flags",
			got:  stage.Project(stage.Field("name", 1), stage.Field("_id", 0)),
			want: bson.D{{Key: "$project", Value: bson.D{
				{Key: "name", Value: 1},
				{Key: "_id", Value: 0},
			}}},
		},
		{
			name: "Project takes a hand-written document too",
			got:  stage.Project(bson.D{{Key: "name", Value: 1}, {Key: "email", Value: 1}}),
			want: bson.D{{Key: "$project", Value: bson.D{
				{Key: "name", Value: 1},
				{Key: "email", Value: 1},
			}}},
		},
		{
			name: "Project keeps the last of two fields sharing a name",
			got:  stage.Project(stage.Field("name", 1), stage.Field("name", "$full_name")),
			want: bson.D{{Key: "$project", Value: bson.D{{Key: "name", Value: "$full_name"}}}},
		},
		{
			name: "AddFields",
			got:  stage.AddFields(stage.Field("total", "$amount")),
			want: bson.D{{Key: "$addFields", Value: bson.D{{Key: "total", Value: "$amount"}}}},
		},
		{
			name: "Set is the newer name of AddFields",
			got:  stage.Set(stage.Field("status", "active")),
			want: bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}},
		},
		{
			name: "Unset with one field is a string",
			got:  stage.Unset("password_hash"),
			want: bson.D{{Key: "$unset", Value: "password_hash"}},
		},
		{
			name: "Unset with several fields is an array",
			got:  stage.Unset("password_hash", "internal.notes"),
			want: bson.D{{Key: "$unset", Value: bson.A{"password_hash", "internal.notes"}}},
		},
		{
			name: "Unset with no fields is an empty array",
			got:  stage.Unset(),
			want: bson.D{{Key: "$unset", Value: bson.A{}}},
		},
		{
			name: "ReplaceRoot wraps its expression",
			got:  stage.ReplaceRoot("$profile"),
			want: bson.D{{Key: "$replaceRoot", Value: bson.D{{Key: "newRoot", Value: "$profile"}}}},
		},
		{
			name: "ReplaceWith takes the expression directly",
			got:  stage.ReplaceWith("$profile"),
			want: bson.D{{Key: "$replaceWith", Value: "$profile"}},
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

func TestStageSetAndUpdateSetCoexist(t *testing.T) {
	// The two $set spellings are what the package split is for: one is a pipeline stage, the other an update
	// operator, and the qualifier says which is meant.
	pipelineStage := stage.Set(stage.Field("status", "active"))
	updateOperator := monq.Set("status", "active")

	wantStage := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}}
	if !reflect.DeepEqual(pipelineStage, wantStage) {
		t.Fatalf("stage.Set() = %v, want %v", pipelineStage, wantStage)
	}

	wantUpdate := bson.D{{Key: "$set", Value: bson.D{{Key: "status", Value: "active"}}}}
	if !reflect.DeepEqual(updateOperator, wantUpdate) {
		t.Fatalf("monq.Set() = %v, want %v", updateOperator, wantUpdate)
	}
}

func ExampleField() {
	field := stage.Field("name", "$profile.display_name")

	printStage(field)
	// Output: {"name":"$profile.display_name"}
}

func ExampleProject() {
	s := stage.Project(stage.Field("name", 1), stage.Field("_id", 0))

	printStage(s)
	// Output: {"$project":{"name":1,"_id":0}}
}

func ExampleAddFields() {
	s := stage.AddFields(stage.Field("total", bson.D{{Key: "$sum", Value: "$items.price"}}))

	printStage(s)
	// Output: {"$addFields":{"total":{"$sum":"$items.price"}}}
}

func ExampleSet() {
	s := stage.Set(stage.Field("status", "active"))

	printStage(s)
	// Output: {"$set":{"status":"active"}}
}

func ExampleUnset() {
	s := stage.Unset("password_hash", "internal.notes")

	printStage(s)
	// Output: {"$unset":["password_hash","internal.notes"]}
}

func ExampleReplaceRoot() {
	s := stage.ReplaceRoot("$profile")

	printStage(s)
	// Output: {"$replaceRoot":{"newRoot":"$profile"}}
}

func ExampleReplaceWith() {
	s := stage.ReplaceWith("$profile")

	printStage(s)
	// Output: {"$replaceWith":"$profile"}
}
