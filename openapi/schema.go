package openapi

import (
	"encoding/json"
	"reflect"
	"strings"
)

// Minimal schema model (enough for most JSON DTOs).
type Schema struct {
	Ref         string                `json:"$ref,omitempty"`
	Type        string                `json:"type,omitempty"`
	Format      string                `json:"format,omitempty"`
	Enum        []any                 `json:"enum,omitempty"`
	Items       *SchemaRef            `json:"items,omitempty"`
	Properties  map[string]*SchemaRef `json:"properties,omitempty"`
	Required    []string              `json:"required,omitempty"`
	Description string                `json:"description,omitempty"`
}

type SchemaRef struct {
	Ref    string  `json:"$ref,omitempty"`
	Schema *Schema `json:"-"`
}

// ensure SchemaRef serializes either $ref or the embedded Schema
func (sr *SchemaRef) MarshalJSON() ([]byte, error) {
	if sr == nil {
		return []byte("null"), nil
	}
	if sr.Ref != "" {
		type refOnly struct {
			Ref string `json:"$ref"`
		}
		return json.Marshal(refOnly{Ref: sr.Ref})
	}
	if sr.Schema == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(sr.Schema)
}

func Ref(s Schema) *SchemaRef { return &SchemaRef{Schema: &s} }

func SchemaFrom(v any) *SchemaRef {
	if v == nil {
		return Ref(Schema{Type: "object"})
	}
	t := reflect.TypeOf(v)
	return schemaOfType(t)
}

func schemaOfType(t reflect.Type) *SchemaRef {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	switch t.Kind() {
	case reflect.Bool:
		return Ref(Schema{Type: "boolean"})
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		return Ref(Schema{Type: "integer", Format: "int32"})
	case reflect.Int64:
		return Ref(Schema{Type: "integer", Format: "int64"})
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return Ref(Schema{Type: "integer", Format: "int32"})
	case reflect.Uint64:
		return Ref(Schema{Type: "integer", Format: "int64"})
	case reflect.Float32:
		return Ref(Schema{Type: "number", Format: "float"})
	case reflect.Float64:
		return Ref(Schema{Type: "number", Format: "double"})
	case reflect.String:
		return Ref(Schema{Type: "string"})
	case reflect.Slice, reflect.Array:
		return &SchemaRef{Schema: &Schema{
			Type:  "array",
			Items: schemaOfType(t.Elem()),
		}}
	case reflect.Map:
		return Ref(Schema{Type: "object"})
	case reflect.Struct:
		props := map[string]*SchemaRef{}
		var req []string
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name := jsonName(f)
			if name == "-" {
				continue
			}
			props[name] = schemaOfType(f.Type)
			if strings.Contains(f.Tag.Get("validate"), "required") {
				req = append(req, name)
			}
		}
		return &SchemaRef{Schema: &Schema{
			Type:       "object",
			Properties: props,
			Required:   req,
		}}
	default:
		return Ref(Schema{Type: "string"})
	}
}

func jsonName(f reflect.StructField) string {
	tag := f.Tag.Get("json")
	if tag == "" {
		return lowerCamel(f.Name)
	}
	part := strings.Split(tag, ",")[0]
	if part == "" {
		return f.Name
	}
	return part
}

func lowerCamel(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}
