package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestGeometryConstructors(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "GeoJSON passes coordinates through",
			got:  monq.GeoJSON("LineString", bson.A{bson.A{-73.97, 40.77}, bson.A{-73.88, 40.78}}),
			want: bson.D{
				{Key: "type", Value: "LineString"},
				{Key: "coordinates", Value: bson.A{bson.A{-73.97, 40.77}, bson.A{-73.88, 40.78}}},
			},
		},
		{
			name: "Point orders longitude before latitude",
			got:  monq.Point(-73.97, 40.77),
			want: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: bson.A{-73.97, 40.77}},
			},
		},
		{
			name: "Polygon with one ring",
			got:  monq.Polygon([][2]float64{{0, 0}, {3, 0}, {3, 3}, {0, 0}}),
			want: bson.D{
				{Key: "type", Value: "Polygon"},
				{Key: "coordinates", Value: bson.A{bson.A{
					bson.A{0.0, 0.0}, bson.A{3.0, 0.0}, bson.A{3.0, 3.0}, bson.A{0.0, 0.0},
				}}},
			},
		},
		{
			name: "Polygon with a hole",
			got: monq.Polygon(
				[][2]float64{{0, 0}, {4, 0}, {4, 4}, {0, 0}},
				[][2]float64{{1, 1}, {2, 1}, {2, 2}, {1, 1}},
			),
			want: bson.D{
				{Key: "type", Value: "Polygon"},
				{Key: "coordinates", Value: bson.A{
					bson.A{bson.A{0.0, 0.0}, bson.A{4.0, 0.0}, bson.A{4.0, 4.0}, bson.A{0.0, 0.0}},
					bson.A{bson.A{1.0, 1.0}, bson.A{2.0, 1.0}, bson.A{2.0, 2.0}, bson.A{1.0, 1.0}},
				}},
			},
		},
		{
			name: "Geometry wraps a bare object",
			got:  monq.Geometry(monq.Point(-73.97, 40.77)),
			want: bson.D{{Key: "$geometry", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: bson.A{-73.97, 40.77}},
			}}},
		},
		{
			name: "Box holds both corners",
			got:  monq.Box([2]float64{0, 0}, [2]float64{3, 3}),
			want: bson.D{{Key: "$box", Value: bson.A{bson.A{0.0, 0.0}, bson.A{3.0, 3.0}}}},
		},
		{
			name: "Center holds the point and radius",
			got:  monq.Center([2]float64{-73.97, 40.77}, 0.5),
			want: bson.D{{Key: "$center", Value: bson.A{bson.A{-73.97, 40.77}, 0.5}}},
		},
		{
			name: "CenterSphere holds the point and radians",
			got:  monq.CenterSphere([2]float64{-73.97, 40.77}, 0.00078),
			want: bson.D{{Key: "$centerSphere", Value: bson.A{bson.A{-73.97, 40.77}, 0.00078}}},
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

func TestGeoOperators(t *testing.T) {
	tests := []struct {
		name  string
		got   bson.D
		field monq.FieldPath
		op    string
		want  bson.D
	}{
		{
			name:  "GeoWithin with GeoJSON",
			got:   monq.GeoWithin("loc", monq.Geometry(monq.Point(-73.97, 40.77))),
			field: "loc",
			op:    "$geoWithin",
			want:  monq.Geometry(monq.Point(-73.97, 40.77)),
		},
		{
			name:  "GeoWithin with a legacy shape",
			got:   monq.GeoWithin("loc", monq.Box([2]float64{0, 0}, [2]float64{3, 3})),
			field: "loc",
			op:    "$geoWithin",
			want:  monq.Box([2]float64{0, 0}, [2]float64{3, 3}),
		},
		{
			name:  "GeoIntersects",
			got:   monq.GeoIntersects("area", monq.Geometry(monq.Point(-73.97, 40.77))),
			field: "area",
			op:    "$geoIntersects",
			want:  monq.Geometry(monq.Point(-73.97, 40.77)),
		},
		{
			name:  "Near without options",
			got:   monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77))),
			field: "loc",
			op:    "$near",
			want:  monq.Geometry(monq.Point(-73.97, 40.77)),
		},
		{
			name: "Near with both distance bounds",
			got: monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77)),
				monq.MaxDistance(1000), monq.MinDistance(10)),
			field: "loc",
			op:    "$near",
			want: append(monq.Geometry(monq.Point(-73.97, 40.77)),
				bson.E{Key: "$maxDistance", Value: 1000.0},
				bson.E{Key: "$minDistance", Value: 10.0},
			),
		},
		{
			name:  "NearSphere with a maximum distance",
			got:   monq.NearSphere("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MaxDistance(1000)),
			field: "loc",
			op:    "$nearSphere",
			want: append(monq.Geometry(monq.Point(-73.97, 40.77)),
				bson.E{Key: "$maxDistance", Value: 1000.0},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got[0].Key != string(tt.field) {
				t.Fatalf("field = %q, want %q", tt.got[0].Key, tt.field)
			}

			inner, ok := tt.got[0].Value.(bson.D)
			if !ok || inner[0].Key != tt.op {
				t.Fatalf("got %v, want a %s operator document", tt.got, tt.op)
			}

			if !reflect.DeepEqual(inner[0].Value, tt.want) {
				t.Fatalf("%s argument = %v, want %v", tt.op, inner[0].Value, tt.want)
			}
		})
	}
}

func TestNearLeavesGeometryUntouched(t *testing.T) {
	geometry := monq.Geometry(monq.Point(-73.97, 40.77))

	monq.Near("loc", geometry, monq.MaxDistance(1000))
	monq.NearSphere("loc", geometry, monq.MinDistance(10))

	if len(geometry) != 1 || geometry[0].Key != "$geometry" {
		t.Fatalf("geometry = %v, want the caller's document unchanged", geometry)
	}
}

func TestGeoOperatorsComposeWithNot(t *testing.T) {
	tests := []struct {
		name string
		expr bson.D
		op   string
	}{
		{
			name: "negates GeoWithin",
			expr: monq.GeoWithin("loc", monq.Box([2]float64{0, 0}, [2]float64{3, 3})),
			op:   "$geoWithin",
		},
		{
			name: "negates GeoIntersects",
			expr: monq.GeoIntersects("area", monq.Geometry(monq.Point(-73.97, 40.77))),
			op:   "$geoIntersects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := monq.Not(tt.expr)

			outer, ok := got[0].Value.(bson.D)
			if !ok || outer[0].Key != "$not" {
				t.Fatalf("Not() = %v, want a $not operator document", got)
			}

			inner, ok := outer[0].Value.(bson.D)
			if !ok || inner[0].Key != tt.op {
				t.Fatalf("Not() inner operator = %v, want %q", outer[0].Value, tt.op)
			}
		})
	}
}

func ExampleGeoJSON() {
	geometry := monq.GeoJSON("LineString", bson.A{bson.A{-73.97, 40.77}, bson.A{-73.88, 40.78}})

	printFilter(geometry)
	// Output: {"type":"LineString","coordinates":[[-73.97,40.77],[-73.88,40.78]]}
}

func ExamplePoint() {
	geometry := monq.Point(-73.97, 40.77)

	printFilter(geometry)
	// Output: {"type":"Point","coordinates":[-73.97,40.77]}
}

func ExamplePolygon() {
	geometry := monq.Polygon([][2]float64{{0, 0}, {3, 0}, {3, 3}, {0, 0}})

	printFilter(geometry)
	// Output: {"type":"Polygon","coordinates":[[[0.0,0.0],[3.0,0.0],[3.0,3.0],[0.0,0.0]]]}
}

func ExampleGeometry() {
	argument := monq.Geometry(monq.Point(-73.97, 40.77))

	printFilter(argument)
	// Output: {"$geometry":{"type":"Point","coordinates":[-73.97,40.77]}}
}

func ExampleBox() {
	shape := monq.Box([2]float64{0, 0}, [2]float64{3, 3})

	printFilter(shape)
	// Output: {"$box":[[0.0,0.0],[3.0,3.0]]}
}

func ExampleCenter() {
	shape := monq.Center([2]float64{-73.97, 40.77}, 0.5)

	printFilter(shape)
	// Output: {"$center":[[-73.97,40.77],0.5]}
}

func ExampleCenterSphere() {
	shape := monq.CenterSphere([2]float64{-73.97, 40.77}, 0.00078)

	printFilter(shape)
	// Output: {"$centerSphere":[[-73.97,40.77],0.00078]}
}

func ExampleGeoWithin() {
	filter := monq.GeoWithin("loc", monq.Geometry(monq.Polygon([][2]float64{{0, 0}, {3, 0}, {3, 3}, {0, 0}})))

	printFilter(filter)
	// Output: {"loc":{"$geoWithin":{"$geometry":{"type":"Polygon","coordinates":[[[0.0,0.0],[3.0,0.0],[3.0,3.0],[0.0,0.0]]]}}}}
}

func ExampleGeoIntersects() {
	filter := monq.GeoIntersects("area", monq.Geometry(monq.Point(-73.97, 40.77)))

	printFilter(filter)
	// Output: {"area":{"$geoIntersects":{"$geometry":{"type":"Point","coordinates":[-73.97,40.77]}}}}
}

func ExampleNear() {
	filter := monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MaxDistance(1000))

	printFilter(filter)
	// Output: {"loc":{"$near":{"$geometry":{"type":"Point","coordinates":[-73.97,40.77]},"$maxDistance":1000.0}}}
}

func ExampleNearSphere() {
	filter := monq.NearSphere("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MinDistance(10))

	printFilter(filter)
	// Output: {"loc":{"$nearSphere":{"$geometry":{"type":"Point","coordinates":[-73.97,40.77]},"$minDistance":10.0}}}
}

func ExampleMaxDistance() {
	filter := monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MaxDistance(1000))

	printFilter(filter)
	// Output: {"loc":{"$near":{"$geometry":{"type":"Point","coordinates":[-73.97,40.77]},"$maxDistance":1000.0}}}
}

func ExampleMinDistance() {
	filter := monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MinDistance(10))

	printFilter(filter)
	// Output: {"loc":{"$near":{"$geometry":{"type":"Point","coordinates":[-73.97,40.77]},"$minDistance":10.0}}}
}
