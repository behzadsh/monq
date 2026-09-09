package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/stage"
)

// runGeo covers the geospatial operators, which come with their geometry so no GeoJSON has to be written by hand.
// Every position is [longitude, latitude], the reverse of the order coordinates are usually quoted in, and the
// queries need the 2dsphere index the seed created.
func runGeo(ctx context.Context, db *mongo.Database) error {
	return runParts(
		ctx, db.Collection(productsColl),
		geoNear,
		geoWithin,
		geoStage,
	)
}

// manhattan is a point in New York, where three of the seeded warehouses are.
var manhattan = monq.Point(-73.98, 40.75)

// geoOnly is the projection the geo demo reads results through.
var geoOnly = monq.Projection(monq.Include(ProductPaths.SKU, warehousePath), monq.Exclude(ProductPaths.ID))

// geoNear shows the distance queries, which sort their results by distance on their own.
func geoNear(ctx context.Context, coll *mongo.Collection) error {
	step("Point builds the GeoJSON, Geometry hands it to the operator, and the options are typed to it")

	query("monq.Point(-73.98, 40.75)", manhattan)

	near := monq.Near(warehousePath, monq.Geometry(manhattan), monq.MaxDistance(20000))
	if err := showFind(
		ctx, coll, "monq.Near(warehouse, monq.Geometry(point), monq.MaxDistance(20000))", near,
		options.Find().SetProjection(geoOnly),
	); err != nil {
		return err
	}

	step("$nearSphere is the same query on a sphere, which for GeoJSON data is what $near already does")

	sphere := monq.NearSphere(warehousePath, monq.Geometry(manhattan), monq.MinDistance(500_000))

	return showFind(
		ctx, coll, "monq.NearSphere(warehouse, monq.Geometry(point), monq.MinDistance(500000))", sphere,
		options.Find().SetProjection(geoOnly),
	)
}

// geoWithin shows the containment queries and the shapes they take.
func geoWithin(ctx context.Context, coll *mongo.Collection) error {
	step("Polygon takes its rings as [longitude, latitude] pairs and closes nothing for you: the last must repeat the first")

	midwest := monq.Polygon(
		[][2]float64{
			{-95.0, 36.0},
			{-80.0, 36.0},
			{-80.0, 45.0},
			{-95.0, 45.0},
			{-95.0, 36.0},
		},
	)
	query("monq.Polygon(ring)", midwest)

	if err := showFind(
		ctx, coll, "monq.GeoWithin(warehouse, monq.Geometry(polygon))", monq.GeoWithin(warehousePath, monq.Geometry(midwest)),
		options.Find().SetProjection(geoOnly),
	); err != nil {
		return err
	}

	step("$geoIntersects asks whether the shapes touch at all, and GeoJSON builds the types Polygon and Point do not cover")

	westCoast := monq.GeoJSON(
		"Polygon", [][][2]float64{
			{
				{-125.0, 45.0},
				{-118.0, 45.0},
				{-118.0, 49.0},
				{-125.0, 49.0},
				{-125.0, 45.0},
			},
		},
	)
	query(`monq.GeoJSON("Polygon", rings)`, westCoast)

	if err := showFind(
		ctx, coll, "monq.GeoIntersects(warehouse, monq.Geometry(shape))", monq.GeoIntersects(warehousePath, monq.Geometry(westCoast)),
		options.Find().SetProjection(geoOnly),
	); err != nil {
		return err
	}

	step("the legacy shapes are for 2d data, and $centerSphere takes its radius in radians")

	radians := 100.0 / 6378.1
	query("monq.CenterSphere(center, 100/6378.1)", monq.CenterSphere([2]float64{-73.98, 40.75}, radians))
	query("monq.Box(bottomLeft, topRight)", monq.Box([2]float64{-90, 30}, [2]float64{-70, 45}))
	query("monq.Center(center, 5)", monq.Center([2]float64{-73.98, 40.75}, 5))

	within := monq.GeoWithin(warehousePath, monq.CenterSphere([2]float64{-73.98, 40.75}, radians))

	return showFind(ctx, coll, "monq.GeoWithin(warehouse, monq.CenterSphere(center, radians))", within, options.Find().SetProjection(geoOnly))
}

// geoStage shows $geoNear, which is the aggregation counterpart and the only way to get the distance itself back.
func geoStage(ctx context.Context, coll *mongo.Collection) error {
	step("$geoNear takes the same geometry constructors, which is why they return bare GeoJSON")

	stages := stage.Pipeline(
		stage.GeoNear(
			manhattan, "metersAway",
			stage.Spherical(),
			stage.MaxDistance(2_000_000),
			stage.GeoNearQuery(monq.Gt(ProductPaths.Price, 20)),
			stage.GeoNearKey(warehousePath),
		),
		stage.Project(stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field("metersAway", 1)),
		stage.Limit(4),
	)
	if err := showAggregate(ctx, coll, `stage.GeoNear(point, "metersAway", stage.Spherical(), stage.MaxDistance(2000000), ...)`, stages); err != nil {
		return err
	}

	fmt.Println("\n   Note: $geoNear has to be the first stage of its pipeline, which is a MongoDB rule monq does not enforce.")

	return nil
}
