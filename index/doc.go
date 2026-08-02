// Package index builds MongoDB index models.
//
// Index keys and index options are two different documents in the driver's API, and this package puts them behind
// one call: every argument to [New] is an Option, whether it adds a key or sets a property, and the result is the
// mongo.IndexModel that Collection.Indexes().CreateOne takes.
//
//	model := index.New(
//		index.Asc("email"),
//		index.Unique(),
//	)
//
//	_, err := collection.Indexes().CreateOne(ctx, model)
//
// It lives in a package of its own for one reason: mongo.IndexModel comes from the driver's mongo package, which
// carries a handful of third-party dependencies that the rest of monq does not need. Keeping index models here
// means a program that only builds queries never compiles any of that.
//
// The rest of the driver's option surface is left alone. Anything this package does not cover is reachable by
// building the mongo.IndexModel directly, the same way [github.com/behzadsh/monq.Raw] covers query operators monq
// has no function for.
package index
