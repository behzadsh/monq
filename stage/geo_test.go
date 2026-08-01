package stage_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/stage"
)

func TestGeoNear(t *testing.T) {
	point := monq.Point(-73.97, 40.77)

	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "required fields only",
			got:  stage.GeoNear(point, "distance"),
			want: bson.D{{Key: "$geoNear", Value: bson.D{
				{Key: "near", Value: point},
				{Key: "distanceField", Value: "distance"},
			}}},
		},
		{
			name: "with a query in place of a preceding Match",
			got:  stage.GeoNear(point, "distance", stage.GeoNearQuery(monq.Eq("open", true))),
			want: bson.D{{Key: "$geoNear", Value: bson.D{
				{Key: "near", Value: point},
				{Key: "distanceField", Value: "distance"},
				{Key: "query", Value: bson.D{{Key: "open", Value: bson.D{{Key: "$eq", Value: true}}}}},
			}}},
		},
		{
			name: "with every option",
			got: stage.GeoNear(point, "distance",
				stage.Spherical(),
				stage.MaxDistance(1000),
				stage.MinDistance(10),
				stage.DistanceMultiplier(0.001),
				stage.IncludeLocs("matched_location"),
				stage.GeoNearKey("loc"),
			),
			want: bson.D{{Key: "$geoNear", Value: bson.D{
				{Key: "near", Value: point},
				{Key: "distanceField", Value: "distance"},
				{Key: "spherical", Value: true},
				{Key: "maxDistance", Value: 1000.0},
				{Key: "minDistance", Value: 10.0},
				{Key: "distanceMultiplier", Value: 0.001},
				{Key: "includeLocs", Value: "matched_location"},
				{Key: "key", Value: "loc"},
			}}},
		},
		{
			name: "with a legacy coordinate pair",
			got:  stage.GeoNear(bson.A{-73.97, 40.77}, "distance"),
			want: bson.D{{Key: "$geoNear", Value: bson.D{
				{Key: "near", Value: bson.A{-73.97, 40.77}},
				{Key: "distanceField", Value: "distance"},
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

func ExampleGeoNear() {
	s := stage.GeoNear(monq.Point(-73.97, 40.77), "distance", stage.MaxDistance(1000), stage.Spherical())

	printStage(s)
	// Output: {"$geoNear":{"near":{"type":"Point","coordinates":[-73.97,40.77]},"distanceField":"distance","maxDistance":1000.0,"spherical":true}}
}
