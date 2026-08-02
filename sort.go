package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Asc returns a sort entry ordering documents by field from smallest to largest.
//
// A direction of 1 is ascending, which for strings means alphabetical, for dates oldest first, and across types
// follows BSON comparison order. Pass it to [Sort] to build a full sort document, or use it on its own where a
// single-field sort is wanted.
//
// Example:
//
//	monq.Asc("created_at")
//	// bson.D{{Key: "created_at", Value: 1}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/method/cursor.sort/
func Asc(field FieldPath) bson.D {
	return bson.D{{Key: string(field), Value: 1}}
}

// Desc returns a sort entry ordering documents by field from largest to smallest.
//
// A direction of -1 is descending, which is what puts the newest dates or the highest scores first.
//
// Example:
//
//	monq.Desc("created_at")
//	// bson.D{{Key: "created_at", Value: -1}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/method/cursor.sort/
func Desc(field FieldPath) bson.D {
	return bson.D{{Key: string(field), Value: -1}}
}

// Sort returns a sort document ordering documents by the given entries.
//
// Sorting applies the entries in the order they are given, each one breaking the ties left by the ones before it,
// which is why a sort document is a bson.D and not a map. Entries come from [Asc], [Desc], and [TextScore], and a
// field named twice keeps its first position with the later direction, the same rule [Update] follows.
//
// The result goes into options.Find().SetSort or into the $sort stage. A sort with no index behind it buffers its
// input, and MongoDB gives up past 100 MB unless the query is allowed to use disk. Adding a unique field such as
// _id as the last entry makes the order total, which is what keeps paging stable.
//
// Example:
//
//	monq.Sort(monq.Desc("created_at"), monq.Asc("_id"))
//	// bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: 1}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/method/cursor.sort/
func Sort(entries ...bson.D) bson.D {
	sort := bson.D{}
	for _, entry := range entries {
		for _, e := range entry {
			if i := indexOfKey(sort, e.Key); i >= 0 {
				sort[i].Value = e.Value

				continue
			}

			sort = append(sort, e)
		}
	}

	return sort
}

// TextScore returns a sort entry ordering documents by how well they matched a text search.
//
// It sorts on the relevance score $text computes, best first, and is only meaningful in a query that runs one. The
// field is an output name rather than a stored path: it names where the score would appear, and by convention is
// called "score". Sorting by it without a $text search in the same query is an error.
//
// Example:
//
//	monq.Sort(monq.TextScore("score"))
//	// bson.D{{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/meta/
func TextScore(field FieldPath) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$meta", Value: "textScore"}}}}
}
