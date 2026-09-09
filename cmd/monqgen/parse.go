package main

import (
	"fmt"
	"go/types"
	"reflect"
	"strings"

	"golang.org/x/tools/go/packages"
)

// Kind says what a field contributes to the path tree.
type Kind int

const (
	// KindScalar is a field with no paths underneath it: a number, a string, a date, an ObjectID, binary data, or
	// any type that marshals itself.
	KindScalar Kind = iota

	// KindDocument is an embedded document whose own fields carry on the path.
	KindDocument

	// KindArray is a slice or an array. When its elements are documents, Elem holds their fields.
	KindArray

	// KindMap is a document whose keys are not known until run time, so the field itself has a path and nothing
	// under it does.
	KindMap
)

// String returns the kind's name, which is what the model prints.
func (k Kind) String() string {
	switch k {
	case KindScalar:
		return "scalar"
	case KindDocument:
		return "document"
	case KindArray:
		return "array"
	case KindMap:
		return "map"
	default:
		return "unknown"
	}
}

// Field is one field of a struct, named as MongoDB sees it.
type Field struct {
	// GoName is the field's name in Go, which the generated identifier is built from.
	GoName string

	// Key is the field's name in the document, from the bson tag or the field name lowercased.
	Key string

	// Kind says whether anything lives underneath this field.
	Kind Kind

	// Elem holds the fields of a document, or of an array's element when that element is a document. It is nil
	// for a scalar, for an array of scalars, and for a walk stopped by Recursive.
	Elem *Struct

	// Recursive marks a field whose type is already being walked further up the tree. The walk stops there to keep
	// the tree finite, so paths below it are not generated.
	Recursive bool
}

// Struct is a document shape: a name and the fields directly under it.
//
// Nothing here records a full dotted path. The tree does that, and keeping it a tree is what lets a generator put
// something between one segment and the next, which is exactly what the positional operators of an array need.
type Struct struct {
	// Name is the Go type name, empty for the element of an array of anonymous structs.
	Name string

	// Fields are in the order they are declared, with inlined fields spliced in where they appear.
	Fields []Field
}

// Walk calls fn for every field in the tree, passing the document keys leading to it, that field included.
//
// A field of an array of documents is reached through the array's own key, so the path of an Item's sku under an
// Items array is ["items", "sku"]. Where those two segments join is where a positional operator goes.
func (s *Struct) Walk(fn func(path []string, field Field)) {
	s.walkFrom(nil, fn)
}

func (s *Struct) walkFrom(prefix []string, fn func(path []string, field Field)) {
	for _, f := range s.Fields {
		path := append(append([]string{}, prefix...), f.Key)
		fn(path, f)

		if f.Elem != nil {
			f.Elem.walkFrom(path, fn)
		}
	}
}

// Paths returns every dotted path in the tree, in declaration order. It is what the model looks like from the
// outside, and what the parser's tests compare against.
func (s *Struct) Paths() []string {
	var paths []string

	s.Walk(
		func(path []string, _ Field) {
			paths = append(paths, strings.Join(path, "."))
		},
	)

	return paths
}

// Parse resolves typeName in pkg and walks it into a path tree.
//
// It follows the driver's rules rather than encoding/json's: a key defaults to the field name lowercased whole, a
// tag of "-" drops the field, and an embedded struct nests under its own name unless the tag says ",inline".
// Unexported fields never appear in a document and are skipped.
func Parse(pkg *packages.Package, typeName string) (*Struct, error) {
	obj := pkg.Types.Scope().Lookup(typeName)
	if obj == nil {
		return nil, fmt.Errorf("type %s not found in package %s", typeName, pkg.PkgPath)
	}

	named, ok := obj.Type().(*types.Named)
	if !ok {
		return nil, fmt.Errorf("%s is not a named type", typeName)
	}

	structType, ok := named.Underlying().(*types.Struct)
	if !ok {
		return nil, fmt.Errorf("%s is not a struct", typeName)
	}

	p := &parser{seen: map[string]bool{named.String(): true}}

	return &Struct{Name: typeName, Fields: p.fields(structType)}, nil
}

// parser carries the set of named types already being walked, so a type that refers to itself stops rather than
// recursing forever.
type parser struct {
	seen map[string]bool
}

// fields walks one struct, splicing inlined fields into the list at the point they are declared.
func (p *parser) fields(s *types.Struct) []Field {
	var fields []Field

	for i := range s.NumFields() {
		v := s.Field(i)
		if !v.Exported() {
			continue
		}

		tag := parseTag(v.Name(), s.Tag(i))
		if tag.skip {
			continue
		}

		if tag.inline {
			fields = append(fields, p.inlined(v.Type())...)

			continue
		}

		fields = append(fields, p.field(v.Name(), tag.key, v.Type()))
	}

	return fields
}

// inlined returns the fields an inlined member contributes to its parent. An inlined map has dynamic keys, so it
// contributes nothing that can be named ahead of time.
func (p *parser) inlined(t types.Type) []Field {
	switch under := deref(t).Underlying().(type) {
	case *types.Struct:
		return p.fields(under)
	default:
		return nil
	}
}

// field describes one field by its type.
func (p *parser) field(goName, key string, t types.Type) Field {
	f := Field{GoName: goName, Key: key}

	t = deref(t)
	if isLeaf(t) {
		f.Kind = KindScalar

		return f
	}

	switch under := t.Underlying().(type) {
	case *types.Slice:
		return p.arrayField(f, under.Elem())
	case *types.Array:
		return p.arrayField(f, under.Elem())
	case *types.Map:
		f.Kind = KindMap

		return f
	case *types.Struct:
		f.Kind = KindDocument
		f.Elem, f.Recursive = p.document(t, under)

		return f
	default:
		f.Kind = KindScalar

		return f
	}
}

// arrayField fills in a slice or array field, describing its element when that element is a document.
func (p *parser) arrayField(f Field, elem types.Type) Field {
	f.Kind = KindArray

	elem = deref(elem)
	if isLeaf(elem) {
		return f
	}

	if under, ok := elem.Underlying().(*types.Struct); ok {
		f.Elem, f.Recursive = p.document(elem, under)
	}

	return f
}

// document walks a struct type unless it is already on the way down, in which case the walk stops there.
func (p *parser) document(t types.Type, s *types.Struct) (*Struct, bool) {
	name := ""
	if named, ok := t.(*types.Named); ok {
		name = named.Obj().Name()

		key := named.String()
		if p.seen[key] {
			return nil, true
		}

		p.seen[key] = true
		defer delete(p.seen, key)
	}

	return &Struct{Name: name, Fields: p.fields(s)}, false
}

// deref removes pointers, which change nothing about the document a value produces.
func deref(t types.Type) types.Type {
	for {
		ptr, ok := t.Underlying().(*types.Pointer)
		if !ok {
			return t
		}

		t = ptr.Elem()
	}
}

// isLeaf reports whether a type produces a value with no paths inside it worth naming.
//
// Byte slices become binary data rather than arrays. Types that marshal themselves, such as time.Time and the
// driver's own ObjectID and Decimal128, produce whatever their marshaler says, so their Go fields say nothing
// about the document. Interfaces are opaque for the same reason.
func isLeaf(t types.Type) bool {
	if isByteSlice(t) {
		return true
	}

	if _, ok := t.Underlying().(*types.Interface); ok {
		return true
	}

	named, ok := t.(*types.Named)
	if !ok {
		return false
	}

	if isDriverLeaf(named) {
		return true
	}

	return marshalsItself(named)
}

// isByteSlice reports whether t is []byte, which MongoDB stores as binary data rather than as an array.
func isByteSlice(t types.Type) bool {
	slice, ok := t.Underlying().(*types.Slice)
	if !ok {
		return false
	}

	basic, ok := slice.Elem().Underlying().(*types.Basic)

	return ok && basic.Kind() == types.Byte
}

// isDriverLeaf reports whether t is one of the types the driver encodes as a single BSON value.
func isDriverLeaf(named *types.Named) bool {
	obj := named.Obj()
	if obj.Pkg() == nil {
		return false
	}

	switch obj.Pkg().Path() {
	case "time":
		return obj.Name() == "Time" || obj.Name() == "Duration"
	case "go.mongodb.org/mongo-driver/v2/bson":
		switch obj.Name() {
		case "ObjectID", "DateTime", "Decimal128", "Binary", "Regex", "Timestamp", "JavaScript", "Symbol", "Raw",
			"RawValue", "Undefined", "MinKey", "MaxKey", "Null", "A":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// marshalsItself reports whether a type carries a BSON marshaler, in which case its fields say nothing about the
// document it produces.
func marshalsItself(named *types.Named) bool {
	for _, t := range []types.Type{named, types.NewPointer(named)} {
		ms := types.NewMethodSet(t)
		for i := range ms.Len() {
			switch ms.At(i).Obj().Name() {
			case "MarshalBSON", "MarshalBSONValue":
				return true
			}
		}
	}

	return false
}

// tag is what the bson struct tag says about one field.
type tag struct {
	key    string
	inline bool
	skip   bool
}

// parseTag applies the driver's rules to one struct tag, with goName supplying the default key.
//
// The key is the field name lowercased in full, so CreatedAt becomes createdat rather than created_at. A tag of
// exactly "-" drops the field. Flags other than inline change how a value is encoded rather than where it lives,
// so nothing here looks at them.
func parseTag(goName, raw string) tag {
	t := tag{key: strings.ToLower(goName)}

	value, ok := lookupTag(raw, "bson")
	if !ok {
		return t
	}

	if value == "-" {
		t.skip = true

		return t
	}

	for i, part := range strings.Split(value, ",") {
		if i == 0 && part != "" {
			t.key = part
		}

		if part == "inline" {
			t.inline = true
		}
	}

	return t
}

// lookupTag reads one key out of a raw struct tag, which types.Struct hands over as the whole tag string.
func lookupTag(raw, key string) (string, bool) {
	return reflect.StructTag(raw).Lookup(key)
}
