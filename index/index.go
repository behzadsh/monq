package index

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
)

// Option is one part of an index model, either a key or a property of the index as a whole.
//
// Keys and options share a type so that [New] takes them in one list, in the order they are written. That ordering
// is what makes a compound index a compound index: {email: 1, created_at: -1} is a different index from
// {created_at: -1, email: 1}, and only the first can answer a query on email alone.
type Option func(*model)

// model gathers what an [Option] contributes, until [New] hands it over as a mongo.IndexModel.
type model struct {
	keys    bson.D
	options *options.IndexOptionsBuilder
}

// Asc returns an [Option] adding an ascending key on field.
//
// Ascending and descending matter only for compound indexes, where the directions decide which sort orders the
// index can serve without a further sort. A single-field index can be read either way.
//
// Example:
//
//	index.New(index.Asc("email"))
//	// mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/indexes/
func Asc(field monq.FieldPath) Option {
	return key(field, 1)
}

// Desc returns an [Option] adding a descending key on field.
//
// Example:
//
//	index.New(index.Asc("email"), index.Desc("created_at"))
//	// mongo.IndexModel{Keys: bson.D{{Key: "email", Value: 1}, {Key: "created_at", Value: -1}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/indexes/
func Desc(field monq.FieldPath) Option {
	return key(field, -1)
}

// Geo2D returns an [Option] adding a legacy 2d key on field.
//
// A 2d index covers legacy coordinate pairs on a flat plane, which is what the $box, $center, and $polygon shapes
// need. Data in GeoJSON belongs under [Geo2DSphere] instead.
//
// Example:
//
//	index.New(index.Geo2D("loc"))
//	// mongo.IndexModel{Keys: bson.D{{Key: "loc", Value: "2d"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/2d/
func Geo2D(field monq.FieldPath) Option {
	return key(field, "2d")
}

// Geo2DSphere returns an [Option] adding a 2dsphere key on field.
//
// A 2dsphere index measures on the Earth's surface and is what $near, $geoWithin, and $geoIntersects want for
// GeoJSON data. The $geoNear stage needs one of these or a 2d index to exist at all.
//
// Example:
//
//	index.New(index.Geo2DSphere("loc"))
//	// mongo.IndexModel{Keys: bson.D{{Key: "loc", Value: "2dsphere"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/2dsphere/
func Geo2DSphere(field monq.FieldPath) Option {
	return key(field, "2dsphere")
}

// Hashed returns an [Option] adding a hashed key on field.
//
// A hashed index stores the hash of a value rather than the value, which spreads writes evenly across a sharded
// cluster. It answers equality only: range queries and sorts cannot use it.
//
// Example:
//
//	index.New(index.Hashed("user_id"))
//	// mongo.IndexModel{Keys: bson.D{{Key: "user_id", Value: "hashed"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/indexes/index-types/index-hashed/
func Hashed(field monq.FieldPath) Option {
	return key(field, "hashed")
}

// Hidden returns an [Option] hiding the index from the query planner.
//
// A hidden index is still built and still maintained on every write, but no query uses it. That makes it the safe
// way to find out whether dropping an index would hurt, since unhiding it is instant while rebuilding is not.
func Hidden() Option {
	return func(m *model) {
		m.options.SetHidden(true)
	}
}

// Name returns an [Option] naming the index.
//
// Without one MongoDB builds a name from the keys, such as "email_1_created_at_-1", which is fine until it grows
// past the length the server allows. A name is also what dropIndex takes.
func Name(name string) Option {
	return func(m *model) {
		m.options.SetName(name)
	}
}

// New returns the index model described by opts.
//
// Keys are taken in the order they are given, since that order decides which queries the index can serve. Options
// that are not keys may appear anywhere in the list. Nothing is validated here: an index with no keys at all, or
// one naming the same field twice, is reported by the server when the index is created.
//
// Example:
//
//	index.New(index.Asc("email"), index.Unique())
//	// mongo.IndexModel{
//	//     Keys:    bson.D{{Key: "email", Value: 1}},
//	//     Options: options.Index().SetUnique(true),
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/method/db.collection.createIndex/
func New(opts ...Option) mongo.IndexModel {
	m := model{keys: bson.D{}, options: options.Index()}
	for _, opt := range opts {
		opt(&m)
	}

	return mongo.IndexModel{Keys: m.keys, Options: m.options}
}

// PartialFilter returns an [Option] indexing only the documents matching filter.
//
// A partial index is smaller and cheaper to maintain than one covering the whole collection, and it is how a
// unique constraint gets applied to a subset, such as unique emails among the accounts that are not deleted. The
// query planner uses it only when the query implies the filter, so a query without that condition falls back to a
// collection scan. The filter is an ordinary query document, built with the root monq package.
//
// Example:
//
//	index.New(index.Asc("email"), index.Unique(), index.PartialFilter(monq.Exists("deleted_at", false)))
//	// the unique constraint applies only to documents without a deleted_at field
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/index-partial/
func PartialFilter(filter bson.D) Option {
	return func(m *model) {
		m.options.SetPartialFilterExpression(filter)
	}
}

// Sparse returns an [Option] indexing only the documents that have the field.
//
// A sparse index leaves out documents missing the indexed field entirely, where an ordinary index stores them
// under null. That makes it smaller, and it makes a sort on the field skip those documents rather than listing
// them. [PartialFilter] does the same job with a condition of your own and is the one MongoDB now recommends.
func Sparse() Option {
	return func(m *model) {
		m.options.SetSparse(true)
	}
}

// Text returns an [Option] adding a text key on field.
//
// Text keys are what the $text query operator searches. A collection may hold only one text index, though that one
// index may cover several fields, so more than one [Text] in the same call is the usual way to search across them.
//
// Example:
//
//	index.New(index.Text("title"), index.Text("body"))
//	// mongo.IndexModel{Keys: bson.D{{Key: "title", Value: "text"}, {Key: "body", Value: "text"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/indexes/index-types/index-text/
func Text(field monq.FieldPath) Option {
	return key(field, "text")
}

// TTL returns an [Option] deleting documents once the indexed date is that far in the past.
//
// A TTL index works on a single field holding a date, and a background task removes the documents some time after
// they expire, so deletion is eventual rather than exact. MongoDB counts in whole seconds: a duration is taken as
// the number of seconds it holds, and anything finer is dropped. A duration of 0 expires documents at the date
// itself, which is how an expiry timestamp stored in the document is honored.
//
// Example:
//
//	index.New(index.Asc("created_at"), index.TTL(24*time.Hour))
//	// mongo.IndexModel{
//	//     Keys:    bson.D{{Key: "created_at", Value: 1}},
//	//     Options: options.Index().SetExpireAfterSeconds(86400),
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/core/index-ttl/
func TTL(d time.Duration) Option {
	return func(m *model) {
		m.options.SetExpireAfterSeconds(int32(d.Seconds()))
	}
}

// Unique returns an [Option] rejecting documents that repeat an indexed value.
//
// On a compound index the whole key has to be unique, not each field on its own. Documents missing the field count
// as holding null, so only one of them can exist unless [PartialFilter] or [Sparse] leaves them out. Creating the
// index fails outright if the collection already holds duplicates.
func Unique() Option {
	return func(m *model) {
		m.options.SetUnique(true)
	}
}

// key returns an [Option] appending one key to the index.
func key(field monq.FieldPath, value any) Option {
	return func(m *model) {
		m.keys = append(m.keys, bson.E{Key: string(field), Value: value})
	}
}
