# monq

[![Release](https://img.shields.io/github/v/release/behzadsh/monq)](https://github.com/behzadsh/monq/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/behzadsh/monq.svg)](https://pkg.go.dev/github.com/behzadsh/monq)
[![Go Report Card](https://goreportcard.com/badge/github.com/behzadsh/monq)](https://goreportcard.com/report/github.com/behzadsh/monq)
[![License](https://img.shields.io/github/license/behzadsh/monq)](LICENSE)
![Coverage](https://img.shields.io/badge/Coverage-0.0%25-red)

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

| Category   | Functions                                                |
| ---------- | -------------------------------------------------------- |
| Comparison | `Eq` `Ne` `Gt` `Gte` `Lt` `Lte` `In` `Nin`                |
| Logical    | `And` `Or` `Nor` `Not`                                    |
| Element    | `Exists` `Type`                                           |
| Array      | `All` `ElemMatch` `Size`                                  |
| Evaluation | `Expr` `JSONSchema` `Mod` `Regex` `Text`                  |
| Bitwise    | `BitsAllClear` `BitsAllSet` `BitsAnyClear` `BitsAnySet`   |
| Geospatial | `GeoWithin` `GeoIntersects` `Near` `NearSphere`           |
| Escape     | `Raw`                                                     |

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
| ------------ | ------------------------------------------------------------------------------------------------- |
| Field update | `Set` `SetOnInsert` `Unset` `Inc` `Mul` `Min` `Max` `Rename` `CurrentDate` `CurrentDateTimestamp`  |
| Array update | `Push` `PushEach` `AddToSet` `AddToSetEach` `Pull` `PullAll` `PopFirst` `PopLast`                  |
| Bitwise      | `BitAnd` `BitOr` `BitXor`                                                                          |
| Composition  | `Update`                                                                                            |

The `$each` form of `$push` is its own function, since the `$position`, `$slice`, and `$sort` modifiers only exist there:

```go
monq.PushEach("scores", []any{90, 80}, monq.PushSort(-1), monq.PushSlice(3))
// {"$push": {"scores": {"$each": [90, 80], "$sort": -1, "$slice": 3}}}
```

## Install

```sh
go get github.com/behzadsh/monq
```

## License

[MIT](LICENSE)
