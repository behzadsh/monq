package expr

import "go.mongodb.org/mongo-driver/v2/bson"

// Acos returns an expression yielding the inverse cosine of a value, in radians.
//
// $acos takes a value between -1 and 1 and returns an angle between 0 and pi. Anything outside that range fails
// the aggregation rather than yielding null. Wrap the result in [RadiansToDegrees] for degrees.
//
// Example:
//
//	expr.Acos(expr.Field("ratio"))
//	// bson.D{{Key: "$acos", Value: "$ratio"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/acos/
func Acos(value any) bson.D {
	return bson.D{{Key: "$acos", Value: value}}
}

// Acosh returns an expression yielding the inverse hyperbolic cosine of a value.
//
// $acosh takes a value of 1 or greater; anything below that fails the aggregation.
//
// Example:
//
//	expr.Acosh(expr.Field("ratio"))
//	// bson.D{{Key: "$acosh", Value: "$ratio"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/acosh/
func Acosh(value any) bson.D {
	return bson.D{{Key: "$acosh", Value: value}}
}

// Asin returns an expression yielding the inverse sine of a value, in radians.
//
// $asin takes a value between -1 and 1 and returns an angle between -pi/2 and pi/2. Anything outside that range
// fails the aggregation.
//
// Example:
//
//	expr.Asin(expr.Field("ratio"))
//	// bson.D{{Key: "$asin", Value: "$ratio"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/asin/
func Asin(value any) bson.D {
	return bson.D{{Key: "$asin", Value: value}}
}

// Asinh returns an expression yielding the inverse hyperbolic sine of a value.
//
// $asinh accepts any number, unlike [Acosh] and [Atanh], which have a restricted domain.
//
// Example:
//
//	expr.Asinh(expr.Field("ratio"))
//	// bson.D{{Key: "$asinh", Value: "$ratio"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/asinh/
func Asinh(value any) bson.D {
	return bson.D{{Key: "$asinh", Value: value}}
}

// Atan returns an expression yielding the inverse tangent of a value, in radians.
//
// $atan accepts any number and returns an angle between -pi/2 and pi/2. [Atan2] is the two-argument form that
// keeps track of which quadrant the point is in.
//
// Example:
//
//	expr.Atan(expr.Field("slope"))
//	// bson.D{{Key: "$atan", Value: "$slope"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/atan/
func Atan(value any) bson.D {
	return bson.D{{Key: "$atan", Value: value}}
}

// Atan2 returns an expression yielding the inverse tangent of y divided by x, in radians.
//
// $atan2 takes the two coordinates separately, which is what lets it return an angle across the full circle,
// between -pi and pi, where [Atan] can only cover half of it. The order is y first, then x, matching the operator
// of the same name in most languages.
//
// Example:
//
//	expr.Atan2(expr.Field("dy"), expr.Field("dx"))
//	// bson.D{{Key: "$atan2", Value: bson.A{"$dy", "$dx"}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/atan2/
func Atan2(y, x any) bson.D {
	return bson.D{{Key: "$atan2", Value: bson.A{y, x}}}
}

// Atanh returns an expression yielding the inverse hyperbolic tangent of a value.
//
// $atanh takes a value between -1 and 1. The bounds themselves give negative and positive infinity, and anything
// outside them fails the aggregation.
//
// Example:
//
//	expr.Atanh(expr.Field("ratio"))
//	// bson.D{{Key: "$atanh", Value: "$ratio"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/atanh/
func Atanh(value any) bson.D {
	return bson.D{{Key: "$atanh", Value: value}}
}

// Cos returns an expression yielding the cosine of an angle given in radians.
//
// $cos measures in radians, not degrees, as every trigonometric operator here does; [DegreesToRadians] converts.
// A null or missing input yields null, and the result is a double unless the input is a decimal.
//
// Example:
//
//	expr.Cos(expr.DegreesToRadians(expr.Field("angle")))
//	// bson.D{{Key: "$cos", Value: bson.D{{Key: "$degreesToRadians", Value: "$angle"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/cos/
func Cos(radians any) bson.D {
	return bson.D{{Key: "$cos", Value: radians}}
}

// Cosh returns an expression yielding the hyperbolic cosine of a value in radians.
//
// $cosh grows quickly, so a large input overflows to infinity rather than failing.
//
// Example:
//
//	expr.Cosh(expr.Field("angle"))
//	// bson.D{{Key: "$cosh", Value: "$angle"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/cosh/
func Cosh(radians any) bson.D {
	return bson.D{{Key: "$cosh", Value: radians}}
}

// DegreesToRadians returns an expression converting degrees to radians.
//
// Every trigonometric operator in this package measures angles in radians, so data stored in degrees goes through
// this first. [RadiansToDegrees] converts the other way.
//
// Example:
//
//	expr.DegreesToRadians(expr.Field("heading"))
//	// bson.D{{Key: "$degreesToRadians", Value: "$heading"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/degreesToRadians/
func DegreesToRadians(degrees any) bson.D {
	return bson.D{{Key: "$degreesToRadians", Value: degrees}}
}

// RadiansToDegrees returns an expression converting radians to degrees.
//
// It is the counterpart of [DegreesToRadians], and what turns the output of [Acos], [Asin], [Atan], or [Atan2]
// back into an angle a reader recognizes.
//
// Example:
//
//	expr.RadiansToDegrees(expr.Atan2(expr.Field("dy"), expr.Field("dx")))
//	// bson.D{{Key: "$radiansToDegrees", Value: bson.D{{Key: "$atan2", Value: bson.A{"$dy", "$dx"}}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/radiansToDegrees/
func RadiansToDegrees(radians any) bson.D {
	return bson.D{{Key: "$radiansToDegrees", Value: radians}}
}

// Sin returns an expression yielding the sine of an angle given in radians.
//
// $sin measures in radians; [DegreesToRadians] converts data stored in degrees. A null or missing input yields
// null.
//
// Example:
//
//	expr.Sin(expr.DegreesToRadians(expr.Field("angle")))
//	// bson.D{{Key: "$sin", Value: bson.D{{Key: "$degreesToRadians", Value: "$angle"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sin/
func Sin(radians any) bson.D {
	return bson.D{{Key: "$sin", Value: radians}}
}

// Sinh returns an expression yielding the hyperbolic sine of a value in radians.
//
// $sinh grows quickly, so a large input overflows to infinity rather than failing.
//
// Example:
//
//	expr.Sinh(expr.Field("angle"))
//	// bson.D{{Key: "$sinh", Value: "$angle"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/sinh/
func Sinh(radians any) bson.D {
	return bson.D{{Key: "$sinh", Value: radians}}
}

// Tan returns an expression yielding the tangent of an angle given in radians.
//
// $tan measures in radians. Near an odd multiple of pi/2 the result grows without bound, which shows up as a very
// large number rather than an error, since no double lands exactly on the asymptote.
//
// Example:
//
//	expr.Tan(expr.DegreesToRadians(expr.Field("angle")))
//	// bson.D{{Key: "$tan", Value: bson.D{{Key: "$degreesToRadians", Value: "$angle"}}}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/tan/
func Tan(radians any) bson.D {
	return bson.D{{Key: "$tan", Value: radians}}
}

// Tanh returns an expression yielding the hyperbolic tangent of a value in radians.
//
// $tanh stays between -1 and 1 whatever it is given, so unlike [Sinh] and [Cosh] it cannot overflow.
//
// Example:
//
//	expr.Tanh(expr.Field("angle"))
//	// bson.D{{Key: "$tanh", Value: "$angle"}}
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/aggregation/tanh/
func Tanh(radians any) bson.D {
	return bson.D{{Key: "$tanh", Value: radians}}
}
