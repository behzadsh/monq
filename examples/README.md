# monq by example

A runnable program that uses `monq` against a real MongoDB, one demo per area of the API. Every step prints two things: the `bson.D` that monq built, and
the documents MongoDB returned for it. The point is to show the whole path from a Go call to a server answer, since the interesting part of a query builder
is what the driver does with what it builds.

## Running it

The only requirement is a MongoDB to talk to:

```sh
docker run --rm -p 27017:27017 mongo:8
go run ./examples
```

Pick one or more demos by name to run only those, in the order named:

```sh
go run ./examples find update
go run ./examples aggregate
```

Point it somewhere else with flags or the `MONGODB_URI` environment variable:

```sh
go run ./examples -uri "mongodb+srv://user:pass@cluster.example.net" -db scratch
MONGODB_URI="mongodb://localhost:27018" go run ./examples geo
```

The database (`monq_examples` by default) is dropped and reseeded before each demo, so the demos are independent of each other and of anything a previous
run left behind. Nothing outside that database is touched, and the seeded data is left in place at the end if you want to query it yourself.

## The demos

| Demo        | What it covers                                                                                                        |
|-------------|-----------------------------------------------------------------------------------------------------------------------|
| `find`      | comparison, logical, element, evaluation, array, and bitwise query operators, plus `Raw`, `CountDocuments`, `Distinct` |
| `project`   | projections built from `Include`, `Exclude`, `Slice`, `SliceFrom`, `Meta`, and sort documents that keep their order    |
| `update`    | field, array, and bitwise update operators, `Update` merging, upserts, array filters, `FindOneAndUpdate`               |
| `aggregate` | pipeline stages, from `$match` and `$group` through `$lookup`, `$graphLookup`, `$densify`, `$merge`, and `$out`        |
| `compute`   | the expression operators stages compute with: dates, strings, conversions, sets, objects, statistics, windows          |
| `geo`       | GeoJSON built by `Point`, `Polygon`, and `GeoJSON`, the query operators taking it, and the `$geoNear` stage            |
| `index`     | index models from `monq/index`, including unique, partial, hashed, text, and TTL indexes, and an `explain` of one      |
| `paths`     | the field path constants `monqgen` generates, and the four ways to name a position inside an array                     |

## The files

| File                                 | What is in it                                                                          |
|--------------------------------------|----------------------------------------------------------------------------------------|
| `main.go`                            | the runner: flags, connection, demo selection                                          |
| `model.go`                           | the structs the collections store, with the `go:generate` lines for `monqgen`          |
| `product_paths.go`, `order_paths.go` | what `monqgen` wrote from those structs, committed rather than generated at build time |
| `seed.go`                            | the fixture documents, and the two indexes the geo and text demos query through        |
| `print.go`                           | the printing helpers, so each demo reads as a list of examples rather than plumbing    |
| `find.go` and the rest               | one file per demo, named after it                                                      |

## Things worth noticing

- Nothing in here wraps the driver. Every monq function returns a `bson.D`, and it goes into `Find`, `Aggregate`, `UpdateMany`, `CreateOne`, and the rest
  with no adapter in between. `showFind` and friends in `print.go` are about printing, not about querying.
- The paths come from generated constants (`ProductPaths.Ratings.Score`), which is what makes a renamed `bson` tag a compile error instead of a query that
  quietly matches nothing. Hand-written strings work exactly as well: the `paths` demo shows both.
- Two places take a path that is not a document path, and both are MongoDB rules rather than monq ones: the conditions inside `$elemMatch` and `$pull` are
  relative to the array element, so they read `"score"` rather than `"ratings.score"`.
- One `$` is a field reference and two are a variable, so `expr.Field("price")` gives `"$price"` while the element inside a `$map` is the literal string
  `"$$this"`. monq has no constructor for the second, since a variable is not a field path.
- Coordinates are `[longitude, latitude]` everywhere, which is the reverse of the order they are usually quoted in.
- This program has no tests of its own: it is compiled by CI, and verified by running it. Codecov ignores the directory, but a local
  `go test -coverprofile` over `./...` still counts these files as uncovered, which is why the number it prints is lower than the library's own.
