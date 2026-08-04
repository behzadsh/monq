package expr_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq/expr"
)

func TestDateExpressions(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "DateAdd",
			got:  expr.DateAdd(expr.Field("created_at"), "day", 30),
			want: bson.D{{Key: "$dateAdd", Value: bson.D{
				{Key: "startDate", Value: "$created_at"},
				{Key: "unit", Value: "day"},
				{Key: "amount", Value: 30},
			}}},
		},
		{
			name: "DateAdd in a time zone",
			got:  expr.DateAdd(expr.Field("created_at"), "day", 30, expr.Timezone("America/New_York")),
			want: bson.D{{Key: "$dateAdd", Value: bson.D{
				{Key: "startDate", Value: "$created_at"},
				{Key: "unit", Value: "day"},
				{Key: "amount", Value: 30},
				{Key: "timezone", Value: "America/New_York"},
			}}},
		},
		{
			name: "DateSubtract",
			got:  expr.DateSubtract(expr.Field("expires_at"), "day", 7),
			want: bson.D{{Key: "$dateSubtract", Value: bson.D{
				{Key: "startDate", Value: "$expires_at"},
				{Key: "unit", Value: "day"},
				{Key: "amount", Value: 7},
			}}},
		},
		{
			name: "DateDiff",
			got:  expr.DateDiff(expr.Field("created_at"), expr.Field("shipped_at"), "day"),
			want: bson.D{{Key: "$dateDiff", Value: bson.D{
				{Key: "startDate", Value: "$created_at"},
				{Key: "endDate", Value: "$shipped_at"},
				{Key: "unit", Value: "day"},
			}}},
		},
		{
			name: "DateDiff in weeks from Monday",
			got: expr.DateDiff(expr.Field("created_at"), expr.Field("shipped_at"), "week",
				expr.StartOfWeek("monday")),
			want: bson.D{{Key: "$dateDiff", Value: bson.D{
				{Key: "startDate", Value: "$created_at"},
				{Key: "endDate", Value: "$shipped_at"},
				{Key: "unit", Value: "week"},
				{Key: "startOfWeek", Value: "monday"},
			}}},
		},
		{
			name: "DateFromParts passes its parts through",
			got: expr.DateFromParts(bson.D{
				{Key: "year", Value: 2026},
				{Key: "month", Value: 8},
				{Key: "day", Value: 1},
			}),
			want: bson.D{{Key: "$dateFromParts", Value: bson.D{
				{Key: "year", Value: 2026},
				{Key: "month", Value: 8},
				{Key: "day", Value: 1},
			}}},
		},
		{
			name: "DateToParts",
			got:  expr.DateToParts(expr.Field("created_at")),
			want: bson.D{{Key: "$dateToParts", Value: bson.D{{Key: "date", Value: "$created_at"}}}},
		},
		{
			name: "DateToParts with ISO fields",
			got:  expr.DateToParts(expr.Field("created_at"), expr.ISO8601()),
			want: bson.D{{Key: "$dateToParts", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "iso8601", Value: true},
			}}},
		},
		{
			name: "DateFromString with a format and a fallback",
			got: expr.DateFromString(expr.Field("created_on"),
				expr.DateFormat("%Y-%m-%d"), expr.DateOnError(nil)),
			want: bson.D{{Key: "$dateFromString", Value: bson.D{
				{Key: "dateString", Value: "$created_on"},
				{Key: "format", Value: "%Y-%m-%d"},
				{Key: "onError", Value: nil},
			}}},
		},
		{
			name: "DateToString with a format",
			got:  expr.DateToString(expr.Field("created_at"), expr.DateFormat("%Y-%m-%d")),
			want: bson.D{{Key: "$dateToString", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "format", Value: "%Y-%m-%d"},
			}}},
		},
		{
			name: "DateToString with a fallback for null",
			got:  expr.DateToString(expr.Field("created_at"), expr.DateOnNull("unknown")),
			want: bson.D{{Key: "$dateToString", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "onNull", Value: "unknown"},
			}}},
		},
		{
			name: "Year takes the date on its own",
			got:  expr.Year(expr.Field("created_at")),
			want: bson.D{{Key: "$year", Value: "$created_at"}},
		},
		{
			name: "Year in a time zone switches to the document form",
			got:  expr.Year(expr.Field("created_at"), expr.Timezone("America/New_York")),
			want: bson.D{{Key: "$year", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "timezone", Value: "America/New_York"},
			}}},
		},
		{
			name: "Month",
			got:  expr.Month(expr.Field("created_at")),
			want: bson.D{{Key: "$month", Value: "$created_at"}},
		},
		{
			name: "DayOfMonth",
			got:  expr.DayOfMonth(expr.Field("created_at")),
			want: bson.D{{Key: "$dayOfMonth", Value: "$created_at"}},
		},
		{
			name: "DayOfWeek",
			got:  expr.DayOfWeek(expr.Field("created_at")),
			want: bson.D{{Key: "$dayOfWeek", Value: "$created_at"}},
		},
		{
			name: "DayOfYear",
			got:  expr.DayOfYear(expr.Field("created_at")),
			want: bson.D{{Key: "$dayOfYear", Value: "$created_at"}},
		},
		{
			name: "Hour in a time zone",
			got:  expr.Hour(expr.Field("created_at"), expr.Timezone("America/New_York")),
			want: bson.D{{Key: "$hour", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "timezone", Value: "America/New_York"},
			}}},
		},
		{
			name: "Minute",
			got:  expr.Minute(expr.Field("created_at")),
			want: bson.D{{Key: "$minute", Value: "$created_at"}},
		},
		{
			name: "Second",
			got:  expr.Second(expr.Field("created_at")),
			want: bson.D{{Key: "$second", Value: "$created_at"}},
		},
		{
			name: "Millisecond",
			got:  expr.Millisecond(expr.Field("created_at")),
			want: bson.D{{Key: "$millisecond", Value: "$created_at"}},
		},
		{
			name: "Week",
			got:  expr.Week(expr.Field("created_at")),
			want: bson.D{{Key: "$week", Value: "$created_at"}},
		},
		{
			name: "IsoDayOfWeek",
			got:  expr.IsoDayOfWeek(expr.Field("created_at")),
			want: bson.D{{Key: "$isoDayOfWeek", Value: "$created_at"}},
		},
		{
			name: "IsoWeek",
			got:  expr.IsoWeek(expr.Field("created_at")),
			want: bson.D{{Key: "$isoWeek", Value: "$created_at"}},
		},
		{
			name: "IsoWeekYear",
			got:  expr.IsoWeekYear(expr.Field("created_at")),
			want: bson.D{{Key: "$isoWeekYear", Value: "$created_at"}},
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

func ExampleDateAdd() {
	e := expr.DateAdd(expr.Field("created_at"), "day", 30)

	printExpr(e)
	// Output: {"$dateAdd":{"startDate":"$created_at","unit":"day","amount":30}}
}

func ExampleDateSubtract() {
	e := expr.DateSubtract(expr.Field("expires_at"), "day", 7)

	printExpr(e)
	// Output: {"$dateSubtract":{"startDate":"$expires_at","unit":"day","amount":7}}
}

func ExampleDateDiff() {
	e := expr.DateDiff(expr.Field("created_at"), expr.Field("shipped_at"), "day")

	printExpr(e)
	// Output: {"$dateDiff":{"startDate":"$created_at","endDate":"$shipped_at","unit":"day"}}
}

func ExampleDateFromParts() {
	e := expr.DateFromParts(bson.D{{Key: "year", Value: 2026}, {Key: "month", Value: 8}, {Key: "day", Value: 1}})

	printExpr(e)
	// Output: {"$dateFromParts":{"year":2026,"month":8,"day":1}}
}

func ExampleDateToParts() {
	e := expr.DateToParts(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$dateToParts":{"date":"$created_at"}}
}

func ExampleDateFromString() {
	e := expr.DateFromString(expr.Field("created_on"), expr.DateFormat("%Y-%m-%d"))

	printExpr(e)
	// Output: {"$dateFromString":{"dateString":"$created_on","format":"%Y-%m-%d"}}
}

func ExampleDateToString() {
	e := expr.DateToString(expr.Field("created_at"), expr.DateFormat("%Y-%m-%d"))

	printExpr(e)
	// Output: {"$dateToString":{"date":"$created_at","format":"%Y-%m-%d"}}
}

func ExampleYear() {
	e := expr.Year(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$year":"$created_at"}
}

func ExampleTimezone() {
	e := expr.Hour(expr.Field("created_at"), expr.Timezone("America/New_York"))

	printExpr(e)
	// Output: {"$hour":{"date":"$created_at","timezone":"America/New_York"}}
}

func ExampleMonth() {
	e := expr.Month(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$month":"$created_at"}
}

func ExampleDayOfMonth() {
	e := expr.DayOfMonth(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$dayOfMonth":"$created_at"}
}

func ExampleDayOfWeek() {
	e := expr.DayOfWeek(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$dayOfWeek":"$created_at"}
}

func ExampleDayOfYear() {
	e := expr.DayOfYear(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$dayOfYear":"$created_at"}
}

func ExampleHour() {
	e := expr.Hour(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$hour":"$created_at"}
}

func ExampleMinute() {
	e := expr.Minute(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$minute":"$created_at"}
}

func ExampleSecond() {
	e := expr.Second(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$second":"$created_at"}
}

func ExampleMillisecond() {
	e := expr.Millisecond(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$millisecond":"$created_at"}
}

func ExampleWeek() {
	e := expr.Week(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$week":"$created_at"}
}

func ExampleIsoDayOfWeek() {
	e := expr.IsoDayOfWeek(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$isoDayOfWeek":"$created_at"}
}

func ExampleIsoWeek() {
	e := expr.IsoWeek(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$isoWeek":"$created_at"}
}

func ExampleIsoWeekYear() {
	e := expr.IsoWeekYear(expr.Field("created_at"))

	printExpr(e)
	// Output: {"$isoWeekYear":"$created_at"}
}

func ExampleDateFormat() {
	e := expr.DateToString(expr.Field("created_at"), expr.DateFormat("%Y-%m-%d"))

	printExpr(e)
	// Output: {"$dateToString":{"date":"$created_at","format":"%Y-%m-%d"}}
}

func ExampleDateOnNull() {
	e := expr.DateToString(expr.Field("created_at"), expr.DateOnNull("unknown"))

	printExpr(e)
	// Output: {"$dateToString":{"date":"$created_at","onNull":"unknown"}}
}

func ExampleDateOnError() {
	e := expr.DateFromString(expr.Field("created_on"), expr.DateOnError(nil))

	printExpr(e)
	// Output: {"$dateFromString":{"dateString":"$created_on","onError":null}}
}

func ExampleISO8601() {
	e := expr.DateToParts(expr.Field("created_at"), expr.ISO8601())

	printExpr(e)
	// Output: {"$dateToParts":{"date":"$created_at","iso8601":true}}
}

func ExampleStartOfWeek() {
	e := expr.DateDiff(expr.Field("created_at"), expr.Field("shipped_at"), "week", expr.StartOfWeek("monday"))

	printExpr(e)
	// Output: {"$dateDiff":{"startDate":"$created_at","endDate":"$shipped_at","unit":"week","startOfWeek":"monday"}}
}

func TestDateTrunc(t *testing.T) {
	tests := []struct {
		name string
		got  bson.D
		want bson.D
	}{
		{
			name: "truncates to a whole unit",
			got:  expr.DateTrunc(expr.Field("created_at"), "day"),
			want: bson.D{{Key: "$dateTrunc", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "unit", Value: "day"},
			}}},
		},
		{
			name: "bins several units at a time",
			got:  expr.DateTrunc(expr.Field("created_at"), "minute", expr.BinSize(15)),
			want: bson.D{{Key: "$dateTrunc", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "unit", Value: "minute"},
				{Key: "binSize", Value: 15},
			}}},
		},
		{
			name: "weeks take a starting day and a time zone",
			got: expr.DateTrunc(expr.Field("created_at"), "week",
				expr.StartOfWeek("monday"), expr.Timezone("America/New_York")),
			want: bson.D{{Key: "$dateTrunc", Value: bson.D{
				{Key: "date", Value: "$created_at"},
				{Key: "unit", Value: "week"},
				{Key: "startOfWeek", Value: "monday"},
				{Key: "timezone", Value: "America/New_York"},
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

func ExampleDateTrunc() {
	e := expr.DateTrunc(expr.Field("created_at"), "minute", expr.BinSize(15))

	printExpr(e)
	// Output: {"$dateTrunc":{"date":"$created_at","unit":"minute","binSize":15}}
}

func ExampleBinSize() {
	e := expr.DateTrunc(expr.Field("created_at"), "month", expr.BinSize(6))

	printExpr(e)
	// Output: {"$dateTrunc":{"date":"$created_at","unit":"month","binSize":6}}
}
