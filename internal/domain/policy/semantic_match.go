package policy

import (
	"reflect"
	"strings"
)

func (s SemanticMatchSpec) fieldsUsed() []string {
	v := reflect.ValueOf(s)
	t := v.Type()
	fields := make([]string, 0, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		field := v.Field(i)
		if field.IsZero() {
			continue
		}
		name := semanticYAMLName(t.Field(i))
		if name == "" {
			continue
		}
		fields = append(fields, name)
	}
	return fields
}

func semanticYAMLName(field reflect.StructField) string {
	tag := field.Tag.Get("yaml")
	if tag == "" || tag == "-" {
		return ""
	}
	name, _, _ := strings.Cut(tag, ",")
	return name
}
