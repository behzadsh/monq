package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// Abs returns an expression yielding the absolute value of a number.
//
// $abs takes a single expression rather than a list, and keeps the numeric type of what it is given. A null or
// missing input yields null.
//
// Example:
//
//	expr.Abs(expr.Field("balance"))
//	// bson.D{{Key: "$abs", Value: "$balance"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/abs/
func Abs(value any) bson.D {
	return bson.D{{Key: "$abs", Value: value}}
}

// Add returns an expression summing its arguments.
//
// $add adds numbers, and also adds milliseconds to a date: with one date among the arguments the result is a date,
// which is how a deadline is computed from a timestamp and a duration. Two dates are an error, since adding one
// point in time to another means nothing. All-numeric arguments follow MongoDB's type promotion, so an int plus a
// double is a double.
//
// Example:
//
//	expr.Add(expr.Field("price"), expr.Field("tax"))
//	// bson.D{{Key: "$add", Value: bson.A{"$price", "$tax"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/add/
func Add(values ...any) bson.D {
	return bson.D{{Key: "$add", Value: operands(values)}}
}

// Ceil returns an expression yielding the smallest integer not less than a number.
//
// $ceil rounds up toward positive infinity, so -3.2 becomes -3. It takes a single expression and returns an
// integer of the same width as the input where it can.
//
// Example:
//
//	expr.Ceil(expr.Field("rating"))
//	// bson.D{{Key: "$ceil", Value: "$rating"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ceil/
func Ceil(value any) bson.D {
	return bson.D{{Key: "$ceil", Value: value}}
}

// Divide returns an expression yielding dividend divided by divisor.
//
// $divide always produces a double, even for two integers that divide evenly. A divisor of 0 fails the whole
// aggregation rather than yielding null, so a divisor that could be zero belongs behind a [Cond].
//
// Example:
//
//	expr.Divide(expr.Field("total"), expr.Field("count"))
//	// bson.D{{Key: "$divide", Value: bson.A{"$total", "$count"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/divide/
func Divide(dividend, divisor any) bson.D {
	return bson.D{{Key: "$divide", Value: bson.A{dividend, divisor}}}
}

// Exp returns an expression raising e to the given power.
//
// $exp is the inverse of [Ln] and takes a single expression, returning a double.
//
// Example:
//
//	expr.Exp(expr.Field("rate"))
//	// bson.D{{Key: "$exp", Value: "$rate"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/exp/
func Exp(exponent any) bson.D {
	return bson.D{{Key: "$exp", Value: exponent}}
}

// Floor returns an expression yielding the largest integer not greater than a number.
//
// $floor rounds down toward negative infinity, so -3.2 becomes -4, which is what separates it from [Trunc] on
// negative numbers.
//
// Example:
//
//	expr.Floor(expr.Field("rating"))
//	// bson.D{{Key: "$floor", Value: "$rating"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/floor/
func Floor(value any) bson.D {
	return bson.D{{Key: "$floor", Value: value}}
}

// Ln returns an expression yielding the natural logarithm of a number.
//
// $ln is [Log] with e as its base and takes a single expression. A value of 0 or less fails the aggregation.
//
// Example:
//
//	expr.Ln(expr.Field("views"))
//	// bson.D{{Key: "$ln", Value: "$views"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/ln/
func Ln(value any) bson.D {
	return bson.D{{Key: "$ln", Value: value}}
}

// Log returns an expression yielding the logarithm of number in the given base.
//
// $log takes both arguments, unlike [Ln] and [Log10], which fix the base. The base has to be greater than 1, and a
// number of 0 or less fails the aggregation.
//
// Example:
//
//	expr.Log(expr.Field("views"), 2)
//	// bson.D{{Key: "$log", Value: bson.A{"$views", 2}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/log/
func Log(number, base any) bson.D {
	return bson.D{{Key: "$log", Value: bson.A{number, base}}}
}

// Log10 returns an expression yielding the base 10 logarithm of a number.
//
// $log10 is [Log] with a base of 10 and takes a single expression. A value of 0 or less fails the aggregation.
//
// Example:
//
//	expr.Log10(expr.Field("views"))
//	// bson.D{{Key: "$log10", Value: "$views"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/log10/
func Log10(value any) bson.D {
	return bson.D{{Key: "$log10", Value: value}}
}

// Mod returns an expression yielding the remainder of dividend divided by divisor.
//
// The result takes the sign of the dividend, so -7 modulo 3 is -1 rather than 2. A divisor of 0 fails the
// aggregation. It computes a remainder, where the query operator of the same name tests one:
// monq.Mod(field, divisor, remainder) is a filter with three arguments, while this takes two and produces a value.
//
// Example:
//
//	expr.Mod(expr.Field("total"), 4)
//	// bson.D{{Key: "$mod", Value: bson.A{"$total", 4}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/mod/
func Mod(dividend, divisor any) bson.D {
	return bson.D{{Key: "$mod", Value: bson.A{dividend, divisor}}}
}

// Multiply returns an expression multiplying its arguments together.
//
// $multiply takes numbers only and follows MongoDB's type promotion, so an int times a double is a double.
//
// Example:
//
//	expr.Multiply(expr.Field("price"), expr.Field("quantity"))
//	// bson.D{{Key: "$multiply", Value: bson.A{"$price", "$quantity"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/multiply/
func Multiply(values ...any) bson.D {
	return bson.D{{Key: "$multiply", Value: operands(values)}}
}

// Pow returns an expression raising base to exponent.
//
// $pow returns an integer when both arguments are integers and the result fits, and a double otherwise. A negative
// exponent on a base of 0 fails the aggregation.
//
// Example:
//
//	expr.Pow(expr.Field("side"), 2)
//	// bson.D{{Key: "$pow", Value: bson.A{"$side", 2}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/pow/
func Pow(base, exponent any) bson.D {
	return bson.D{{Key: "$pow", Value: bson.A{base, exponent}}}
}

// Round returns an expression rounding a number to the given decimal place.
//
// A place of 0 rounds to an integer, a positive place keeps that many decimals, and a negative place rounds to
// tens, hundreds, and so on: -1 turns 1234 into 1230. Halfway values round to even, so 2.5 becomes 2 and 3.5
// becomes 4. MongoDB treats the place as optional and defaults it to 0; monq asks for it either way, since passing
// 0 says the same thing without a variadic parameter that would also accept three numbers.
//
// Example:
//
//	expr.Round(expr.Field("price"), 2)
//	// bson.D{{Key: "$round", Value: bson.A{"$price", 2}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/round/
func Round(value, place any) bson.D {
	return bson.D{{Key: "$round", Value: bson.A{value, place}}}
}

// Sqrt returns an expression yielding the square root of a number.
//
// $sqrt takes a single expression and returns a double. A negative value fails the aggregation.
//
// Example:
//
//	expr.Sqrt(expr.Field("area"))
//	// bson.D{{Key: "$sqrt", Value: "$area"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sqrt/
func Sqrt(value any) bson.D {
	return bson.D{{Key: "$sqrt", Value: value}}
}

// Subtract returns an expression yielding minuend minus subtrahend.
//
// $subtract works on dates as well as numbers, and the order of the arguments decides what comes out: a date minus
// a date is the milliseconds between them, and a date minus a number is an earlier date. A number minus a date is
// an error, which makes this one of the few expression operators where swapping the arguments is not merely a sign
// change.
//
// Example:
//
//	expr.Subtract(expr.Field("total"), expr.Field("discount"))
//	// bson.D{{Key: "$subtract", Value: bson.A{"$total", "$discount"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/subtract/
func Subtract(minuend, subtrahend any) bson.D {
	return bson.D{{Key: "$subtract", Value: bson.A{minuend, subtrahend}}}
}

// Trunc returns an expression cutting a number off at the given decimal place.
//
// $trunc drops the digits past the place rather than rounding them, so it always moves toward zero: -3.7 truncated
// to an integer is -3, where [Floor] gives -4. The place works as it does in [Round], including negative values,
// and is likewise required here though MongoDB defaults it to 0.
//
// Example:
//
//	expr.Trunc(expr.Field("price"), 2)
//	// bson.D{{Key: "$trunc", Value: bson.A{"$price", 2}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/trunc/
func Trunc(value, place any) bson.D {
	return bson.D{{Key: "$trunc", Value: bson.A{value, place}}}
}
