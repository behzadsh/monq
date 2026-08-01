package stage

import "go.mongodb.org/mongo-driver/v2/bson"

// Count returns a stage that replaces the input documents with a single document holding their count.
//
// $count counts the documents reaching it and emits one document, {field: n}, discarding everything else. Field is
// an output name rather than a path: it cannot contain a dot, cannot start with a $, and cannot be empty. The
// stage after it sees that one small document and nothing of the original shape.
//
// Example:
//
//	stage.Count("active_users")
//	// bson.D{{Key: "$count", Value: "active_users"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/count/
func Count(field string) bson.D {
	return bson.D{{Key: "$count", Value: field}}
}

// Limit returns a stage that passes on at most n documents.
//
// $limit stops the pipeline after n documents have gone through. Which n it keeps depends on the order they arrive
// in, so a $limit without a preceding $sort takes whatever the storage engine hands over first. Placing it right
// after a $sort lets MongoDB combine the two into a top-n sort that does not buffer the whole input.
//
// Example:
//
//	stage.Limit(20)
//	// bson.D{{Key: "$limit", Value: int64(20)}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/limit/
func Limit(n int64) bson.D {
	return bson.D{{Key: "$limit", Value: n}}
}

// Match returns a stage that keeps only the documents matching filter.
//
// $match is where the filter operators of the root monq package enter a pipeline, and it takes the same documents
// that Find does. Putting it as early as possible is what lets the query planner use indexes for it.
//
// A few operators do not work here. $near and $nearSphere are not allowed in $match at all: use the $geoNear stage
// instead, which has to be the first stage of the pipeline. $text works only when the $match is itself the very
// first stage. Neither restriction is checked here; the server reports them at execution time.
//
// Example:
//
//	stage.Match(monq.Eq("status", "active"))
//	// bson.D{{Key: "$match", Value: bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/match/
func Match(filter bson.D) bson.D {
	return bson.D{{Key: "$match", Value: filter}}
}

// Sample returns a stage that passes on size documents picked at random.
//
// $sample draws documents pseudo-randomly, so two runs over unchanged data give different results. When it is the
// first stage, size is under 5% of the collection, and the collection holds more than 100 documents, MongoDB uses
// a random cursor and may return the same document more than once; otherwise it sorts the input, which costs
// memory. A collection smaller than size yields every document.
//
// Example:
//
//	stage.Sample(10)
//	// bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: int64(10)}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sample/
func Sample(size int64) bson.D {
	return bson.D{{Key: "$sample", Value: bson.D{{Key: "size", Value: size}}}}
}

// Skip returns a stage that drops the first n documents.
//
// $skip discards n documents and passes on the rest, in whatever order they arrive, so it only means something
// definite after a $sort. Paging deep into a result set with $skip stays expensive no matter the index, since the
// skipped documents are still read; a range condition on the sort key is the cheaper way to page.
//
// Example:
//
//	stage.Skip(40)
//	// bson.D{{Key: "$skip", Value: int64(40)}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/skip/
func Skip(n int64) bson.D {
	return bson.D{{Key: "$skip", Value: n}}
}

// Sort returns a stage that orders documents by the given sort document.
//
// $sort takes a document of field-to-direction pairs, 1 for ascending and -1 for descending, applied in the order
// they appear, which is why the argument is a bson.D and not a map. Sorting without an index buffers the input and
// fails past 100 MB unless the aggregation is allowed to use disk. A monq sort document built by monq.Sort drops
// straight in here once that helper lands.
//
// Example:
//
//	stage.Sort(bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: 1}})
//	// bson.D{{Key: "$sort", Value: bson.D{{Key: "created_at", Value: -1}, {Key: "_id", Value: 1}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sort/
func Sort(spec bson.D) bson.D {
	return bson.D{{Key: "$sort", Value: spec}}
}
