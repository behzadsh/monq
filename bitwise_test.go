package monq_test

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/behzadsh/monq"
)

func TestBitwiseOperators(t *testing.T) {
	tests := []struct {
		name  string
		got   bson.D
		field monq.FieldPath
		op    string
		mask  any
	}{
		{
			name:  "BitsAllClear with a numeric mask",
			got:   monq.BitsAllClear("flags", 6),
			field: "flags",
			op:    "$bitsAllClear",
			mask:  6,
		},
		{
			name:  "BitsAllSet with a position array",
			got:   monq.BitsAllSet("flags", bson.A{1, 5}),
			field: "flags",
			op:    "$bitsAllSet",
			mask:  bson.A{1, 5},
		},
		{
			name:  "BitsAnyClear with a numeric mask",
			got:   monq.BitsAnyClear("flags", 6),
			field: "flags",
			op:    "$bitsAnyClear",
			mask:  6,
		},
		{
			name:  "BitsAnySet with a position array",
			got:   monq.BitsAnySet("perms.bits", bson.A{1, 5}),
			field: "perms.bits",
			op:    "$bitsAnySet",
			mask:  bson.A{1, 5},
		},
		{
			name:  "BitsAllSet with binary data",
			got:   monq.BitsAllSet("flags", bson.Binary{Subtype: 0x00, Data: []byte{0x30}}),
			field: "flags",
			op:    "$bitsAllSet",
			mask:  bson.Binary{Subtype: 0x00, Data: []byte{0x30}},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				if tt.got[0].Key != string(tt.field) {
					t.Fatalf("field = %q, want %q", tt.got[0].Key, tt.field)
				}

				inner, ok := tt.got[0].Value.(bson.D)
				if !ok || inner[0].Key != tt.op {
					t.Fatalf("got %v, want a %s operator document", tt.got, tt.op)
				}

				if !reflect.DeepEqual(inner[0].Value, tt.mask) {
					t.Fatalf("mask = %v, want %v", inner[0].Value, tt.mask)
				}
			},
		)
	}
}

func TestBitwiseOperatorsComposeWithNot(t *testing.T) {
	got := monq.Not(monq.BitsAnySet("flags", 6))

	outer, ok := got[0].Value.(bson.D)
	if !ok || got[0].Key != "flags" || outer[0].Key != "$not" {
		t.Fatalf("Not() = %v, want a $not operator document on flags", got)
	}

	inner, ok := outer[0].Value.(bson.D)
	if !ok || inner[0].Key != "$bitsAnySet" {
		t.Fatalf("Not() inner operator = %v, want $bitsAnySet", outer[0].Value)
	}
}

func ExampleBitsAllClear() {
	filter := monq.BitsAllClear("flags", 6)

	printFilter(filter)
	// Output: {"flags":{"$bitsAllClear":6}}
}

func ExampleBitsAllSet() {
	filter := monq.BitsAllSet("flags", bson.A{1, 5})

	printFilter(filter)
	// Output: {"flags":{"$bitsAllSet":[1,5]}}
}

func ExampleBitsAnyClear() {
	filter := monq.BitsAnyClear("flags", 6)

	printFilter(filter)
	// Output: {"flags":{"$bitsAnyClear":6}}
}

func ExampleBitsAnySet() {
	filter := monq.BitsAnySet("flags", bson.A{1, 5})

	printFilter(filter)
	// Output: {"flags":{"$bitsAnySet":[1,5]}}
}
