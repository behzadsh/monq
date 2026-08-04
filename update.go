package monq

import "go.mongodb.org/mongo-driver/v2/bson"

// Update combines update operators into a single update document.
//
// Every monq update operator sets one field, so each returns a one-field document such as {$set: {name: "ada"}},
// already usable on its own. Update is what puts several of them together: operators sharing a key are merged
// under it, so Update(Set("a", 1), Inc("b", 2), Set("c", 3)) yields {$set: {a: 1, c: 3}, $inc: {b: 2}}. Operators
// keep the order they were first seen in, and fields keep their order inside each operator. Update is to update
// documents what [And] is to filters.
//
// Concatenating two operators by hand instead ({$set: {a: 1}} followed by {$set: {c: 3}}) builds a document with
// two $set keys, which MongoDB does not merge. That is what Update exists to prevent.
//
// Setting the same field twice under one operator keeps the last value, silently, since MongoDB would reject the
// duplicate. Two different operators touching the same field ([Set] and [Unset] on it, say) are conflicting mods
// that the server rejects at execution time, because merging by operator key cannot detect them. Update with no
// arguments returns an empty document, which the driver rejects for the same reason it rejects any empty update.
//
// Example:
//
//	monq.Update(monq.Set("status", "active"), monq.Inc("logins", 1), monq.Set("name", "ada"))
//	// bson.D{
//	//     {Key: "$set", Value: bson.D{{Key: "status", Value: "active"}, {Key: "name", Value: "ada"}}},
//	//     {Key: "$inc", Value: bson.D{{Key: "logins", Value: 1}}},
//	// }
//
// MongoDB docs: https://www.mongodb.com/docs/manual/reference/operator/update/
func Update(ops ...bson.D) bson.D {
	update := bson.D{}
	at := make(map[string]int, len(ops))

	for _, op := range ops {
		for _, e := range op {
			i, seen := at[e.Key]
			if !seen {
				at[e.Key] = len(update)
				update = append(update, bson.E{Key: e.Key, Value: copyOperand(e.Value)})

				continue
			}

			update[i].Value = mergeOperand(update[i].Value, e.Value)
		}
	}

	return update
}

// copyOperand copies an operator's operand document, so merging never writes into a document the caller still
// holds. Operands that are not documents, which only [Raw] can produce, are taken as they are.
func copyOperand(operand any) any {
	fields, ok := operand.(bson.D)
	if !ok {
		return operand
	}

	out := make(bson.D, len(fields))
	copy(out, fields)

	return out
}

// mergeOperand folds incoming's fields into existing, overwriting fields already present. If either side is not a
// document, incoming replaces existing outright: there is nothing to merge field by field.
func mergeOperand(existing, incoming any) any {
	target, ok := existing.(bson.D)
	if !ok {
		return copyOperand(incoming)
	}

	fields, ok := incoming.(bson.D)
	if !ok {
		return copyOperand(incoming)
	}

	for _, f := range fields {
		if i := indexOfKey(target, f.Key); i >= 0 {
			target[i].Value = f.Value

			continue
		}

		target = append(target, f)
	}

	return target
}
