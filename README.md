# monq

[![CI](https://github.com/behzadsh/monq/actions/workflows/ci.yml/badge.svg)](https://github.com/behzadsh/monq/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/behzadsh/monq)](https://github.com/behzadsh/monq/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/behzadsh/monq.svg)](https://pkg.go.dev/github.com/behzadsh/monq)
[![License](https://img.shields.io/github/license/behzadsh/monq)](LICENSE)
[![codecov](https://codecov.io/gh/behzadsh/monq/branch/main/graph/badge.svg)](https://codecov.io/gh/behzadsh/monq)

A composable query-building library for the [official MongoDB Go driver](https://pkg.go.dev/go.mongodb.org/mongo-driver/v2/mongo).

## Why

Hand-written MongoDB queries with the Go driver mean nesting `bson.D`, `bson.A`, and `bson.E` values, and remembering operator strings like `$eq`, `$gte`,
or `$elemMatch`, usually with the MongoDB docs open in another tab. `monq` replaces that with plain functions named after the operators they build, so the
function signature tells you what it does and your editor's autocomplete surfaces the operators you have available.

`monq` is not an ODM. There are no models, no sessions, no query execution, it only builds `bson.D` values and hands them back. Every function's output
plugs directly into `Find`, `Aggregate`, `UpdateOne`, and the rest of the driver's API with no adapter layer in between.

## How it works

Each function takes a field path and a value (or, for logical operators, other `monq` expressions) and returns a `bson.D`:

```go
monq.Eq("status", "active")
// bson.D{{Key: "status", Value: bson.D{{Key: "$eq", Value: "active"}}}}
```

Filters compose by nesting function calls, no builder object, no method chaining:

```go
filter := monq.And(
monq.Eq("status", "active"),
monq.Or(
monq.Gte("stats.followers", 10000),
monq.Exists("verified_at", true),
),
)

cursor, err := collection.Find(ctx, filter)
```

For anything `monq` doesn't have a function for yet, [`Raw`](raw.go) accepts a hand-written `bson.D` anywhere a monq expression is expected, so it composes
with the rest of a query instead of forcing an all-or-nothing rewrite:

```go
monq.And(
monq.Eq("status", "active"),
monq.Raw(bson.D{{Key: "$where", Value: "this.credits == this.debits"}}),
)
```

Array conditions read the same way. `ElemMatch` takes the criteria a single array element has to satisfy:

```go
filter := monq.ElemMatch("items",
monq.Eq("sku", "abc"),
monq.Gte("qty", 2),
)
// {"items": {"$elemMatch": {"sku": {"$eq": "abc"}, "qty": {"$gte": 2}}}}
```

## Operators

Query operators available today:

| Category   | Functions                                               |
|------------|---------------------------------------------------------|
| Comparison | `Eq` `Ne` `Gt` `Gte` `Lt` `Lte` `In` `Nin`              |
| Logical    | `And` `Or` `Nor` `Not`                                  |
| Element    | `Exists` `Type`                                         |
| Array      | `All` `ElemMatch` `Size`                                |
| Evaluation | `Expr` `JSONSchema` `Mod` `Regex` `Text` `SampleRate`   |
| Bitwise    | `BitsAllClear` `BitsAllSet` `BitsAnyClear` `BitsAnySet` |
| Geospatial | `GeoWithin` `GeoIntersects` `Near` `NearSphere`         |
| Escape     | `Raw`                                                   |

Operators with optional parts take them as variadic options named after the MongoDB field they set:

```go
monq.Text("coffee shop", monq.Language("en"), monq.CaseSensitive())
// {"$text": {"$search": "coffee shop", "$language": "en", "$caseSensitive": true}}
```

Geospatial queries come with geometry constructors, so there is no hand-written GeoJSON. `Point`, `Polygon`, and `GeoJSON` build the shape, `Geometry`
hands it to an operator, and the legacy `Box`, `Center`, and `CenterSphere` shapes are there for 2d data. Positions are `[longitude, latitude]`:

```go
monq.Near("loc", monq.Geometry(monq.Point(-73.97, 40.77)), monq.MaxDistance(1000))
// {"loc": {"$near": {"$geometry": {"type": "Point", "coordinates": [-73.97, 40.77]}, "$maxDistance": 1000.0}}}
```

## Projections

A projection decides which fields come back. Entries merge the same way sort entries do, in the order given:

```go
projection := monq.Projection(
monq.Include("email", "items.sku"),
monq.Exclude("_id"),
monq.Slice("comments", -5),
)

cursor, err := collection.Find(ctx, filter, options.Find().SetProjection(projection))
// {"email": 1, "items.sku": 1, "_id": 0, "comments": {"$slice": -5}}
```

Two operators you already have double as projection entries, because MongoDB spells them the same way: `ElemMatch` returns the first matching element of an
array, and `TextScore` adds a `$text` relevance score to the result. `ArrayPath.Positional()` gives the `$` positional form, so
`monq.Include(items.Positional())` returns just the element the query matched.

| Category   | Functions                                                   |
|------------|-------------------------------------------------------------|
| Projection | `Projection` `Include` `Exclude` `Slice` `SliceFrom` `Meta` |

## Updates

Update operators work the same way, one field each, and a single one is already a valid update document:

```go
collection.UpdateOne(ctx, filter, monq.Set("status", "active"))
// {"$set": {"status": "active"}}
```

Several of them go through `Update`, which merges operators sharing a key. It is to update documents what `And` is to filters, and it exists because two
`$set` documents concatenated by hand end up as a duplicate key that MongoDB does not merge:

```go
update := monq.Update(
monq.Set("status", "active"),
monq.Inc("logins", 1),
monq.Set("name", "ada"),
)
// {"$set": {"status": "active", "name": "ada"}, "$inc": {"logins": 1}}
```

| Category     | Functions                                                                                         |
|--------------|---------------------------------------------------------------------------------------------------|
| Field update | `Set` `SetOnInsert` `Unset` `Inc` `Mul` `Min` `Max` `Rename` `CurrentDate` `CurrentDateTimestamp` |
| Array update | `Push` `PushEach` `AddToSet` `AddToSetEach` `Pull` `PullAll` `PopFirst` `PopLast`                 |
| Bitwise      | `BitAnd` `BitOr` `BitXor`                                                                         |
| Composition  | `Update`                                                                                          |

The `$each` form of `$push` is its own function, since the `$position`, `$slice`, and `$sort` modifiers only exist there:

```go
monq.PushEach("scores", []any{90, 80}, monq.PushSort(-1), monq.PushSlice(3))
// {"$push": {"scores": {"$each": [90, 80], "$sort": -1, "$slice": 3}}}
```

## Aggregation stages

Pipeline stages live in `monq/stage`, one function per stage. The package split is what keeps names honest: `stage.Set` is the `$set` stage while
`monq.Set` is the `$set` update operator, and the qualifier says which one is meant.

```go
pipeline := stage.Pipeline(
stage.Match(monq.Eq("status", "active")),
stage.Sort(bson.D{{Key: "created_at", Value: -1}}),
stage.Limit(20),
)

cursor, err := collection.Aggregate(ctx, pipeline)
```

`Pipeline` returns a `[]bson.D`, which is what the driver's `mongo.Pipeline` is defined as, so it goes into `Aggregate` as is. A plain slice literal works
too.

Grouping stages take their output fields as `Accumulator` pieces, merged into one document by the stage:

```go
stage.Group("$category",
stage.Accumulator("total", bson.D{{Key: "$sum", Value: "$amount"}}),
)
// {"$group": {"_id": "$category", "total": {"$sum": "$amount"}}}
```

| Category   | Functions                                                                    |
|------------|------------------------------------------------------------------------------|
| Filtering  | `Match` `Limit` `Skip` `Sample` `Count` `Sort`                               |
| Grouping   | `Group` `Bucket` `BucketAuto` `SortByCount` `Facet` `Unwind`                 |
| Joining    | `Lookup` `LookupPipeline` `GraphLookup` `UnionWith`                          |
| Reshaping  | `Project` `AddFields` `Set` `Unset` `ReplaceRoot` `ReplaceWith`              |
| Output     | `Out` `Merge` `Documents`                                                    |
| Windows    | `SetWindowFields` `WindowField` `WindowDocuments` `WindowRange` `WindowUnit` |
| Series     | `Densify` `DensifyRange` `Fill` `FillValue` `FillMethod` `Redact`            |
| Geospatial | `GeoNear`                                                                    |
| Building   | `Field` `Accumulator` `FacetPipeline` `Namespace`                            |
| Assembly   | `Pipeline`                                                                   |

`GeoNear` takes the same geometry constructors the query operators do, which is why they return bare GeoJSON:

```go
stage.GeoNear(monq.Point(-73.97, 40.77), "distance", stage.MaxDistance(1000), stage.Spherical())
// {"$geoNear": {"near": {"type": "Point", "coordinates": [-73.97, 40.77]}, "distanceField": "distance", ...}}
```

## Aggregation expressions

`monq/expr` builds the expressions stages compute with. Expression operators compare values rather than naming a field, so a field goes in as a reference:

```go
stage.Project(
stage.Field("name", 1),
stage.Field("grade", expr.Switch(
expr.Branch(expr.Gte(expr.Field("score"), 90), "A"),
expr.Branch(expr.Gte(expr.Field("score"), 80), "B"),
expr.DefaultCase("F"),
)),
)
```

`expr.Field("score")` is the string `"$score"`, which is how the aggregation framework tells a field from a constant. The rule runs both ways: any string
starting with `$` is read as a reference, so a literal one goes through `expr.Literal`. That is also why `expr.Eq(a, b)` and `monq.Eq(field, value)` are
different functions rather than one name; `expr.Eq("status", "active")` compares two constants and is false everywhere.

Accumulators are the same functions, since MongoDB spells them the same way. `Sum` reads its argument count: one argument totals a field across a group,
several add them up inside each document.

```go
stage.Group("$category",
stage.Accumulator("total", expr.Sum(expr.Field("amount"))),
stage.Accumulator("best", expr.Top(bson.D{{Key: "score", Value: -1}}, expr.Field("name"))),
)
```

| Category     | Functions                                                                                                                                |
|--------------|------------------------------------------------------------------------------------------------------------------------------------------|
| References   | `Field` `Literal`                                                                                                                        |
| Comparison   | `Cmp` `Eq` `Ne` `Gt` `Gte` `Lt` `Lte`                                                                                                    |
| Boolean      | `And` `Or` `Not`                                                                                                                         |
| Conditional  | `Cond` `IfNull` `Switch` `Branch` `DefaultCase`                                                                                          |
| Arithmetic   | `Abs` `Add` `Ceil` `Divide` `Exp` `Floor` `Ln` `Log` `Log10` `Mod` `Multiply` `Pow` `Round` `Sqrt` `Subtract` `Trunc`                    |
| String       | `Concat` `Split` `SubstrBytes` `SubstrCP` `StrLenBytes` `StrLenCP` `Strcasecmp` `ToLower` `ToUpper`                                      |
| Trimming     | `Trim` `Ltrim` `Rtrim` `TrimChars`                                                                                                       |
| Searching    | `IndexOfBytes` `IndexOfCP` `RegexFind` `RegexFindAll` `RegexMatch` `ReplaceOne` `ReplaceAll`                                             |
| Array        | `ArrayElemAt` `ConcatArrays` `First` `Last` `FirstN` `LastN` `MaxN` `MinN` `In` `IndexOfArray` `IsArray` `Size`                          |
| Array shape  | `Filter` `Map` `Reduce` `Range` `ReverseArray` `Slice` `SliceFrom` `SortArray` `Zip`                                                     |
| Object       | `ArrayToObject` `ObjectToArray` `MergeObjects` `GetField` `SetField` `UnsetField`                                                        |
| Set          | `AllElementsTrue` `AnyElementTrue` `SetDifference` `SetEquals` `SetIntersection` `SetIsSubset` `SetUnion`                                |
| Date         | `DateAdd` `DateSubtract` `DateDiff` `DateTrunc` `DateFromParts` `DateToParts` `DateFromString` `DateToString`                            |
| Date parts   | `Year` `Month` `DayOfMonth` `DayOfWeek` `DayOfYear` `Hour` `Minute` `Second` `Millisecond` `Week` `IsoDayOfWeek` `IsoWeek` `IsoWeekYear` |
| Conversion   | `Convert` `IsNumber` `Type` `ToBool` `ToDate` `ToDecimal` `ToDouble` `ToInt` `ToLong` `ToObjectID` `ToString`                            |
| Accumulator  | `Sum` `Avg` `Max` `Min` `Push` `AddToSet` `Count` `StdDevPop` `StdDevSamp` `Top` `TopN` `Bottom` `BottomN` `Median` `Percentile`         |
| Trigonometry | `Sin` `Cos` `Tan` `Asin` `Acos` `Atan` `Atan2` `Sinh` `Cosh` `Tanh` `Asinh` `Acosh` `Atanh` `DegreesToRadians` `RadiansToDegrees`        |
| Bitwise      | `BitAnd` `BitOr` `BitXor` `BitNot`                                                                                                       |
| Misc         | `Let` `Rand` `BinarySize` `BSONSize` `TsSecond` `TsIncrement`                                                                            |
| Window rank  | `Rank` `DenseRank` `DocumentNumber` `Shift`                                                                                              |
| Window fill  | `Locf` `LinearFill`                                                                                                                      |
| Window calc  | `Derivative` `Integral` `ExpMovingAvgN` `ExpMovingAvgAlpha` `CovariancePop` `CovarianceSamp`                                             |

## Sorting and indexes

Sort documents are built the same way as everything else, and order matters, so the entries stay in the order given:

```go
sort := monq.Sort(monq.Desc("created_at"), monq.Asc("_id"))
// {"created_at": -1, "_id": 1}

cursor, err := collection.Find(ctx, filter, options.Find().SetSort(sort))
```

Index models live in `monq/index`, where keys and options are one argument list:

```go
model := index.New(
index.Asc("email"),
index.Unique(),
index.PartialFilter(monq.Exists("deleted_at", false)),
)

_, err := collection.Indexes().CreateOne(ctx, model)
```

That package is separate on purpose. `mongo.IndexModel` comes from the driver's `mongo` package, which carries several third-party dependencies; keeping it
out of the root means a program that only builds queries never compiles any of them.

| Package      | Functions                                                                                                        |
|--------------|------------------------------------------------------------------------------------------------------------------|
| `monq`       | `Sort` `Asc` `Desc` `TextScore`                                                                                  |
| `monq/index` | `New` `Asc` `Desc` `Text` `Hashed` `Geo2D` `Geo2DSphere` `Unique` `Sparse` `TTL` `PartialFilter` `Name` `Hidden` |

## Generated field paths

Field paths are strings, so a renamed field turns into a query that silently matches nothing. `monqgen` reads the struct a collection stores and writes its
paths out as typed constants:

```go
//go:generate go run github.com/behzadsh/monq/cmd/monqgen -type User

type User struct {
ID    bson.ObjectID `bson:"_id"`
Email string        `bson:"email"`
Items []Item        `bson:"items"`
}
```

`go generate ./...` writes `user_paths.go` next to it, and the paths go straight into any operator that takes one:

```go
monq.Eq(UserPaths.Email, "a@b.c") // {"email": {"$eq": "a@b.c"}}
monq.Eq(UserPaths.Items.SKU, "abc") // {"items.sku": {"$eq": "abc"}}
monq.Size(UserPaths.Items.Path, 3) // {"items": {"$size": 3}}
```

Arrays carry the four ways MongoDB names a position, so the punctuation never has to be remembered:

```go
UserPaths.Items.At(3).Quantity // "items.3.qty"
UserPaths.Items.Positional().Price // "items.$.price"
UserPaths.Items.All().SKU // "items.$[].sku"
UserPaths.Items.Filtered("cheap").SKU // "items.$[cheap].sku"
```

Paths follow the driver's own tag rules rather than `encoding/json`'s: a key defaults to the field name lowercased whole, `bson:"-"` drops a field, and an
embedded struct nests under its own name unless it is tagged `,inline`. Types that encode themselves, such as `time.Time` and `bson.ObjectID`, are leaves.
The same positional helpers are available by hand through `monq.ArrayPath` when there is no generated struct.

The command reads the package directory given as its argument, defaulting to the current one, which is why a `go:generate` line needs nothing but the type:

| Flag     | Meaning                                                                             |
|----------|-------------------------------------------------------------------------------------|
| `-type`  | the struct to read paths from; required                                             |
| `-out`   | the file to write, defaulting to the type name lowercased with `_paths.go` appended |
| `-print` | write the path tree to standard output and generate nothing                         |

`-print` answers what a struct yields without touching the disk, which is the quickest way to check a tag change:

```sh
$ go run github.com/behzadsh/monq/cmd/monqgen -type User -print ./internal/model
_id
email
items
items.sku
```

Two types in one package means two `go:generate` lines, and the default output name keeps them in separate files. monqgen refuses to overwrite a file that
does not carry its generated header, so a mistyped `-out` cannot eat source.

The generated file is regular Go source with no runtime magic: [`cmd/monqgen/internal/example`](cmd/monqgen/internal/example) holds a struct, the file
monqgen wrote from it, and tests using those paths against every part of the API. Codegen is entirely optional, and paths written by hand work the same
way.

## Examples

[`examples/`](examples) is a runnable program covering every area of the API against a real MongoDB. Each step prints the `bson.D` that monq built and then
the documents the server returned for it, which is the part a godoc example cannot show:

```sh
docker run --rm -p 27017:27017 mongo:8
go run ./examples              # every demo, in order
go run ./examples find update  # or just the ones named
```

The demos are `find`, `project`, `update`, `aggregate`, `compute`, `geo`, `index`, and `paths`, one file each. The database is reseeded before every demo,
so they are independent and safe to run in any order. See [`examples/README.md`](examples/README.md) for what each one covers.

## Install

```sh
go get github.com/behzadsh/monq
```

Go 1.25 or newer, matching the releases the Go team still supports.

## License

[MIT](LICENSE)
