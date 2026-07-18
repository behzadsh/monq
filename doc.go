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
// monq is not an ODM: there are no models, no sessions, and no query execution. Every function returns a raw driver
// value (bson.D) that plugs directly into Find, Aggregate, UpdateOne, and friends with zero adapter layer. Values monq
// does not yet cover can be dropped in anywhere via [Raw], so adoption can be partial.
package monq
