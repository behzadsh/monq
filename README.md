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

| Category   | Functions                                        |
| ---------- | ------------------------------------------------ |
| Comparison | `Eq` `Ne` `Gt` `Gte` `Lt` `Lte` `In` `Nin`        |
| Logical    | `And` `Or` `Nor` `Not`                           |
| Element    | `Exists` `Type`                                  |
| Array      | `All` `ElemMatch` `Size`                         |
| Evaluation | `Regex`                                          |
| Escape     | `Raw`                                            |

## Install

```sh
go get github.com/behzadsh/monq
```

## License

[MIT](LICENSE)
