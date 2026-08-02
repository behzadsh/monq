package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// Concat returns an expression joining its arguments into one string.
//
// $concat takes strings only, and a null or missing argument makes the whole result null rather than being skipped
// over, so an optional field belongs behind an [IfNull]. Numbers are not converted; run them through ToString
// first.
//
// Example:
//
//	expr.Concat(expr.Field("first_name"), " ", expr.Field("last_name"))
//	// bson.D{{Key: "$concat", Value: bson.A{"$first_name", " ", "$last_name"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/concat/
func Concat(values ...any) bson.D {
	return bson.D{{Key: "$concat", Value: operands(values)}}
}

// IndexOfBytes returns an expression yielding the byte offset of substring within str.
//
// $indexOfBytes counts in bytes, so on text outside ASCII its result does not line up with character positions and
// cannot be handed to [SubstrCP] or [IndexOfCP]; [IndexOfCP] is the one that counts characters. The result is -1
// when the substring is absent, and null when str is null or missing.
//
// The optional trailing arguments are the byte offsets to search between: pass none to search the whole string,
// one to start there, or two for a start and an end. A start past the end, or a start greater than the end, yields
// -1.
//
// Example:
//
//	expr.IndexOfBytes(expr.Field("sku"), "-")
//	// bson.D{{Key: "$indexOfBytes", Value: bson.A{"$sku", "-"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/indexOfBytes/
func IndexOfBytes(str, substring any, startEnd ...any) bson.D {
	return bson.D{{Key: "$indexOfBytes", Value: indexOfArgs(str, substring, startEnd)}}
}

// IndexOfCP returns an expression yielding the code point offset of substring within str.
//
// $indexOfCP counts in code points, which is what lines up with [SubstrCP] and [StrLenCP] and with how a reader
// counts characters. The result is -1 when the substring is absent, and null when str is null or missing.
//
// The optional trailing arguments are the code point offsets to search between, exactly as in [IndexOfBytes]: none
// searches the whole string, one gives a start, two give a start and an end.
//
// Example:
//
//	expr.IndexOfCP(expr.Field("title"), "é")
//	// bson.D{{Key: "$indexOfCP", Value: bson.A{"$title", "é"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/indexOfCP/
func IndexOfCP(str, substring any, startEnd ...any) bson.D {
	return bson.D{{Key: "$indexOfCP", Value: indexOfArgs(str, substring, startEnd)}}
}

// Ltrim returns an expression removing characters from the start of a string.
//
// $ltrim is [Trim] applied to the left side only, and [TrimChars] means the same set of characters here: the
// argument is a set, not a prefix to match. Without it, leading whitespace is removed.
//
// Example:
//
//	expr.Ltrim(expr.Field("code"), expr.TrimChars("0"))
//	// bson.D{{Key: "$ltrim", Value: bson.D{{Key: "input", Value: "$code"}, {Key: "chars", Value: "0"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ltrim/
func Ltrim(input any, opts ...TrimOption) bson.D {
	return bson.D{{Key: "$ltrim", Value: trimSpec(input, opts)}}
}

// RegexFind returns an expression yielding the first match of a regular expression.
//
// $regexFind produces a document of {match, idx, captures}, where idx is a code point offset and captures holds
// the capture groups, or null when nothing matches. That shape is what separates it from [RegexMatch], which
// yields a boolean, and [RegexFindAll], which yields an array of these documents.
//
// The pattern is a string or a bson.Regex value. Flags go through [RegexOptions], and passing them alongside a
// bson.Regex that carries its own options is an error.
//
// Example:
//
//	expr.RegexFind(expr.Field("email"), "^[^@]+", expr.RegexOptions("i"))
//	// bson.D{{Key: "$regexFind", Value: bson.D{
//	//     {Key: "input", Value: "$email"},
//	//     {Key: "regex", Value: "^[^@]+"},
//	//     {Key: "options", Value: "i"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/regexFind/
func RegexFind(input, regex any, opts ...RegexOption) bson.D {
	return bson.D{{Key: "$regexFind", Value: regexSpec(input, regex, opts)}}
}

// RegexFindAll returns an expression yielding every match of a regular expression.
//
// $regexFindAll produces an array of the same {match, idx, captures} documents [RegexFind] returns one of, and an
// empty array rather than null when nothing matches.
//
// Example:
//
//	expr.RegexFindAll(expr.Field("body"), "#\\w+")
//	// bson.D{{Key: "$regexFindAll", Value: bson.D{
//	//     {Key: "input", Value: "$body"},
//	//     {Key: "regex", Value: "#\\w+"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/regexFindAll/
func RegexFindAll(input, regex any, opts ...RegexOption) bson.D {
	return bson.D{{Key: "$regexFindAll", Value: regexSpec(input, regex, opts)}}
}

// RegexMatch returns an expression that is true when a regular expression matches.
//
// $regexMatch yields a boolean and nothing more, which makes it the one to use in a [Cond] or a $match through
// monq.Expr. Use [RegexFind] when the matched text or its position is what is wanted.
//
// Example:
//
//	expr.RegexMatch(expr.Field("email"), "@example\\.com$")
//	// bson.D{{Key: "$regexMatch", Value: bson.D{
//	//     {Key: "input", Value: "$email"},
//	//     {Key: "regex", Value: "@example\\.com$"},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/regexMatch/
func RegexMatch(input, regex any, opts ...RegexOption) bson.D {
	return bson.D{{Key: "$regexMatch", Value: regexSpec(input, regex, opts)}}
}

// RegexOption configures the optional flags of [RegexFind], [RegexFindAll], and [RegexMatch].
type RegexOption func(*bson.D)

// RegexOptions returns a [RegexOption] carrying the regular expression flags.
//
// The flags are the usual ones: "i" for case-insensitivity, "m" for multiline anchors, "s" so a dot matches
// newlines, and "x" for extended patterns. Combining this with a bson.Regex pattern that already carries options
// is an error.
func RegexOptions(flags string) RegexOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "options", Value: flags})
	}
}

// ReplaceAll returns an expression replacing every occurrence of find within input.
//
// $replaceAll matches a plain string rather than a pattern; [RegexFindAll] is the one that takes a regular
// expression. A null input or a null find makes the result null.
//
// Example:
//
//	expr.ReplaceAll(expr.Field("phone"), "-", "")
//	// bson.D{{Key: "$replaceAll", Value: bson.D{
//	//     {Key: "input", Value: "$phone"},
//	//     {Key: "find", Value: "-"},
//	//     {Key: "replacement", Value: ""},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceAll/
func ReplaceAll(input, find, replacement any) bson.D {
	return bson.D{{Key: "$replaceAll", Value: replaceSpec(input, find, replacement)}}
}

// ReplaceOne returns an expression replacing the first occurrence of find within input.
//
// $replaceOne matches a plain string rather than a pattern and stops after the first hit. A null input or a null
// find makes the result null.
//
// Example:
//
//	expr.ReplaceOne(expr.Field("title"), "draft ", "")
//	// bson.D{{Key: "$replaceOne", Value: bson.D{
//	//     {Key: "input", Value: "$title"},
//	//     {Key: "find", Value: "draft "},
//	//     {Key: "replacement", Value: ""},
//	// }}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/replaceOne/
func ReplaceOne(input, find, replacement any) bson.D {
	return bson.D{{Key: "$replaceOne", Value: replaceSpec(input, find, replacement)}}
}

// Rtrim returns an expression removing characters from the end of a string.
//
// $rtrim is [Trim] applied to the right side only, and [TrimChars] means the same set of characters here. Without
// it, trailing whitespace is removed.
//
// Example:
//
//	expr.Rtrim(expr.Field("name"))
//	// bson.D{{Key: "$rtrim", Value: bson.D{{Key: "input", Value: "$name"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/rtrim/
func Rtrim(input any, opts ...TrimOption) bson.D {
	return bson.D{{Key: "$rtrim", Value: trimSpec(input, opts)}}
}

// Split returns an expression cutting a string into an array on a delimiter.
//
// $split removes the delimiter from the result and keeps empty pieces, so splitting "a,,b" on "," gives three
// elements. An empty delimiter is an error rather than a split into characters, and a delimiter that never occurs
// yields a one-element array holding the whole string.
//
// Example:
//
//	expr.Split(expr.Field("tags"), ",")
//	// bson.D{{Key: "$split", Value: bson.A{"$tags", ","}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/split/
func Split(str, delimiter any) bson.D {
	return bson.D{{Key: "$split", Value: bson.A{str, delimiter}}}
}

// StrLenBytes returns an expression yielding the length of a string in bytes.
//
// $strLenBytes counts encoded bytes, so "café" is 5 while [StrLenCP] gives 4. It pairs with [SubstrBytes] and
// [IndexOfBytes], which count the same way; mixing a byte count with a code point offset is where text outside
// ASCII goes wrong.
//
// Example:
//
//	expr.StrLenBytes(expr.Field("name"))
//	// bson.D{{Key: "$strLenBytes", Value: "$name"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/strLenBytes/
func StrLenBytes(str any) bson.D {
	return bson.D{{Key: "$strLenBytes", Value: str}}
}

// StrLenCP returns an expression yielding the length of a string in code points.
//
// $strLenCP counts characters as a reader would, so "café" is 4. It pairs with [SubstrCP] and [IndexOfCP].
//
// Example:
//
//	expr.StrLenCP(expr.Field("name"))
//	// bson.D{{Key: "$strLenCP", Value: "$name"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/strLenCP/
func StrLenCP(str any) bson.D {
	return bson.D{{Key: "$strLenCP", Value: str}}
}

// Strcasecmp returns an expression comparing two strings without regard to case, yielding -1, 0, or 1.
//
// $strcasecmp reports whether a sorts before, with, or after b, folding case for ASCII letters only, so accented
// letters compare by their code points. [Cmp] is the case-sensitive comparison over any type.
//
// Example:
//
//	expr.Strcasecmp(expr.Field("name"), "ada")
//	// bson.D{{Key: "$strcasecmp", Value: bson.A{"$name", "ada"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/strcasecmp/
func Strcasecmp(a, b any) bson.D {
	return bson.D{{Key: "$strcasecmp", Value: bson.A{a, b}}}
}

// SubstrBytes returns an expression taking a slice of a string by byte offsets.
//
// $substrBytes counts in bytes, and a start or a length that lands in the middle of a multi-byte character fails
// the aggregation rather than returning damaged text, which makes [SubstrCP] the safer choice for anything that
// might not be ASCII. Offsets come from [IndexOfBytes] and lengths from [StrLenBytes]; a code point offset from
// their CP counterparts does not belong here.
//
// Example:
//
//	expr.SubstrBytes(expr.Field("sku"), 0, 3)
//	// bson.D{{Key: "$substrBytes", Value: bson.A{"$sku", 0, 3}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/substrBytes/
func SubstrBytes(str, start, length any) bson.D {
	return bson.D{{Key: "$substrBytes", Value: bson.A{str, start, length}}}
}

// SubstrCP returns an expression taking a slice of a string by code point offsets.
//
// $substrCP counts characters, so it never splits one in half, and a length running past the end simply stops
// there. Offsets come from [IndexOfCP] and lengths from [StrLenCP].
//
// Example:
//
//	expr.SubstrCP(expr.Field("title"), 0, 10)
//	// bson.D{{Key: "$substrCP", Value: bson.A{"$title", 0, 10}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/substrCP/
func SubstrCP(str, start, length any) bson.D {
	return bson.D{{Key: "$substrCP", Value: bson.A{str, start, length}}}
}

// ToLower returns an expression lowercasing a string.
//
// $toLower is defined for ASCII letters; other characters come back as they went in, so it is not a substitute for
// proper case folding. A null or missing input yields an empty string rather than null.
//
// Example:
//
//	expr.ToLower(expr.Field("email"))
//	// bson.D{{Key: "$toLower", Value: "$email"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toLower/
func ToLower(str any) bson.D {
	return bson.D{{Key: "$toLower", Value: str}}
}

// ToUpper returns an expression uppercasing a string.
//
// $toUpper is defined for ASCII letters, exactly as [ToLower] is, and likewise yields an empty string for a null
// or missing input.
//
// Example:
//
//	expr.ToUpper(expr.Field("country"))
//	// bson.D{{Key: "$toUpper", Value: "$country"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/toUpper/
func ToUpper(str any) bson.D {
	return bson.D{{Key: "$toUpper", Value: str}}
}

// Trim returns an expression removing characters from both ends of a string.
//
// $trim strips whitespace by default, newlines and tabs included. [TrimChars] replaces that with characters of
// your own, and the argument is a set rather than a prefix: TrimChars("ab") removes any run of a's and b's in any
// order from either end, not the literal text "ab". [Ltrim] and [Rtrim] do the same to one end only.
//
// Example:
//
//	expr.Trim(expr.Field("name"))
//	// bson.D{{Key: "$trim", Value: bson.D{{Key: "input", Value: "$name"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/trim/
func Trim(input any, opts ...TrimOption) bson.D {
	return bson.D{{Key: "$trim", Value: trimSpec(input, opts)}}
}

// TrimChars returns a [TrimOption] naming the characters to remove.
//
// The value is a set of characters, so every one of them is stripped from the end being trimmed, in any order and
// any number of times, until a character outside the set turns up. Without this option the default is whitespace.
func TrimChars(chars any) TrimOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "chars", Value: chars})
	}
}

// TrimOption configures the optional character set of [Trim], [Ltrim], and [Rtrim].
type TrimOption func(*bson.D)

// indexOfArgs renders the arguments of $indexOfBytes and $indexOfCP, whose trailing start and end are optional.
func indexOfArgs(str, substring any, startEnd []any) bson.A {
	args := make(bson.A, 0, len(startEnd)+2)
	args = append(args, str, substring)

	return append(args, startEnd...)
}

// regexSpec builds the document the $regex expression operators take.
func regexSpec(input, regex any, opts []RegexOption) bson.D {
	spec := bson.D{{Key: "input", Value: input}, {Key: "regex", Value: regex}}
	for _, opt := range opts {
		opt(&spec)
	}

	return spec
}

// replaceSpec builds the document $replaceOne and $replaceAll take.
func replaceSpec(input, find, replacement any) bson.D {
	return bson.D{
		{Key: "input", Value: input},
		{Key: "find", Value: find},
		{Key: "replacement", Value: replacement},
	}
}

// trimSpec builds the document the $trim operators take.
func trimSpec(input any, opts []TrimOption) bson.D {
	spec := bson.D{{Key: "input", Value: input}}
	for _, opt := range opts {
		opt(&spec)
	}

	return spec
}
