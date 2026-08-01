package stage

import (
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

// GeoNearOption configures one optional field of a [GeoNear] stage.
type GeoNearOption func(*bson.D)

// DistanceMultiplier returns a [GeoNearOption] that scales every computed distance by factor.
//
// It is how a distance in radians becomes one in a real unit: multiply by 6378.1 for kilometers or 3963.2 for
// miles. On a 2dsphere index, where distances already come out in meters, dividing by 1000 gives kilometers.
func DistanceMultiplier(factor float64) GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "distanceMultiplier", Value: factor})
	}
}

// GeoNearKey returns a [GeoNearOption] naming which geospatial index to use.
//
// A collection with two or more geospatial indexes leaves $geoNear unable to choose, and the aggregation fails
// until this option says which field to search on. With a single index it can be left out.
func GeoNearKey(field monq.FieldPath) GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "key", Value: string(field)})
	}
}

// GeoNearQuery returns a [GeoNearOption] that keeps only the documents matching filter.
//
// $geoNear has to be the first stage of a pipeline, so there is no room for a [Match] before it; this option is
// where that condition goes instead. The filter is an ordinary query document.
func GeoNearQuery(filter bson.D) GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "query", Value: filter})
	}
}

// IncludeLocs returns a [GeoNearOption] that records which stored location matched.
//
// The name is an output field name. It matters when a document holds several locations in an array, since the
// distance alone does not say which of them was the closest.
func IncludeLocs(name string) GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "includeLocs", Value: name})
	}
}

// MaxDistance returns a [GeoNearOption] that drops documents farther than the given distance from the query point.
//
// The unit follows the same rule as the computed distance: meters on a 2dsphere index, radians on a 2d index with
// legacy coordinates. The root package has a [github.com/behzadsh/monq.MaxDistance] of its own for the $near query
// operator; the two are separate option types, so the compiler rejects the wrong one.
func MaxDistance(distance float64) GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "maxDistance", Value: distance})
	}
}

// MinDistance returns a [GeoNearOption] that drops documents closer than the given distance to the query point.
//
// The unit follows the same rule as [MaxDistance], and the two together describe a ring around the query point.
func MinDistance(distance float64) GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "minDistance", Value: distance})
	}
}

// Spherical returns a [GeoNearOption] that measures distances on a sphere.
//
// A 2dsphere index always measures that way and ignores this option. On a 2d index it switches the calculation
// from flat plane geometry to spherical, which is what real-world coordinates need.
func Spherical() GeoNearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "spherical", Value: true})
	}
}

// GeoNear returns a stage that orders documents by distance from a point and records that distance.
//
// $geoNear finds the documents nearest the given point, closest first, and writes each one's distance into
// distanceField. It is the aggregation counterpart of the $near query operator, and like it the collection needs a
// geospatial index; with more than one, [GeoNearKey] has to say which.
//
// Two rules set it apart from every other stage. It has to be the first stage of the pipeline, which is why the
// condition that would otherwise be a [Match] before it goes into [GeoNearQuery] instead. And the unit of the
// distance depends on the index: meters on a 2dsphere index, radians on a 2d index with legacy coordinates, with
// [DistanceMultiplier] to convert either into what the application wants.
//
// The point is a bare GeoJSON object, so [github.com/behzadsh/monq.Point] output goes straight in, as does a
// legacy coordinate pair on a 2d index.
//
// Example:
//
//	stage.GeoNear(monq.Point(-73.97, 40.77), "distance", stage.MaxDistance(1000), stage.Spherical())
//	// bson.D{{Key: "$geoNear", Value: bson.D{
//	//     {Key: "near", Value: bson.D{{Key: "type", Value: "Point"}, {Key: "coordinates", Value: ...}}},
//	//     {Key: "distanceField", Value: "distance"},
//	//     {Key: "maxDistance", Value: 1000.0},
//	//     {Key: "spherical", Value: true},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/geoNear/
func GeoNear(near any, distanceField string, opts ...GeoNearOption) bson.D {
	spec := bson.D{
		{Key: "near", Value: near},
		{Key: "distanceField", Value: distanceField},
	}

	for _, opt := range opts {
		opt(&spec)
	}

	return bson.D{{Key: "$geoNear", Value: spec}}
}
