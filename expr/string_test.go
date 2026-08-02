package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestStringExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "Concat",
			got:  expr.Concat(expr.Field("first_name"), " ", expr.Field("last_name")),
			want: bson.D{{Key: "$concat", Value: bson.A{"$first_name", " ", "$last_name"}}},
		},
		{
			name: "IndexOfBytes over the whole string",
			got:  expr.IndexOfBytes(expr.Field("sku"), "-"),
			want: bson.D{{Key: "$indexOfBytes", Value: bson.A{"$sku", "-"}}},
		},
		{
			name: "IndexOfBytes from a start offset",
			got:  expr.IndexOfBytes(expr.Field("sku"), "-", 4),
			want: bson.D{{Key: "$indexOfBytes", Value: bson.A{"$sku", "-", 4}}},
		},
		{
			name: "IndexOfBytes between a start and an end",
			got:  expr.IndexOfBytes(expr.Field("sku"), "-", 4, 12),
			want: bson.D{{Key: "$indexOfBytes", Value: bson.A{"$sku", "-", 4, 12}}},
		},
		{
			name: "IndexOfCP over the whole string",
			got:  expr.IndexOfCP(expr.Field("title"), "e"),
			want: bson.D{{Key: "$indexOfCP", Value: bson.A{"$title", "e"}}},
		},
		{
			name: "IndexOfCP between a start and an end",
			got:  expr.IndexOfCP(expr.Field("title"), "e", 0, 10),
			want: bson.D{{Key: "$indexOfCP", Value: bson.A{"$title", "e", 0, 10}}},
		},
		{
			name: "Trim without a character set",
			got:  expr.Trim(expr.Field("name")),
			want: bson.D{{Key: "$trim", Value: bson.D{{Key: "input", Value: "$name"}}}},
		},
		{
			name: "Trim with a character set",
			got:  expr.Trim(expr.Field("name"), expr.TrimChars(" \t")),
			want: bson.D{{Key: "$trim", Value: bson.D{
				{Key: "input", Value: "$name"},
				{Key: "chars", Value: " \t"},
			}}},
		},
		{
			name: "Ltrim with a character set",
			got:  expr.Ltrim(expr.Field("code"), expr.TrimChars("0")),
			want: bson.D{{Key: "$ltrim", Value: bson.D{
				{Key: "input", Value: "$code"},
				{Key: "chars", Value: "0"},
			}}},
		},
		{
			name: "Rtrim without a character set",
			got:  expr.Rtrim(expr.Field("name")),
			want: bson.D{{Key: "$rtrim", Value: bson.D{{Key: "input", Value: "$name"}}}},
		},
		{
			name: "RegexFind without flags",
			got:  expr.RegexFind(expr.Field("email"), "^[^@]+"),
			want: bson.D{{Key: "$regexFind", Value: bson.D{
				{Key: "input", Value: "$email"},
				{Key: "regex", Value: "^[^@]+"},
			}}},
		},
		{
			name: "RegexFind with flags",
			got:  expr.RegexFind(expr.Field("email"), "^[^@]+", expr.RegexOptions("i")),
			want: bson.D{{Key: "$regexFind", Value: bson.D{
				{Key: "input", Value: "$email"},
				{Key: "regex", Value: "^[^@]+"},
				{Key: "options", Value: "i"},
			}}},
		},
		{
			name: "RegexFindAll",
			got:  expr.RegexFindAll(expr.Field("body"), `#\w+`),
			want: bson.D{{Key: "$regexFindAll", Value: bson.D{
				{Key: "input", Value: "$body"},
				{Key: "regex", Value: `#\w+`},
			}}},
		},
		{
			name: "RegexMatch",
			got:  expr.RegexMatch(expr.Field("email"), `@example\.com$`),
			want: bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: "$email"},
				{Key: "regex", Value: `@example\.com$`},
			}}},
		},
		{
			name: "RegexMatch with a driver regex value",
			got:  expr.RegexMatch(expr.Field("email"), bson.Regex{Pattern: "^a", Options: "i"}),
			want: bson.D{{Key: "$regexMatch", Value: bson.D{
				{Key: "input", Value: "$email"},
				{Key: "regex", Value: bson.Regex{Pattern: "^a", Options: "i"}},
			}}},
		},
		{
			name: "ReplaceOne",
			got:  expr.ReplaceOne(expr.Field("title"), "draft ", ""),
			want: bson.D{{Key: "$replaceOne", Value: bson.D{
				{Key: "input", Value: "$title"},
				{Key: "find", Value: "draft "},
				{Key: "replacement", Value: ""},
			}}},
		},
		{
			name: "ReplaceAll",
			got:  expr.ReplaceAll(expr.Field("phone"), "-", ""),
			want: bson.D{{Key: "$replaceAll", Value: bson.D{
				{Key: "input", Value: "$phone"},
				{Key: "find", Value: "-"},
				{Key: "replacement", Value: ""},
			}}},
		},
		{
			name: "Split",
			got:  expr.Split(expr.Field("tags"), ","),
			want: bson.D{{Key: "$split", Value: bson.A{"$tags", ","}}},
		},
		{
			name: "StrLenBytes takes a bare expression",
			got:  expr.StrLenBytes(expr.Field("name")),
			want: bson.D{{Key: "$strLenBytes", Value: "$name"}},
		},
		{
			name: "StrLenCP takes a bare expression",
			got:  expr.StrLenCP(expr.Field("name")),
			want: bson.D{{Key: "$strLenCP", Value: "$name"}},
		},
		{
			name: "Strcasecmp",
			got:  expr.Strcasecmp(expr.Field("name"), "ada"),
			want: bson.D{{Key: "$strcasecmp", Value: bson.A{"$name", "ada"}}},
		},
		{
			name: "SubstrBytes",
			got:  expr.SubstrBytes(expr.Field("sku"), 0, 3),
			want: bson.D{{Key: "$substrBytes", Value: bson.A{"$sku", 0, 3}}},
		},
		{
			name: "SubstrCP",
			got:  expr.SubstrCP(expr.Field("title"), 0, 10),
			want: bson.D{{Key: "$substrCP", Value: bson.A{"$title", 0, 10}}},
		},
		{
			name: "ToLower takes a bare expression",
			got:  expr.ToLower(expr.Field("email")),
			want: bson.D{{Key: "$toLower", Value: "$email"}},
		},
		{
			name: "ToUpper takes a bare expression",
			got:  expr.ToUpper(expr.Field("country")),
			want: bson.D{{Key: "$toUpper", Value: "$country"}},
		},
		{
			name: "operators nest",
			got:  expr.ToUpper(expr.SubstrCP(expr.Field("country"), 0, 2)),
			want: bson.D{{Key: "$toUpper", Value: bson.D{
				{Key: "$substrCP", Value: bson.A{"$country", 0, 2}},
			}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func ExampleConcat() {
	e := expr.Concat(expr.Field("first_name"), " ", expr.Field("last_name"))

	printExpr(e)
	// Output: {"$concat":["$first_name"," ","$last_name"]}
}

func ExampleIndexOfBytes() {
	e := expr.IndexOfBytes(expr.Field("sku"), "-")

	printExpr(e)
	// Output: {"$indexOfBytes":["$sku","-"]}
}

func ExampleIndexOfCP() {
	e := expr.IndexOfCP(expr.Field("title"), "e", 0, 10)

	printExpr(e)
	// Output: {"$indexOfCP":["$title","e",0,10]}
}

func ExampleTrim() {
	e := expr.Trim(expr.Field("name"))

	printExpr(e)
	// Output: {"$trim":{"input":"$name"}}
}

func ExampleTrimChars() {
	e := expr.Trim(expr.Field("code"), expr.TrimChars("0"))

	printExpr(e)
	// Output: {"$trim":{"input":"$code","chars":"0"}}
}

func ExampleLtrim() {
	e := expr.Ltrim(expr.Field("code"), expr.TrimChars("0"))

	printExpr(e)
	// Output: {"$ltrim":{"input":"$code","chars":"0"}}
}

func ExampleRtrim() {
	e := expr.Rtrim(expr.Field("name"))

	printExpr(e)
	// Output: {"$rtrim":{"input":"$name"}}
}

func ExampleRegexFind() {
	e := expr.RegexFind(expr.Field("email"), "^[^@]+", expr.RegexOptions("i"))

	printExpr(e)
	// Output: {"$regexFind":{"input":"$email","regex":"^[^@]+","options":"i"}}
}

func ExampleRegexFindAll() {
	e := expr.RegexFindAll(expr.Field("body"), `#\w+`)

	printExpr(e)
	// Output: {"$regexFindAll":{"input":"$body","regex":"#\\w+"}}
}

func ExampleRegexMatch() {
	e := expr.RegexMatch(expr.Field("email"), `@example\.com$`)

	printExpr(e)
	// Output: {"$regexMatch":{"input":"$email","regex":"@example\\.com$"}}
}

func ExampleRegexOptions() {
	e := expr.RegexMatch(expr.Field("email"), "^ada", expr.RegexOptions("i"))

	printExpr(e)
	// Output: {"$regexMatch":{"input":"$email","regex":"^ada","options":"i"}}
}

func ExampleReplaceOne() {
	e := expr.ReplaceOne(expr.Field("title"), "draft ", "")

	printExpr(e)
	// Output: {"$replaceOne":{"input":"$title","find":"draft ","replacement":""}}
}

func ExampleReplaceAll() {
	e := expr.ReplaceAll(expr.Field("phone"), "-", "")

	printExpr(e)
	// Output: {"$replaceAll":{"input":"$phone","find":"-","replacement":""}}
}

func ExampleSplit() {
	e := expr.Split(expr.Field("tags"), ",")

	printExpr(e)
	// Output: {"$split":["$tags",","]}
}

func ExampleStrLenBytes() {
	e := expr.StrLenBytes(expr.Field("name"))

	printExpr(e)
	// Output: {"$strLenBytes":"$name"}
}

func ExampleStrLenCP() {
	e := expr.StrLenCP(expr.Field("name"))

	printExpr(e)
	// Output: {"$strLenCP":"$name"}
}

func ExampleStrcasecmp() {
	e := expr.Strcasecmp(expr.Field("name"), "ada")

	printExpr(e)
	// Output: {"$strcasecmp":["$name","ada"]}
}

func ExampleSubstrBytes() {
	e := expr.SubstrBytes(expr.Field("sku"), 0, 3)

	printExpr(e)
	// Output: {"$substrBytes":["$sku",0,3]}
}

func ExampleSubstrCP() {
	e := expr.SubstrCP(expr.Field("title"), 0, 10)

	printExpr(e)
	// Output: {"$substrCP":["$title",0,10]}
}

func ExampleToLower() {
	e := expr.ToLower(expr.Field("email"))

	printExpr(e)
	// Output: {"$toLower":"$email"}
}

func ExampleToUpper() {
	e := expr.ToUpper(expr.Field("country"))

	printExpr(e)
	// Output: {"$toUpper":"$country"}
}
