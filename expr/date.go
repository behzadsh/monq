package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// DateOption configures the optional fields of a date expression.
//
// One type covers the whole category, since the date operators share most of their optional fields. The ones that
// belong to a single operator say so: [StartOfWeek] is for [DateDiff], [ISO8601] for [DateToParts], [DateFormat]
// for [DateToString] and [DateFromString]. Passing an option to an operator that has no such field compiles and
// produces a document the server rejects, in keeping with monq validating nothing itself.
type DateOption func(*bson.D)

// BinSize returns a [DateOption] setting how many units wide each bin of a [DateTrunc] is.
//
// It multiplies the unit, so a binSize of 15 with a unit of "minute" truncates to quarter hours and a binSize of 6
// with "month" gives half years. Without it the bin is one unit wide. Bins are counted from a fixed reference
// point rather than from the data, so the boundaries are the same for every document.
func BinSize(n any) DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "binSize", Value: n})
	}
}

// DateFormat returns a [DateOption] setting the format string of [DateToString] or [DateFromString].
//
// The format is built from specifiers such as %Y for a four-digit year, %m for the month, %d for the day, and %H,
// %M, %S for the time. [DateToString] without one produces an ISO 8601 string, and [DateFromString] without one
// tries to parse an ISO 8601 string.
func DateFormat(format any) DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "format", Value: format})
	}
}

// DateOnError returns a [DateOption] giving the value to use when a date operator fails.
//
// It applies to [DateFromString], where an unparsable string would otherwise fail the whole aggregation. With this
// option the document survives and carries the fallback value instead.
func DateOnError(value any) DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "onError", Value: value})
	}
}

// DateOnNull returns a [DateOption] giving the value to use when the input is null or missing.
//
// It applies to [DateFromString] and [DateToString]. Without it, a null input yields null.
func DateOnNull(value any) DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "onNull", Value: value})
	}
}

// ISO8601 returns a [DateOption] that makes [DateToParts] report ISO 8601 fields.
//
// The returned document then holds isoWeekYear, isoWeek, and isoDayOfWeek in place of year, month, and day. Under
// ISO rules a week belongs to the year holding its Thursday, so the first days of January can report the previous
// year.
func ISO8601() DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "iso8601", Value: true})
	}
}

// StartOfWeek returns a [DateOption] naming the day a week starts on for [DateDiff].
//
// It only matters when the unit is "week". The day is a string such as "monday" or "sun", and the default is
// Sunday.
func StartOfWeek(day any) DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "startOfWeek", Value: day})
	}
}

// Timezone returns a [DateOption] setting the time zone a date operator works in.
//
// The zone is an Olson name such as "America/New_York" or a UTC offset such as "+04:30". It changes which calendar
// day or hour a moment falls in, not merely how it is printed: without it a date stored as UTC reports its UTC
// hour, which is the usual reason a daily grouping comes out shifted.
func Timezone(tz any) DateOption {
	return func(d *bson.D) {
		*d = append(*d, bson.E{Key: "timezone", Value: tz})
	}
}

// DateAdd returns an expression adding an amount of time to a date.
//
// $dateAdd shifts startDate by amount of the given unit, which is one of "year", "quarter", "month", "week",
// "day", "hour", "minute", "second", or "millisecond". A negative amount subtracts, making this and
// [DateSubtract] two ways of writing the same thing. Adding months or years lands on the last day of a shorter
// month rather than spilling over, so January 31 plus one month is February 28 or 29.
//
// Example:
//
//	expr.DateAdd(expr.Field("created_at"), "day", 30)
//	// bson.D{
//	//     {
//	//         Key: "$dateAdd",
//	//         Value: bson.D{
//	//             {Key: "startDate", Value: "$created_at"},
//	//             {Key: "unit", Value: "day"},
//	//             {Key: "amount", Value: 30},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateAdd/
func DateAdd(startDate, unit, amount any, opts ...DateOption) bson.D {
	return bson.D{{Key: "$dateAdd", Value: dateShiftSpec(startDate, unit, amount, opts)}}
}

// DateDiff returns an expression measuring the time between two dates.
//
// $dateDiff counts whole units from startDate to endDate, where the unit is one of "year", "quarter", "month",
// "week", "day", "hour", "minute", "second", or "millisecond". It counts boundaries crossed rather than elapsed
// time, so two moments an hour apart across midnight are one day apart. The result is negative when endDate comes
// first, and [StartOfWeek] matters only for the "week" unit.
//
// Example:
//
//	expr.DateDiff(expr.Field("created_at"), expr.Field("shipped_at"), "day")
//	// bson.D{
//	//     {
//	//         Key: "$dateDiff",
//	//         Value: bson.D{
//	//             {Key: "startDate", Value: "$created_at"},
//	//             {Key: "endDate", Value: "$shipped_at"},
//	//             {Key: "unit", Value: "day"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateDiff/
func DateDiff(startDate, endDate, unit any, opts ...DateOption) bson.D {
	spec := bson.D{
		{Key: "startDate", Value: startDate},
		{Key: "endDate", Value: endDate},
		{Key: "unit", Value: unit},
	}

	return bson.D{{Key: "$dateDiff", Value: applyDateOptions(spec, opts)}}
}

// DateFromParts returns an expression building a date out of its components.
//
// $dateFromParts takes a document of parts rather than positional arguments, and monq passes it through as given
// because option constructors named year or month would collide with the operators that read those parts back
// out. The accepted fields are year, month, day, hour, minute, second, millisecond, and timezone, or for the ISO
// calendar isoWeekYear, isoWeek, isoDayOfWeek and the same time fields. Either year or isoWeekYear has to be
// there; everything else defaults to its lowest value.
//
// Out-of-range parts roll over rather than failing, so a month of 13 means January of the next year, and a day of
// 0 means the last day of the previous month.
//
// Example:
//
//	expr.DateFromParts(bson.D{{Key: "year", Value: 2026}, {Key: "month", Value: 8}, {Key: "day", Value: 1}})
//	// bson.D{
//	//     {
//	//         Key: "$dateFromParts",
//	//         Value: bson.D{
//	//             {Key: "year", Value: 2026}, {Key: "month", Value: 8}, {Key: "day", Value: 1},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateFromParts/
func DateFromParts(parts bson.D) bson.D {
	return bson.D{{Key: "$dateFromParts", Value: parts}}
}

// DateFromString returns an expression parsing a string into a date.
//
// $dateFromString reads an ISO 8601 string unless [DateFormat] gives a format to follow. A string it cannot parse
// fails the whole aggregation, so [DateOnError] is what keeps one bad document from taking the query down, and
// [DateOnNull] covers a missing input.
//
// Example:
//
//	expr.DateFromString(expr.Field("created_on"), expr.DateFormat("%Y-%m-%d"))
//	// bson.D{
//	//     {
//	//         Key: "$dateFromString",
//	//         Value: bson.D{
//	//             {Key: "dateString", Value: "$created_on"},
//	//             {Key: "format", Value: "%Y-%m-%d"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateFromString/
func DateFromString(dateString any, opts ...DateOption) bson.D {
	spec := bson.D{{Key: "dateString", Value: dateString}}

	return bson.D{{Key: "$dateFromString", Value: applyDateOptions(spec, opts)}}
}

// DateSubtract returns an expression taking an amount of time off a date.
//
// $dateSubtract is [DateAdd] with the sign flipped, and takes the same units. A negative amount adds, so the two
// operators can each stand in for the other.
//
// Example:
//
//	expr.DateSubtract(expr.Field("expires_at"), "day", 7)
//	// bson.D{
//	//     {
//	//         Key: "$dateSubtract",
//	//         Value: bson.D{
//	//             {Key: "startDate", Value: "$expires_at"},
//	//             {Key: "unit", Value: "day"},
//	//             {Key: "amount", Value: 7},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateSubtract/
func DateSubtract(startDate, unit, amount any, opts ...DateOption) bson.D {
	return bson.D{{Key: "$dateSubtract", Value: dateShiftSpec(startDate, unit, amount, opts)}}
}

// DateToParts returns an expression breaking a date into a document of its components.
//
// $dateToParts yields year, month, day, hour, minute, second, and millisecond, or the ISO calendar fields when
// [ISO8601] is on. [Timezone] decides which calendar the parts are read in, so it changes the values themselves
// and not just their presentation.
//
// Example:
//
//	expr.DateToParts(expr.Field("created_at"))
//	// bson.D{{Key: "$dateToParts", Value: bson.D{{Key: "date", Value: "$created_at"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateToParts/
func DateToParts(date any, opts ...DateOption) bson.D {
	spec := bson.D{{Key: "date", Value: date}}

	return bson.D{{Key: "$dateToParts", Value: applyDateOptions(spec, opts)}}
}

// DateToString returns an expression rendering a date as a string.
//
// $dateToString produces an ISO 8601 string unless [DateFormat] gives a format, and [Timezone] decides which
// calendar day and hour are rendered. A null date yields null unless [DateOnNull] supplies something else. It is
// the usual way to build a grouping key such as a day or a month.
//
// Example:
//
//	expr.DateToString(expr.Field("created_at"), expr.DateFormat("%Y-%m-%d"))
//	// bson.D{
//	//     {
//	//         Key: "$dateToString",
//	//         Value: bson.D{
//	//             {Key: "date", Value: "$created_at"},
//	//             {Key: "format", Value: "%Y-%m-%d"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateToString/
func DateToString(date any, opts ...DateOption) bson.D {
	spec := bson.D{{Key: "date", Value: date}}

	return bson.D{{Key: "$dateToString", Value: applyDateOptions(spec, opts)}}
}

// DateTrunc returns an expression rounding a date down to the start of a bin.
//
// $dateTrunc is what turns timestamps into grouping keys: with a unit of "day" every moment in a day becomes that
// day's midnight, so a $group on the result buckets by day while keeping a real date rather than the string
// [DateToString] would give. The unit is one of "year", "quarter", "week", "month", "day", "hour", "minute",
// "second", or "millisecond", and [BinSize] widens the bin to several units at once.
//
// [Timezone] decides which day or hour a moment belongs to, and matters here for the same reason it matters to
// [Hour]. With a unit of "week", [StartOfWeek] says which day the week begins on.
//
// Example:
//
//	expr.DateTrunc(expr.Field("created_at"), "minute", expr.BinSize(15))
//	// bson.D{
//	//     {
//	//         Key: "$dateTrunc",
//	//         Value: bson.D{
//	//             {Key: "date", Value: "$created_at"},
//	//             {Key: "unit", Value: "minute"},
//	//             {Key: "binSize", Value: 15},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dateTrunc/
func DateTrunc(date, unit any, opts ...DateOption) bson.D {
	spec := bson.D{{Key: "date", Value: date}, {Key: "unit", Value: unit}}

	return bson.D{{Key: "$dateTrunc", Value: applyDateOptions(spec, opts)}}
}

// DayOfMonth returns an expression yielding the day of the month of a date, from 1 to 31.
//
// With no options the operator takes the date directly; [Timezone] switches it to the document form and decides
// which calendar day the moment falls on, which for a UTC-stored date is otherwise the UTC day.
//
// Example:
//
//	expr.DayOfMonth(expr.Field("created_at"))
//	// bson.D{{Key: "$dayOfMonth", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dayOfMonth/
func DayOfMonth(date any, opts ...DateOption) bson.D {
	return datePart("$dayOfMonth", date, opts)
}

// DayOfWeek returns an expression yielding the day of the week of a date, from 1 for Sunday to 7 for Saturday.
//
// [IsoDayOfWeek] is the one that counts from Monday. [Timezone] decides which day the moment falls on.
//
// Example:
//
//	expr.DayOfWeek(expr.Field("created_at"))
//	// bson.D{{Key: "$dayOfWeek", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dayOfWeek/
func DayOfWeek(date any, opts ...DateOption) bson.D {
	return datePart("$dayOfWeek", date, opts)
}

// DayOfYear returns an expression yielding the day of the year of a date, from 1 to 366.
//
// [Timezone] decides which calendar day the moment falls on.
//
// Example:
//
//	expr.DayOfYear(expr.Field("created_at"))
//	// bson.D{{Key: "$dayOfYear", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/dayOfYear/
func DayOfYear(date any, opts ...DateOption) bson.D {
	return datePart("$dayOfYear", date, opts)
}

// Hour returns an expression yielding the hour of a date, from 0 to 23.
//
// Without [Timezone] the hour is the UTC one, which is the most common way a report comes out shifted.
//
// Example:
//
//	expr.Hour(expr.Field("created_at"), expr.Timezone("America/New_York"))
//	// bson.D{
//	//     {
//	//         Key: "$hour",
//	//         Value: bson.D{
//	//             {Key: "date", Value: "$created_at"},
//	//             {Key: "timezone", Value: "America/New_York"},
//	//         },
//	//     },
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/hour/
func Hour(date any, opts ...DateOption) bson.D {
	return datePart("$hour", date, opts)
}

// IsoDayOfWeek returns an expression yielding the ISO day of the week, from 1 for Monday to 7 for Sunday.
//
// [DayOfWeek] is the one that counts from Sunday.
//
// Example:
//
//	expr.IsoDayOfWeek(expr.Field("created_at"))
//	// bson.D{{Key: "$isoDayOfWeek", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isoDayOfWeek/
func IsoDayOfWeek(date any, opts ...DateOption) bson.D {
	return datePart("$isoDayOfWeek", date, opts)
}

// IsoWeek returns an expression yielding the ISO week number of a date, from 1 to 53.
//
// ISO weeks start on Monday and belong to the year holding their Thursday, so the first days of January can fall
// in week 52 or 53 of the year before. [Week] is the simpler numbering that counts from the first Sunday.
//
// Example:
//
//	expr.IsoWeek(expr.Field("created_at"))
//	// bson.D{{Key: "$isoWeek", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isoWeek/
func IsoWeek(date any, opts ...DateOption) bson.D {
	return datePart("$isoWeek", date, opts)
}

// IsoWeekYear returns an expression yielding the ISO year a date's week belongs to.
//
// It differs from [Year] around New Year, since an ISO week belongs to the year holding its Thursday. Pair it with
// [IsoWeek] rather than mixing ISO and calendar fields.
//
// Example:
//
//	expr.IsoWeekYear(expr.Field("created_at"))
//	// bson.D{{Key: "$isoWeekYear", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/isoWeekYear/
func IsoWeekYear(date any, opts ...DateOption) bson.D {
	return datePart("$isoWeekYear", date, opts)
}

// Millisecond returns an expression yielding the millisecond of a date, from 0 to 999.
//
// Example:
//
//	expr.Millisecond(expr.Field("created_at"))
//	// bson.D{{Key: "$millisecond", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/millisecond/
func Millisecond(date any, opts ...DateOption) bson.D {
	return datePart("$millisecond", date, opts)
}

// Minute returns an expression yielding the minute of a date, from 0 to 59.
//
// A [Timezone] with a half-hour or quarter-hour offset changes this as well as the hour.
//
// Example:
//
//	expr.Minute(expr.Field("created_at"))
//	// bson.D{{Key: "$minute", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/minute/
func Minute(date any, opts ...DateOption) bson.D {
	return datePart("$minute", date, opts)
}

// Month returns an expression yielding the month of a date, from 1 to 12.
//
// [Timezone] decides which calendar month a moment near a boundary falls in.
//
// Example:
//
//	expr.Month(expr.Field("created_at"))
//	// bson.D{{Key: "$month", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/month/
func Month(date any, opts ...DateOption) bson.D {
	return datePart("$month", date, opts)
}

// Second returns an expression yielding the second of a date, from 0 to 59.
//
// Example:
//
//	expr.Second(expr.Field("created_at"))
//	// bson.D{{Key: "$second", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/second/
func Second(date any, opts ...DateOption) bson.D {
	return datePart("$second", date, opts)
}

// Week returns an expression yielding the week of the year of a date, from 0 to 53.
//
// Weeks start on Sunday, and the days before the first Sunday of the year are week 0. [IsoWeek] is the ISO
// numbering, which starts on Monday and has no week 0.
//
// Example:
//
//	expr.Week(expr.Field("created_at"))
//	// bson.D{{Key: "$week", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/week/
func Week(date any, opts ...DateOption) bson.D {
	return datePart("$week", date, opts)
}

// Year returns an expression yielding the year of a date.
//
// [Timezone] decides which calendar year a moment near New Year falls in, and [IsoWeekYear] is the ISO answer to
// the same question, which can differ by one.
//
// Example:
//
//	expr.Year(expr.Field("created_at"))
//	// bson.D{{Key: "$year", Value: "$created_at"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/year/
func Year(date any, opts ...DateOption) bson.D {
	return datePart("$year", date, opts)
}

// applyDateOptions folds opts into an operator's specification document.
func applyDateOptions(spec bson.D, opts []DateOption) bson.D {
	for _, opt := range opts {
		opt(&spec)
	}

	return spec
}

// datePart builds one of the date component operators, which take the date on its own when there is nothing
// optional to say and a document of {date, timezone} otherwise. The two forms mean the same thing to MongoDB.
func datePart(op string, date any, opts []DateOption) bson.D {
	if len(opts) == 0 {
		return bson.D{{Key: op, Value: date}}
	}

	spec := bson.D{{Key: "date", Value: date}}

	return bson.D{{Key: op, Value: applyDateOptions(spec, opts)}}
}

// dateShiftSpec builds the specification $dateAdd and $dateSubtract share.
func dateShiftSpec(startDate, unit, amount any, opts []DateOption) bson.D {
	spec := bson.D{
		{Key: "startDate", Value: startDate},
		{Key: "unit", Value: unit},
		{Key: "amount", Value: amount},
	}

	return applyDateOptions(spec, opts)
}
