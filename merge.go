package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// mergeEntries folds one-field documents into a single ordered document.
//
// It is what [Sort] and [Projection] are built from: both take entries that each name one field, both care about
// the order those fields appear in, and both let a later entry replace an earlier one for the same field while
// keeping that field's original position.
func mergeEntries(entries []bson.D) bson.D {
	merged := bson.D{}

	for _, entry := range entries {
		for _, e := range entry {
			if i := indexOfKey(merged, e.Key); i >= 0 {
				merged[i].Value = e.Value

				continue
			}

			merged = append(merged, e)
		}
	}

	return merged
}

// indexOfKey returns the position of key in d, or -1 when d has no such key.
func indexOfKey(d bson.D, key string) int {
	for i, e := range d {
		if e.Key == key {
			return i
		}
	}

	return -1
}

// fieldKeys renders paths as the document keys they name, all sharing one value.
func fieldKeys(fields []FieldPath, value any) bson.D {
	keys := make(bson.D, len(fields))
	for i, f := range fields {
		keys[i] = bson.E{Key: string(f), Value: value}
	}

	return keys
}
