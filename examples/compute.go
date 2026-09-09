package main

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"github.com/behzadsh/monq"
	"github.com/behzadsh/monq/expr"
	"github.com/behzadsh/monq/stage"
)

// runCompute covers the expression operators that compute values rather than choose documents: dates, strings,
// conversions, sets, objects, statistics, and windows. They are all from monq/expr, they all go inside a stage,
// and the ones with no fixture to run against use a $documents pipeline, which supplies its own input.
func runCompute(ctx context.Context, db *mongo.Database) error {
	if err := runParts(
		ctx, db.Collection(productsColl),
		computeStrings,
		computeConversions,
		computeTrigonometry,
		computeStatistics,
		computeWindows,
	); err != nil {
		return err
	}

	if err := runParts(ctx, db.Collection(ordersColl), computeDates); err != nil {
		return err
	}

	if err := computeSets(ctx, db); err != nil {
		return err
	}

	return computeObjects(ctx, db)
}

// computeStrings shows the string operators, which count in code points or in bytes depending on the name.
func computeStrings(ctx context.Context, coll *mongo.Collection) error {
	step("string operators: CP counts code points and Bytes counts bytes, which differ outside ASCII")

	stages := stage.Pipeline(
		stage.Match(monq.Eq(ProductPaths.Category, "accessories")),
		stage.Project(
			stage.Field(ProductPaths.ID, 0),
			stage.Field("upper", expr.ToUpper(expr.Field(ProductPaths.SKU))),
			stage.Field("prefix", expr.SubstrCP(expr.Field(ProductPaths.SKU), 0, 3)),
			stage.Field("parts", expr.Split(expr.Field(ProductPaths.SKU), "-")),
			stage.Field("length", expr.StrLenCP(expr.Field(ProductPaths.Name))),
			stage.Field("padded", expr.Concat("<", expr.ToLower(expr.Field(ProductPaths.Name)), ">")),
		),
	)
	if err := showAggregate(ctx, coll, "expr.ToUpper, expr.SubstrCP, expr.Split, expr.StrLenCP, expr.Concat", stages); err != nil {
		return err
	}

	step("trimming takes the characters to strip as an option, and defaults to whitespace without it")

	trimmed := stage.Pipeline(
		stage.Limit(1),
		stage.Project(
			stage.Field(ProductPaths.ID, 0),
			stage.Field("trimmed", expr.Trim("  padded  ")),
			stage.Field("chars", expr.Trim("xxkbdxx", expr.TrimChars("x"))),
			stage.Field("left", expr.Ltrim("..left", expr.TrimChars("."))),
			stage.Field("right", expr.Rtrim("right..", expr.TrimChars("."))),
		),
	)
	if err := showAggregate(ctx, coll, "expr.Trim(input, expr.TrimChars(\".\")), expr.Ltrim, expr.Rtrim", trimmed); err != nil {
		return err
	}

	step("searching and replacing: the regex operators take their flags as an option, not inside the pattern")

	searched := stage.Pipeline(
		stage.Match(monq.Eq(ProductPaths.Category, "peripherals")),
		stage.Project(
			stage.Field(ProductPaths.ID, 0),
			stage.Field(ProductPaths.Name, 1),
			stage.Field("wireless", expr.RegexMatch(expr.Field(ProductPaths.Name), "^wireless", expr.RegexOptions("i"))),
			stage.Field("firstMatch", expr.RegexFind(expr.Field(ProductPaths.Name), "[A-Z]\\w+")),
			stage.Field("dashAt", expr.IndexOfCP(expr.Field(ProductPaths.SKU), "-")),
			stage.Field("renamed", expr.ReplaceOne(expr.Field(ProductPaths.Name), "Wireless", "Cordless")),
			stage.Field("sameName", expr.Strcasecmp(expr.Field(ProductPaths.Name), "wireless mouse")),
		),
	)

	return showAggregate(ctx, coll, "expr.RegexMatch, expr.RegexFind, expr.IndexOfCP, expr.ReplaceOne, expr.Strcasecmp", searched)
}

// computeConversions shows the type operators, where the general Convert takes the fallbacks and the shorthands
// do not.
func computeConversions(ctx context.Context, coll *mongo.Collection) error {
	step("Convert is the general form, and its ConvertOnError keeps a bad value from failing the whole pipeline")

	stages := stage.Pipeline(
		stage.Match(monq.Eq(ProductPaths.Category, "computers")),
		stage.Project(
			stage.Field(ProductPaths.ID, 0),
			stage.Field("priceType", expr.Type(expr.Field(ProductPaths.Price))),
			stage.Field("isNumber", expr.IsNumber(expr.Field(ProductPaths.Price))),
			stage.Field("cents", expr.ToInt(expr.Multiply(expr.Field(ProductPaths.Price), 100))),
			stage.Field("asText", expr.ToString(expr.Field(ProductPaths.Stock))),
			stage.Field("safe", expr.Convert(expr.Field(ProductPaths.SKU), "int", expr.ConvertOnError("not a number"))),
			stage.Field("size", expr.BSONSize("$$ROOT")),
		),
	)

	return showAggregate(ctx, coll, `expr.Type, expr.IsNumber, expr.ToInt, expr.ToString, expr.Convert(..., expr.ConvertOnError(...))`, stages)
}

// computeTrigonometry shows the trigonometric and arithmetic operators, on the one field that is an angle.
func computeTrigonometry(ctx context.Context, coll *mongo.Collection) error {
	step("angles go in as radians, which is what DegreesToRadians is for, and the bitwise operators are here too")

	latitude := expr.ArrayElemAt(expr.Field(ProductPaths.Warehouse.Coordinates.Path), 1)

	stages := stage.Pipeline(
		stage.Group(expr.Field(warehousePath), stage.Accumulator("lat", expr.Max(latitude))),
		stage.Project(
			stage.Field("_id", 0),
			stage.Field("lat", expr.Round(expr.Field("lat"), 2)),
			stage.Field("cos", expr.Round(expr.Cos(expr.DegreesToRadians(expr.Field("lat"))), 4)),
			stage.Field("atan2", expr.Round(expr.RadiansToDegrees(expr.Atan2(expr.Field("lat"), 1)), 2)),
			stage.Field("hypot", expr.Round(expr.Sqrt(expr.Add(expr.Pow(expr.Field("lat"), 2), 1)), 3)),
			stage.Field("bits", expr.BitAnd(expr.ToInt(expr.Abs(expr.Field("lat"))), 12)),
		),
		stage.Sort(monq.Sort(monq.Desc("lat"))),
	)

	return showAggregate(ctx, coll, "expr.Cos(expr.DegreesToRadians(...)), expr.Atan2, expr.Sqrt, expr.Pow, expr.BitAnd", stages)
}

// computeStatistics shows the accumulators that summarize a group beyond a sum, including the ones taking n.
func computeStatistics(ctx context.Context, coll *mongo.Collection) error {
	step("accumulators are expressions too, and the N forms take how many they should keep")

	stages := stage.Pipeline(
		stage.Group(
			nil,
			stage.Accumulator("count", expr.Count()),
			stage.Accumulator("spread", expr.StdDevPop(expr.Field(ProductPaths.Price))),
			stage.Accumulator("sample", expr.StdDevSamp(expr.Field(ProductPaths.Price))),
			stage.Accumulator("median", expr.Median(expr.Field(ProductPaths.Price), "approximate")),
			stage.Accumulator("quartiles", expr.Percentile(expr.Field(ProductPaths.Price), []any{0.25, 0.75}, "approximate")),
			stage.Accumulator("dearest", expr.MaxN(expr.Field(ProductPaths.Price), 3)),
			stage.Accumulator("categories", expr.AddToSet(expr.Field(ProductPaths.Category))),
		),
		// An accumulator is the whole value of its output field, so rounding it happens in a stage of its own.
		stage.Set(
			stage.Field("spread", expr.Round(expr.Field("spread"), 2)),
			stage.Field("sample", expr.Round(expr.Field("sample"), 2)),
		),
	)
	if err := showAggregate(ctx, coll, "expr.Count, expr.StdDevPop, expr.Median, expr.Percentile, expr.MaxN, expr.AddToSet", stages); err != nil {
		return err
	}

	step("Top and Bottom take the order they mean by, so the answer does not depend on a preceding $sort")

	ends := stage.Pipeline(
		stage.Group(
			expr.Field(ProductPaths.Category),
			stage.Accumulator("cheapest", expr.Bottom(monq.Sort(monq.Desc(ProductPaths.Price)), expr.Field(ProductPaths.SKU))),
			stage.Accumulator("twoDearest", expr.TopN(2, monq.Sort(monq.Desc(ProductPaths.Price)), expr.Field(ProductPaths.SKU))),
			stage.Accumulator("firstTwo", expr.FirstN(expr.Field(ProductPaths.SKU), 2)),
			stage.Accumulator("last", expr.Last(expr.Field(ProductPaths.SKU))),
		),
		stage.Sort(monq.Sort(monq.Asc("_id"))),
		stage.Limit(3),
	)

	return showAggregate(ctx, coll, "expr.Bottom, expr.TopN, expr.FirstN, expr.Last", ends)
}

// computeWindows shows the operators that only mean something inside $setWindowFields, where each document sees
// its neighbors.
func computeWindows(ctx context.Context, coll *mongo.Collection) error {
	step("rank operators number the documents, and Shift reaches to the one n places away")

	stages := stage.Pipeline(
		stage.SetWindowFields(
			nil, monq.Sort(monq.Desc(ProductPaths.Price)),
			stage.WindowField("rank", expr.DenseRank()),
			stage.WindowField("row", expr.DocumentNumber()),
			stage.WindowField("nextSKU", expr.Shift(expr.Field(ProductPaths.SKU), 1, expr.ShiftDefault("none"))),
			stage.WindowField("smoothed", expr.ExpMovingAvgN(expr.Field(ProductPaths.Price), 3)),
		),
		stage.Project(
			stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1), stage.Field(ProductPaths.Price, 1),
			stage.Field("rank", 1), stage.Field("row", 1), stage.Field("nextSKU", 1),
			stage.Field("smoothed", expr.Round(expr.Field("smoothed"), 2)),
		),
		stage.Limit(5),
	)
	if err := showAggregate(ctx, coll, "expr.DenseRank, expr.DocumentNumber, expr.Shift(..., expr.ShiftDefault(...)), expr.ExpMovingAvgN", stages); err != nil {
		return err
	}

	step("the calculus ones take the unit their rate is per, and covariance takes the two series to compare")

	rates := stage.Pipeline(
		stage.SetWindowFields(
			nil, monq.Sort(monq.Asc(ProductPaths.CreatedAt)),
			stage.WindowField(
				"perDay", expr.Derivative(expr.Field(ProductPaths.Stock), expr.TimeUnit("day")),
				stage.WindowRange(-1, stage.WindowCurrent),
				stage.WindowUnit("month"),
			),
			stage.WindowField(
				"priceVsStock", expr.CovariancePop(expr.Field(ProductPaths.Price), expr.Field(ProductPaths.Stock)),
				stage.WindowDocuments(stage.WindowUnbounded, stage.WindowCurrent),
			),
		),
		stage.Project(
			stage.Field(ProductPaths.ID, 0), stage.Field(ProductPaths.SKU, 1),
			stage.Field("perDay", expr.Round(expr.IfNull(expr.Field("perDay"), 0), 3)),
			stage.Field("priceVsStock", expr.Round(expr.Field("priceVsStock"), 1)),
		),
		stage.Limit(5),
	)

	return showAggregate(ctx, coll, "expr.Derivative(..., expr.TimeUnit(\"day\")), expr.CovariancePop", rates)
}

// computeDates shows the date operators, which all take the same optional timezone and read the same field.
func computeDates(ctx context.Context, coll *mongo.Collection) error {
	step("date parts come out one operator each, and every one of them takes a timezone")

	placed := expr.Field(OrderPaths.PlacedAt)

	stages := stage.Pipeline(
		stage.Sort(monq.Sort(monq.Asc(OrderPaths.PlacedAt))),
		stage.Project(
			stage.Field(OrderPaths.ID, 0),
			stage.Field("year", expr.Year(placed)),
			stage.Field("month", expr.Month(placed)),
			stage.Field("day", expr.DayOfMonth(placed)),
			stage.Field("weekday", expr.DayOfWeek(placed)),
			stage.Field("isoWeek", expr.IsoWeek(placed)),
			stage.Field("hourInTokyo", expr.Hour(placed, expr.Timezone("Asia/Tokyo"))),
		),
		stage.Limit(3),
	)
	if err := showAggregate(ctx, coll, "expr.Year, expr.Month, expr.DayOfMonth, expr.IsoWeek, expr.Hour(..., expr.Timezone(...))", stages); err != nil {
		return err
	}

	step("date arithmetic works in units, and DateToString formats with the MongoDB format specifiers")

	arithmetic := stage.Pipeline(
		stage.Sort(monq.Sort(monq.Asc(OrderPaths.PlacedAt))),
		stage.Project(
			stage.Field(OrderPaths.ID, 0),
			stage.Field("formatted", expr.DateToString(placed, expr.DateFormat("%Y-%m-%d %H:%M"))),
			stage.Field("dueBy", expr.DateAdd(placed, "day", 14)),
			stage.Field("monthStart", expr.DateTrunc(placed, "month")),
			stage.Field("daysAgo", expr.DateDiff(placed, seedTime, "day")),
			stage.Field("weekBefore", expr.DateSubtract(placed, "week", 1)),
		),
		stage.Limit(3),
	)
	if err := showAggregate(ctx, coll, "expr.DateToString(..., expr.DateFormat(...)), expr.DateAdd, expr.DateTrunc, expr.DateDiff", arithmetic); err != nil {
		return err
	}

	step("dates also come apart into their parts and back together, and parse from a string")

	parts := stage.Pipeline(
		stage.Limit(1),
		stage.Project(
			stage.Field(OrderPaths.ID, 0),
			stage.Field("parts", expr.DateToParts(placed)),
			stage.Field("built", expr.DateFromParts(bson.D{{Key: "year", Value: 2026}, {Key: "month", Value: 3}, {Key: "day", Value: 2}})),
			stage.Field("parsed", expr.DateFromString("2026-03-02T09:00:00Z")),
			stage.Field("bad", expr.DateFromString("not a date", expr.DateOnError(nil))),
		),
	)

	return showAggregate(ctx, coll, "expr.DateToParts, expr.DateFromParts, expr.DateFromString(..., expr.DateOnError(nil))", parts)
}

// computeSets shows the set operators, which treat arrays as sets and so ignore order and duplicates. They run
// over literal documents, since the point is the operator rather than the data.
func computeSets(ctx context.Context, db *mongo.Database) error {
	step("set operators ignore order and duplicates, while the array operators do not")

	stages := stage.Pipeline(
		stage.Documents(
			bson.D{
				{Key: "a", Value: bson.A{"wireless", "rgb", "usb-c"}},
				{Key: "b", Value: bson.A{"usb-c", "wireless"}},
			},
		),
		stage.Project(
			stage.Field("union", expr.SetUnion(expr.Field("a"), expr.Field("b"))),
			stage.Field("shared", expr.SetIntersection(expr.Field("a"), expr.Field("b"))),
			stage.Field("onlyInA", expr.SetDifference(expr.Field("a"), expr.Field("b"))),
			stage.Field("bIsSubset", expr.SetIsSubset(expr.Field("b"), expr.Field("a"))),
			stage.Field("same", expr.SetEquals(expr.Field("a"), expr.Field("b"))),
			stage.Field("allTruthy", expr.AllElementsTrue(expr.Field("a"))),
			stage.Field("anyTruthy", expr.AnyElementTrue(expr.Field("b"))),
		),
	)
	if err := showDBAggregate(ctx, db, "expr.SetUnion, expr.SetIntersection, expr.SetDifference, expr.SetIsSubset, expr.SetEquals", stages); err != nil {
		return err
	}

	step("the array operators keep order, and Range and Zip build arrays rather than reading them")

	arrays := stage.Pipeline(
		stage.Documents(bson.D{{Key: "scores", Value: bson.A{4, 1, 3}}, {Key: "names", Value: bson.A{"ada", "grace", "linus"}}}),
		stage.Project(
			stage.Field("sorted", expr.SortArray(expr.Field("scores"), 1)),
			stage.Field("reversed", expr.ReverseArray(expr.Field("names"))),
			stage.Field("joined", expr.ConcatArrays(expr.Field("names"), bson.A{"hopper"})),
			stage.Field("hasGrace", expr.In("grace", expr.Field("names"))),
			stage.Field("whereIsLinus", expr.IndexOfArray(expr.Field("names"), "linus")),
			stage.Field("firstTwo", expr.Slice(expr.Field("names"), 2)),
			stage.Field("counted", expr.Range(0, expr.Size(expr.Field("names")))),
			stage.Field("paired", expr.Zip([]any{expr.Field("names"), expr.Field("scores")})),
		),
	)

	return showDBAggregate(ctx, db, "expr.SortArray, expr.ReverseArray, expr.In, expr.IndexOfArray, expr.Range, expr.Zip", arrays)
}

// computeObjects shows the operators that treat a document as data: turning it into pairs and back, and reading
// or writing a field whose name is itself computed.
func computeObjects(ctx context.Context, db *mongo.Database) error {
	step("$objectToArray and $arrayToObject are inverses, which is how a document is iterated over at all")

	stages := stage.Pipeline(
		stage.Documents(
			bson.D{
				{Key: "specs", Value: bson.D{{Key: "weight", Value: 1.2}, {Key: "color", Value: "black"}}},
				{Key: "extra", Value: bson.D{{Key: "warranty", Value: "2 years"}}},
			},
		),
		stage.Project(
			stage.Field("pairs", expr.ObjectToArray(expr.Field("specs"))),
			stage.Field("rebuilt", expr.ArrayToObject(expr.ObjectToArray(expr.Field("specs")))),
			stage.Field("merged", expr.MergeObjects(expr.Field("specs"), expr.Field("extra"))),
			stage.Field("color", expr.GetField("color", expr.Field("specs"))),
			stage.Field("withPrice", expr.SetField("price", expr.Field("specs"), 19.99)),
			stage.Field("without", expr.UnsetField("weight", expr.Field("specs"))),
		),
	)
	if err := showDBAggregate(ctx, db, "expr.ObjectToArray, expr.ArrayToObject, expr.MergeObjects, expr.GetField, expr.SetField", stages); err != nil {
		return err
	}

	step("Let names a subexpression, Literal protects a value that would otherwise be read as a reference")

	misc := stage.Pipeline(
		stage.Documents(bson.D{{Key: "price", Value: 40.0}, {Key: "qty", Value: 3}}),
		stage.Project(
			stage.Field(
				"discounted", expr.Let(
					bson.D{{Key: "total", Value: expr.Multiply(expr.Field("price"), expr.Field("qty"))}},
					expr.Round(expr.Multiply("$$total", 0.9), 2),
				),
			),
			stage.Field("marker", expr.Literal("$price")),
			stage.Field("coin", expr.Cond(expr.Gte(expr.Rand(), 0.5), "heads", "tails")),
			// 1772449200000 is milliseconds since the epoch, which is what $toDate reads a number as.
			stage.Field("epoch", expr.ToDate(int64(1_772_449_200_000))),
		),
	)

	return showDBAggregate(ctx, db, `expr.Let(vars, in), expr.Literal("$price"), expr.Rand, expr.ToDate`, misc)
}
