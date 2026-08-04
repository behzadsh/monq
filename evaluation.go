package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Expr returns a filter that matches documents for which the aggregation expression evaluates to true.
//
// $expr embeds an aggregation expression in a query filter, which is what lets a query compare two fields of the
// same document, something plain field conditions cannot express. Field references inside the expression use the
// aggregation "$field" string form rather than a bare field name, and index support is narrower than for plain
// field conditions: $expr cannot use a multikey index at all. The expression is passed through untouched, so a
// malformed one surfaces as a server error at execution time.
//
// Example:
//
//	monq.Expr(bson.D{{Key: "$gt", Value: bson.A{"$spent", "$budget"}}})
//	// bson.D{{Key: "$expr", Value: bson.D{{Key: "$gt", Value: bson.A{"$spent", "$budget"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/expr/
func Expr(expression any) bson.D {
	return bson.D{{Key: "$expr", Value: expression}}
}

// JSONSchema returns a filter that matches documents satisfying the given JSON Schema.
//
// $jsonSchema validates documents against a JSON Schema, the same schema language collection validators use. It
// covers the draft 4 keywords MongoDB implements plus its own bsonType extension, and ignores keywords it does not
// know. Like every other monq operator it validates nothing on its own: a schema MongoDB rejects produces a server
// error, not a monq error.
//
// Example:
//
//	monq.JSONSchema(bson.D{{Key: "required", Value: bson.A{"email"}}})
//	// bson.D{{Key: "$jsonSchema", Value: bson.D{{Key: "required", Value: bson.A{"email"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/jsonSchema/
func JSONSchema(schema bson.D) bson.D {
	return bson.D{{Key: "$jsonSchema", Value: schema}}
}

// Mod returns a filter that matches documents where field divided by divisor leaves the given remainder.
//
// $mod matches when the field's value modulo divisor equals remainder. Only numeric values are considered, and
// they are truncated toward zero before the division, so 4.7 behaves as 4. A divisor of 0 is an error the server
// reports at execution time, not a monq error.
//
// Example:
//
//	monq.Mod("qty", 4, 0)
//	// bson.D{{Key: "qty", Value: bson.D{{Key: "$mod", Value: bson.A{int64(4), int64(0)}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/mod/
func Mod(field FieldPath, divisor, remainder int64) bson.D {
	return bson.D{{Key: string(field), Value: bson.D{{Key: "$mod", Value: bson.A{divisor, remainder}}}}}
}

// Regex returns a filter that matches documents where field matches the regular expression pattern.
//
// $regex matches string values against pattern, applying options as $options flags: "i" for case-insensitivity,
// "m" for multiline (^ and $ match line boundaries), "x" to ignore unescaped whitespace and # comments in pattern,
// and "s" so "." matches newline characters too. An empty options string means no flags, not an error. Unless
// pattern is anchored to the start of the string (e.g. "^prefix") and field is indexed, $regex cannot use an index
// efficiently and scans every document.
//
// Example:
//
//	monq.Regex("email", "^alice", "i")
//	// bson.D{{Key: "email", Value: bson.D{{Key: "$regex", Value: "^alice"}, {Key: "$options", Value: "i"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/regex/
func Regex(field FieldPath, pattern, options string) bson.D {
	return bson.D{
		{Key: string(field), Value: bson.D{{Key: "$regex", Value: pattern}, {Key: "$options", Value: options}}},
	}
}

// SampleRate returns a filter that keeps a random fraction of the documents it sees.
//
// $sampleRate decides per document, with the given probability, so the number kept varies from run to run and only
// approaches the fraction asked for over a large collection. A rate of 0 matches nothing and 1 matches everything;
// anything outside that range is a server error. It is a filter rather than a stage, so it works in Find and in a
// $match, which is what separates it from the $sample stage: $sample takes an exact count and needs a whole pass or
// an index scan, while this one is a cheap coin flip per document with no guaranteed count.
//
// Being random, it is not stable across executions. The aggregation counterpart for a random number in an
// expression is the $rand operator in monq/expr.
//
// Example:
//
//	monq.SampleRate(0.33)
//	// bson.D{{Key: "$sampleRate", Value: 0.33}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/sampleRate/
func SampleRate(rate float64) bson.D {
	return bson.D{{Key: "$sampleRate", Value: rate}}
}

// TextOption configures one optional field of a [Text] filter.
type TextOption func(*bson.D)

// Language returns a [TextOption] that sets the language used for stemming and stop words.
//
// The name comes from the languages the text index supports ("english", "fr", ...). Pass "none" to skip stemming
// and stop-word removal and use simple tokenization. Without this option the text index's default language wins.
func Language(language string) TextOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$language", Value: language})
	}
}

// CaseSensitive returns a [TextOption] that makes the search case-sensitive.
//
// $text is case-insensitive by default. Turning this on bypasses the index's case-insensitive prefixes, so the
// search does more work.
func CaseSensitive() TextOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$caseSensitive", Value: true})
	}
}

// DiacriticSensitive returns a [TextOption] that makes the search diacritic-sensitive.
//
// $text ignores diacritics by default, so "cafe" matches "café". Turning this on treats the two as different terms.
func DiacriticSensitive() TextOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "$diacriticSensitive", Value: true})
	}
}

// Text returns a filter that runs a text search against the collection's text index.
//
// $text tokenizes search and matches documents containing any of the resulting terms. A double-quoted substring is
// matched as a phrase, and a term prefixed with "-" excludes documents containing it. Unlike every other filter
// monq builds, $text is a top-level operator rather than a field expression, which has consequences: it cannot be
// wrapped in [Not], a query may hold at most one $text, and it cannot sit inside an [Or] or [Nor] branch. The
// collection needs a text index or the query fails.
//
// Example:
//
//	monq.Text("coffee shop", monq.Language("en"), monq.CaseSensitive())
//	// bson.D{{Key: "$text", Value: bson.D{
//	//     {Key: "$search", Value: "coffee shop"},
//	//     {Key: "$language", Value: "en"},
//	//     {Key: "$caseSensitive", Value: true},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/query/text/
func Text(search string, opts ...TextOption) bson.D {
	doc := bson.D{{Key: "$search", Value: search}}
	for _, opt := range opts {
		opt(&doc)
	}

	return bson.D{{Key: "$text", Value: doc}}
}
