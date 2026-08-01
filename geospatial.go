package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// GeoJSON returns a bare GeoJSON object of the given type.
//
// It is the escape hatch of the geometry constructors, for the GeoJSON types monq has no dedicated function for
// ("LineString", "MultiPoint", "MultiLineString", "MultiPolygon", "GeometryCollection"). Coordinates are passed
// through untouched and must already follow the nesting that type requires. Positions are ordered
// [longitude, latitude], the opposite of how coordinates are usually spoken aloud.
//
// The result is a bare GeoJSON object, so wrap it in [Geometry] before handing it to a query operator.
//
// Example:
//
//	monq.GeoJSON("LineString", bson.A{bson.A{-73.97, 40.77}, bson.A{-73.88, 40.78}})
//	// bson.D{{Key: "type", Value: "LineString"}, {Key: "coordinates", Value: bson.A{...}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/geojson/
func GeoJSON(geoType string, coordinates any) bson.D {
	return bson.D{{Key: "type", Value: geoType}, {Key: "coordinates", Value: coordinates}}
}

// Point returns a bare GeoJSON Point at the given longitude and latitude.
//
// The argument order is longitude first, then latitude, matching GeoJSON's [longitude, latitude] position order and
// not the "40.7, -73.9" order coordinates are usually quoted in. Longitude runs -180 to 180 and latitude -90 to 90;
// values outside those ranges are rejected by the server, not here.
//
// The result is a bare GeoJSON object, so wrap it in [Geometry] for [GeoWithin], [GeoIntersects], [Near], and
// [NearSphere]. Stages that take a bare point take it as is.
//
// Example:
//
//	monq.Point(-73.97, 40.77)
//	// bson.D{{Key: "type", Value: "Point"}, {Key: "coordinates", Value: bson.A{-73.97, 40.77}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/geojson/#point
func Point(lon, lat float64) bson.D {
	return GeoJSON("Point", bson.A{lon, lat})
}

// Polygon returns a bare GeoJSON Polygon built from the given rings.
//
// The first ring is the outer boundary and every later ring is a hole inside it. Each ring is a closed loop, so its
// last position must repeat its first; MongoDB rejects unclosed rings. Positions are ordered
// [longitude, latitude]. This builds the GeoJSON Polygon, which every geo query operator accepts. The legacy
// $polygon shape, usable only with [GeoWithin] and on a flat plane, has no monq function: build it with [Raw].
//
// The result is a bare GeoJSON object, so wrap it in [Geometry] before handing it to a query operator.
//
// Example:
//
//	monq.Polygon([][2]float64{{0, 0}, {3, 0}, {3, 3}, {0, 3}, {0, 0}})
//	// bson.D{{Key: "type", Value: "Polygon"}, {Key: "coordinates", Value: bson.A{bson.A{
//	//     bson.A{0.0, 0.0}, bson.A{3.0, 0.0}, bson.A{3.0, 3.0}, bson.A{0.0, 3.0}, bson.A{0.0, 0.0},
//	// }}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/geojson/#polygon
func Polygon(rings ...[][2]float64) bson.D {
	coordinates := make(bson.A, len(rings))
	for i, ring := range rings {
		positions := make(bson.A, len(ring))
		for j, p := range ring {
			positions[j] = position(p)
		}

		coordinates[i] = positions
	}

	return GeoJSON("Polygon", coordinates)
}

// Geometry wraps a bare GeoJSON object as the $geometry argument of a geospatial query operator.
//
// MongoDB spells GeoJSON arguments as {$geometry: {type: ..., coordinates: ...}}, and monq keeps that split: the
// geometry constructors ([Point], [Polygon], [GeoJSON]) build the bare object, and Geometry adds the $geometry key
// for query operators. Places that want the bare object instead, such as the near field of the $geoNear stage,
// take a constructor's output directly.
//
// Example:
//
//	monq.Geometry(monq.Point(-73.97, 40.77))
//	// bson.D{{Key: "$geometry", Value: bson.D{
//	//     {Key: "type", Value: "Point"}, {Key: "coordinates", Value: bson.A{-73.97, 40.77}},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/geometry/
func Geometry(geoJSON bson.D) bson.D {
	return bson.D{{Key: "$geometry", Value: geoJSON}}
}

// Box returns the legacy $box shape spanning the two given corners.
//
// $box selects everything inside the rectangle whose opposite corners are bottomLeft and topRight, given as
// [longitude, latitude] pairs. It is a legacy shape: it works only with [GeoWithin], it measures on a flat plane
// rather than a sphere, and it needs legacy coordinate pairs or a 2d index. Handing it to [GeoIntersects], [Near],
// or [NearSphere] produces a query the server rejects.
//
// Example:
//
//	monq.Box([2]float64{0, 0}, [2]float64{3, 3})
//	// bson.D{{Key: "$box", Value: bson.A{bson.A{0.0, 0.0}, bson.A{3.0, 3.0}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/box/
func Box(bottomLeft, topRight [2]float64) bson.D {
	return bson.D{{Key: "$box", Value: bson.A{position(bottomLeft), position(topRight)}}}
}

// Center returns the legacy $center circle around the given center point.
//
// $center selects everything inside a circle of the given radius on a flat plane, with the radius expressed in the
// same units as the stored coordinates (degrees for longitude and latitude data), not in meters. Use
// [CenterSphere] for real distances on the Earth's surface. Like [Box] it is a legacy shape and works only with
// [GeoWithin].
//
// Example:
//
//	monq.Center([2]float64{-73.97, 40.77}, 0.5)
//	// bson.D{{Key: "$center", Value: bson.A{bson.A{-73.97, 40.77}, 0.5}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/center/
func Center(center [2]float64, radius float64) bson.D {
	return bson.D{{Key: "$center", Value: bson.A{position(center), radius}}}
}

// CenterSphere returns the legacy $centerSphere circle around the given center point.
//
// $centerSphere selects everything inside a circle measured on a sphere, which makes it the legacy shape to reach
// for with real-world coordinates. The radius is in radians, not meters: divide the distance in kilometers by
// 6378.1, or the distance in miles by 3963.2. Unlike the other legacy shapes it accepts both GeoJSON and legacy
// coordinate pairs, but it still works only with [GeoWithin].
//
// Example:
//
//	monq.CenterSphere([2]float64{-73.97, 40.77}, 0.00078)
//	// bson.D{{Key: "$centerSphere", Value: bson.A{bson.A{-73.97, 40.77}, 0.00078}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/centerSphere/
func CenterSphere(center [2]float64, radians float64) bson.D {
	return bson.D{{Key: "$centerSphere", Value: bson.A{position(center), radians}}}
}

// GeoWithin returns a filter that matches documents whose field lies entirely inside shape.
//
// $geoWithin matches geometries fully contained by the given shape, with no sorting and no index requirement,
// though a 2dsphere or 2d index makes it far faster. It is the one geo operator that takes either family of
// shapes: [Geometry] for GeoJSON, or the legacy [Box], [Center], and [CenterSphere]. GeoJSON shapes are measured
// on a sphere, legacy ones on a flat plane, so the same coordinates can give different results.
//
// Example:
//
//	monq.GeoWithin("loc", monq.Geometry(monq.Polygon([][2]float64{{0, 0}, {3, 0}, {3, 3}, {0, 0}})))
//	// bson.D{{Key: "loc", Value: bson.D{{Key: "$geoWithin", Value: bson.D{{Key: "$geometry", Value: ...}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/geoWithin/
func GeoWithin(field FieldPath, shape bson.D) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$geoWithin", Value: shape}}}}
}

// GeoIntersects returns a filter that matches documents whose field intersects the given geometry.
//
// $geoIntersects matches geometries that share any point with the argument, so a shape that merely touches or
// overlaps qualifies where [GeoWithin] would need full containment. It accepts GeoJSON only, meaning [Geometry]
// output; the legacy [Box], [Center], and [CenterSphere] shapes are not valid here and produce a server error.
//
// Example:
//
//	monq.GeoIntersects("area", monq.Geometry(monq.Point(-73.97, 40.77)))
//	// bson.D{{Key: "area", Value: bson.D{{Key: "$geoIntersects", Value: bson.D{{Key: "$geometry", Value: ...}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/geoIntersects/
func GeoIntersects(field FieldPath, geometry bson.D) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$geoIntersects", Value: geometry}}}}
}

// NearOption configures an optional distance bound of a [Near] or [NearSphere] filter.
type NearOption func(*bson.D)

// MaxDistance returns a [NearOption] that drops results farther than the given distance from the query point.
//
// The distance is in meters for the GeoJSON form monq builds. Legacy coordinate pairs measure it in radians
// instead, but that form is not reachable through [Near].
func MaxDistance(meters float64) NearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$maxDistance", Value: meters})
	}
}

// MinDistance returns a [NearOption] that drops results closer than the given distance to the query point.
//
// The distance is in meters, the same unit as [MaxDistance], and the two combine into a ring around the query
// point.
func MinDistance(meters float64) NearOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$minDistance", Value: meters})
	}
}

// Near returns a filter that matches documents near the given geometry, closest first.
//
// $near sorts its results by distance from the query point, which no other query operator does, and it requires a
// geospatial index on the field. That special status comes with limits: it cannot be wrapped in [Not] and cannot
// sit inside an [Or] or [Nor] branch. Geometry is GeoJSON only, meaning [Geometry] output; on a 2dsphere index
// distances are measured on the sphere in meters. The legacy form, where the coordinate pair and $maxDistance are
// siblings, has no monq function: build it with [Raw].
//
// Example:
//
//	monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MaxDistance(1000))
//	// bson.D{{Key: "loc", Value: bson.D{{Key: "$near", Value: bson.D{
//	//     {Key: "$geometry", Value: ...}, {Key: "$maxDistance", Value: 1000.0},
//	// }}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/near/
func Near(field FieldPath, geometry bson.D, opts ...NearOption) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$near", Value: withNearOptions(geometry, opts)}}}}
}

// NearSphere returns a filter that matches documents near the given geometry using spherical geometry.
//
// $nearSphere behaves like [Near] and carries the same restrictions (sorted results, geo index required, not
// usable inside [Not], [Or], or [Nor]), the difference being that it always measures on a sphere. With a 2dsphere
// index and GeoJSON that is what [Near] does too, so the two differ only on a 2d index and legacy coordinates.
//
// Example:
//
//	monq.NearSphere("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MaxDistance(1000))
//	// bson.D{{Key: "loc", Value: bson.D{{Key: "$nearSphere", Value: bson.D{
//	//     {Key: "$geometry", Value: ...}, {Key: "$maxDistance", Value: 1000.0},
//	// }}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/nearSphere/
func NearSphere(field FieldPath, geometry bson.D, opts ...NearOption) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$nearSphere", Value: withNearOptions(geometry, opts)}}}}
}

// withNearOptions copies geometry before applying opts, so appending an option never writes into the caller's
// document.
func withNearOptions(geometry bson.D, opts []NearOption) bson.D {
	doc := make(bson.D, len(geometry))
	copy(doc, geometry)

	for _, opt := range opts {
		opt(&doc)
	}

	return doc
}

// position converts a [longitude, latitude] pair into the bson.A form MongoDB expects.
func position(p [2]float64) bson.A {
	return bson.A{p[0], p[1]}
}
