package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"

	"github.com/behzadsh/monq"
)

// runUpdate covers the update operators. Each one returns a finished update document on its own, and Update is
// what puts several together: two $set documents concatenated by hand would be a duplicate key that MongoDB does
// not merge, so Update merges the ones sharing an operator.
func runUpdate(ctx context.Context, db *mongo.Database) error {
	return runParts(ctx, db.Collection(productsColl),
		updateFields,
		updateUpsert,
		updateArrays,
		updateArrayPositions,
		updateBitwiseAndReturn,
	)
}

// skuOnly is the projection the update demo reads results back through.
func skuOnly(fields ...monq.FieldPath) bson.D {
	return monq.Projection(monq.Include(append([]monq.FieldPath{ProductPaths.SKU}, fields...)...), monq.Exclude(ProductPaths.ID))
}

// updateFields shows the field update operators and the merging Update does.
func updateFields(ctx context.Context, coll *mongo.Collection) error {
	step("one operator is already a whole update document")

	if err := showUpdate(ctx, coll, `monq.Set(ProductPaths.Stock, 20)`, bySKU("kbd-001"), monq.Set(ProductPaths.Stock, 20)); err != nil {
		return err
	}

	step("Update merges the operators sharing a key, which is what makes several of them one document")

	both := monq.Update(
		monq.Set(ProductPaths.Category, "input"),
		monq.Inc(ProductPaths.Stock, -5),
		monq.Set(ProductPaths.Name, "Wireless Keyboard Mk II"),
		monq.CurrentDate(ProductPaths.CreatedAt),
	)
	if err := showUpdate(ctx, coll, "monq.Update(monq.Set(category, ...), monq.Inc(stock, -5), monq.Set(name, ...), monq.CurrentDate(created_at))",
		bySKU("kbd-001"), both); err != nil {
		return err
	}

	if err := showDoc(ctx, coll, "the document now", bySKU("kbd-001"), skuOnly(ProductPaths.Name, ProductPaths.Category, ProductPaths.Stock)); err != nil {
		return err
	}

	step("CurrentDateTimestamp is a sibling function rather than an option, since it writes a timestamp not a date")

	stamped := monq.CurrentDateTimestamp("synced_at")
	if err := showUpdate(ctx, coll, `monq.CurrentDateTimestamp("synced_at")`, bySKU("kbd-001"), stamped); err != nil {
		return err
	}

	step("$mul, $min, and $max compute the new value from the old one: $min only writes when its value is lower")

	clamped := monq.Update(
		monq.Mul(ProductPaths.Price, 0.9),
		monq.Min(ProductPaths.Stock, 3),
		monq.Max(ProductPaths.Flags, 16),
	)
	if err := showUpdate(ctx, coll, "monq.Update(monq.Mul(price, 0.9), monq.Min(stock, 3), monq.Max(flags, 16))", bySKU("kbd-001"), clamped); err != nil {
		return err
	}

	return showDoc(ctx, coll, "the document now", bySKU("kbd-001"), skuOnly(ProductPaths.Price, ProductPaths.Stock, ProductPaths.Flags))
}

// updateUpsert shows the operator that only applies when the update creates the document.
func updateUpsert(ctx context.Context, coll *mongo.Collection) error {
	step("$setOnInsert writes only when an upsert inserts, so a created document gets fields an updated one keeps")

	create := monq.Update(
		monq.Set(ProductPaths.Stock, 100),
		monq.SetOnInsert(ProductPaths.Name, "Mechanical Keyboard"),
		monq.SetOnInsert(ProductPaths.Category, "peripherals"),
		monq.SetOnInsert(ProductPaths.Price, 149.0),
	)
	if err := showUpdate(ctx, coll, "monq.Update(monq.Set(stock, 100), monq.SetOnInsert(name, ...), ...) with SetUpsert(true)",
		bySKU("kbd-999"), create, options.UpdateMany().SetUpsert(true)); err != nil {
		return err
	}

	if err := showDoc(ctx, coll, "the upserted document", bySKU("kbd-999"), skuOnly(ProductPaths.Name, ProductPaths.Price, ProductPaths.Stock)); err != nil {
		return err
	}

	step("$unset drops the field rather than nulling it, and $rename moves one, both on the document just created")

	moved := monq.Update(
		monq.Unset(ProductPaths.Stock),
		monq.Rename(ProductPaths.Name, "title"),
	)
	if err := showUpdate(ctx, coll, `monq.Update(monq.Unset(stock), monq.Rename(name, "title"))`, bySKU("kbd-999"), moved); err != nil {
		return err
	}

	return showDoc(ctx, coll, "the document now", bySKU("kbd-999"), skuOnly("title", ProductPaths.Stock))
}

// updateArrays shows the array update operators, including the $each form and its modifiers.
func updateArrays(ctx context.Context, coll *mongo.Collection) error {
	step("$push appends one value, $addToSet appends only what is not there yet")

	tags := monq.Update(
		monq.Push(ProductPaths.Tags.Path, "bluetooth"),
		monq.AddToSet(ProductPaths.Ratings.Path, Rating{User: "linus", Score: 4}),
	)
	if err := showUpdate(ctx, coll, `monq.Update(monq.Push(tags, "bluetooth"), monq.AddToSet(ratings, Rating{...}))`, bySKU("mse-002"), tags); err != nil {
		return err
	}

	step("PushEach is its own function because $position, $slice, and $sort only exist in the $each form")

	capped := monq.PushEach(ProductPaths.Ratings.Path,
		[]any{Rating{User: "ada", Score: 2}, Rating{User: "grace", Score: 5}},
		monq.PushSort(monq.Sort(monq.Desc("score"))),
		monq.PushSlice(3),
	)
	if err := showUpdate(ctx, coll, "monq.PushEach(ratings, values, monq.PushSort(monq.Sort(monq.Desc(score))), monq.PushSlice(3))",
		bySKU("mse-002"), capped); err != nil {
		return err
	}

	if err := showDoc(ctx, coll, "the top three ratings, highest first", bySKU("mse-002"), skuOnly(ProductPaths.Ratings.Path)); err != nil {
		return err
	}

	return updateArrayEnds(ctx, coll)
}

// updateArrayEnds shows the operators that add several elements at once and the ones that take elements away.
func updateArrayEnds(ctx context.Context, coll *mongo.Collection) error {
	step("AddToSetEach adds several values at once, and PushPosition inserts rather than appends")

	several := monq.Update(
		monq.AddToSetEach(ProductPaths.Tags.Path, []any{"wireless", "compact", "usb-c"}),
		monq.PushEach(ProductPaths.Ratings.Path, []any{Rating{User: "hopper", Score: 4}}, monq.PushPosition(0)),
	)
	if err := showUpdate(ctx, coll, "monq.Update(monq.AddToSetEach(tags, values), monq.PushEach(ratings, values, monq.PushPosition(0)))",
		bySKU("mse-002"), several); err != nil {
		return err
	}

	if err := showDoc(ctx, coll, "the arrays now", bySKU("mse-002"), skuOnly(ProductPaths.Tags.Path, ProductPaths.Ratings.Path)); err != nil {
		return err
	}

	step("$pop takes one element off an end: PopFirst and PopLast are separate functions since the value differs")

	popped := monq.Update(monq.PopFirst(ProductPaths.Ratings.Path), monq.PopLast(ProductPaths.Tags.Path))
	if err := showUpdate(ctx, coll, "monq.Update(monq.PopFirst(ratings), monq.PopLast(tags))", bySKU("mse-002"), popped); err != nil {
		return err
	}

	step("$pull takes a condition rather than a value, and it reads relative to the element, so the path is score")

	pulled := monq.Update(
		monq.Pull(ProductPaths.Ratings.Path, monq.Lt("score", 4)),
		monq.PullAll(ProductPaths.Tags.Path, []any{"bluetooth"}),
	)
	if err := showUpdate(ctx, coll, "monq.Update(monq.Pull(ratings, monq.Lt(score, 4)), monq.PullAll(tags, values))", bySKU("mse-002"), pulled); err != nil {
		return err
	}

	return showDoc(ctx, coll, "what is left", bySKU("mse-002"), skuOnly(ProductPaths.Ratings.Path, ProductPaths.Tags.Path))
}

// updateArrayPositions shows the four ways an element inside an array is named, which is what ArrayPath spells
// out so the punctuation does not have to be remembered.
func updateArrayPositions(ctx context.Context, coll *mongo.Collection) error {
	step(`"$" is the element the filter matched, so the filter has to contain a condition on the array`)

	matched := monq.And(bySKU("mon-003"), monq.Eq(ProductPaths.Ratings.User, "ada"))
	if err := showUpdate(ctx, coll, "monq.Set(ProductPaths.Ratings.Positional().Score, 1) // ratings.$.score",
		matched, monq.Set(ProductPaths.Ratings.Positional().Score, 1)); err != nil {
		return err
	}

	step(`"$[name]" picks the elements an array filter matches, and the filter is given to the driver, not to monq`)

	arrayFilters := options.UpdateMany().SetArrayFilters([]any{monq.Lt("low.score", 4)})
	if err := showUpdate(ctx, coll, "monq.Set(ProductPaths.Ratings.Filtered(\"low\").Score, 3) // ratings.$[low].score",
		bySKU("mon-003"), monq.Set(ProductPaths.Ratings.Filtered("low").Score, 3), arrayFilters); err != nil {
		return err
	}

	step(`"$[]" is every element, and At(i) is one fixed position`)

	everyOne := monq.Update(
		monq.Set(ProductPaths.Ratings.All().User, "anonymous"),
		monq.Set(ProductPaths.Tags.At(0), "uhd"),
	)
	if err := showUpdate(ctx, coll, "monq.Update(monq.Set(ratings.$[].user, ...), monq.Set(tags.0, ...))", bySKU("mon-003"), everyOne); err != nil {
		return err
	}

	return showDoc(ctx, coll, "the array now", bySKU("mon-003"), skuOnly(ProductPaths.Ratings.Path, ProductPaths.Tags.Path))
}

// updateBitwiseAndReturn shows the bitwise update operators, and that an update document is just a document: any
// driver call taking one accepts it.
func updateBitwiseAndReturn(ctx context.Context, coll *mongo.Collection) error {
	step("$bit sets, clears, or flips bits in place, one function per operation since they cannot share a field")

	if err := showUpdate(ctx, coll, "monq.BitOr(ProductPaths.Flags, 16)", bySKU("cam-007"), monq.BitOr(ProductPaths.Flags, 16)); err != nil {
		return err
	}

	if err := showUpdate(ctx, coll, "monq.BitAnd(ProductPaths.Flags, 30)", bySKU("cam-007"), monq.BitAnd(ProductPaths.Flags, 30)); err != nil {
		return err
	}

	if err := showUpdate(ctx, coll, "monq.BitXor(ProductPaths.Flags, 1)", bySKU("cam-007"), monq.BitXor(ProductPaths.Flags, 1)); err != nil {
		return err
	}

	step("FindOneAndUpdate takes the same document and hands back the document it wrote")

	update := monq.Update(monq.Inc(ProductPaths.Stock, 25), monq.Set(ProductPaths.Category, "restocked"))
	query("monq.Update(monq.Inc(stock, 25), monq.Set(category, ...))", update)

	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After).
		SetProjection(skuOnly(ProductPaths.Stock, ProductPaths.Category, ProductPaths.Flags))

	var updated bson.D
	if err := coll.FindOneAndUpdate(ctx, bySKU("cam-007"), update, opts).Decode(&updated); err != nil {
		return fmt.Errorf("find one and update: %w", err)
	}

	fmt.Printf("   -> after the write: %s\n", jsonOf(updated))

	return nil
}
