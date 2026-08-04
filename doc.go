// Package monq is a composable, discoverable query-building DSL for the official MongoDB Go driver.
//
// Writing MongoDB queries by hand means nesting bson.D, bson.A, and bson.E values and remembering operator strings like
// $eq, $gte, or $elemMatch. monq replaces that with plain functions named after the Mongo operators they build, so IDE
// autocomplete and godoc double as the MongoDB operator reference:
//
//	filter := monq.And(
//		monq.Eq("status", "active"),
//		monq.Or(
//			monq.Gte("stats.followers", 10000),
//			monq.Exists("verified_at", true),
//		),
//	)
//
// This package holds the query operators, the update operators, and the sort helpers. Filters compose by nesting, and
// each one is a complete filter document on its own, so [Eq] can go straight into Find without any wrapper.
//
// Update operators work the same way, one field each: [Set], [Inc], [Push], and the rest each return a finished update
// document. [Update] is what puts several together, merging the ones that share an operator key, since two $set
// documents concatenated by hand would collide. It is to updates what [And] is to filters.
//
// Sort documents come from [Sort] with [Asc], [Desc], and [TextScore] entries, and keep the order they are given,
// which is the order MongoDB applies them in.
//
// Geospatial queries come with their geometry, so GeoJSON never has to be written by hand: [Point], [Polygon], and
// [GeoJSON] build the shape, [Geometry] hands it to [GeoWithin], [GeoIntersects], [Near], or [NearSphere], and the
// legacy [Box], [Center], and [CenterSphere] shapes are there for 2d data. Every position is ordered
// [longitude, latitude], the reverse of the order coordinates are usually quoted in.
//
// Three subpackages cover the rest of MongoDB's surface, split out because MongoDB reuses names across contexts and
// the package qualifier is what tells them apart: monq/stage builds aggregation pipeline stages, where stage.Set is
// the $set stage rather than the $set update operator; monq/expr builds the aggregation expressions those stages
// compute with, where expr.Eq compares two expressions rather than testing a field; and monq/index builds index
// models, kept separate so that programs which only build queries do not compile the driver's mongo package and its
// dependencies.
//
// monq is not an ODM: there are no models, no sessions, and no query execution. Every function returns a raw driver
// value (bson.D) that plugs directly into Find, Aggregate, UpdateOne, and friends with zero adapter layer. Nothing is
// validated either, so a malformed query is reported by MongoDB rather than by monq. Operators monq does not cover
// can be dropped in anywhere via [Raw], so adoption can be partial.
package monq
