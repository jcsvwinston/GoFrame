package quark

import (
	"reflect"
	"sync"
)

// ModelMeta holds cached metadata about a model struct.
// Computed once per type and stored in a global registry.
type ModelMeta struct {
	Table      string
	PK         pkMeta
	HasPK      bool
	Fields     []FieldMeta
	FieldByCol map[string]*FieldMeta // lookup by db column name
}

// FieldMeta holds metadata about a single struct field.
type FieldMeta struct {
	Index  int
	Column string // value of the db:"" tag
	Kind   reflect.Kind
	Type   reflect.Type
	IsPK   bool
}

// modelRegistry caches ModelMeta by reflect.Type.
var modelRegistry sync.Map // map[reflect.Type]*ModelMeta

// GetModelMeta returns the cached metadata for model type T.
// If not cached, it computes and stores it.
func GetModelMeta[T any]() *ModelMeta {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Fast path: already cached
	if cached, ok := modelRegistry.Load(t); ok {
		return cached.(*ModelMeta)
	}

	// Slow path: compute metadata
	meta := computeModelMeta(t)
	actual, _ := modelRegistry.LoadOrStore(t, meta)
	return actual.(*ModelMeta)
}

// computeModelMeta builds ModelMeta from a reflect.Type.
func computeModelMeta(t reflect.Type) *ModelMeta {
	meta := &ModelMeta{
		Table:      toSnakeCase(pluralize(t.Name())),
		FieldByCol: make(map[string]*FieldMeta),
	}

	// Find PK: first look for pk:"true", then fall back to db:"id"
	pkIndex := -1
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Tag.Get("pk") == "true" {
			pkIndex = i
			break
		}
	}
	if pkIndex == -1 {
		for i := 0; i < t.NumField(); i++ {
			if t.Field(i).Tag.Get("db") == "id" {
				pkIndex = i
				break
			}
		}
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag == "" || dbTag == "-" {
			continue
		}

		isPK := i == pkIndex
		fm := FieldMeta{
			Index:  i,
			Column: dbTag,
			Kind:   field.Type.Kind(),
			Type:   field.Type,
			IsPK:   isPK,
		}
		meta.Fields = append(meta.Fields, fm)
		meta.FieldByCol[dbTag] = &meta.Fields[len(meta.Fields)-1]

		if isPK {
			meta.PK = pkMeta{column: dbTag, index: i, kind: field.Type.Kind()}
			meta.HasPK = true
		}
	}

	return meta
}
