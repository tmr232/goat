package python

import (
	"reflect"
	"strings"
)

func derefValue(value reflect.Value) reflect.Value {
	for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func deinterfaceValue(value reflect.Value) reflect.Value {
	if value.Kind() == reflect.Interface {
		if value.IsNil() {
			return reflect.Value{}
		}
		return value.Elem()
	}
	return value
}

func lookupValue(obj any, attrs ...string) (value reflect.Value, found bool) {
	if obj == nil {
		return reflect.Value{}, false
	}

	value = reflect.ValueOf(obj)

	for _, attr := range attrs {
		for value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface {
			if value.IsNil() {
				return reflect.Value{}, false
			}
			value = value.Elem()
		}
		field := value.FieldByName(attr)
		if field == *new(reflect.Value) {
			return reflect.Value{}, false
		}
		value = field
	}

	return value, true
}

func Lookup(obj any, attrs ...string) (attribute any, found bool) {
	value, found := lookupValue(obj, attrs...)
	attribute = value.Interface()
	return
}

func HasAttr(obj any, attr string) bool {
	attrs := strings.Split(attr, ".")

	_, found := Lookup(obj, attrs...)
	return found
}

func GetAttr(obj any, attr string) any {
	attrs := strings.Split(attr, ".")

	attribute, _ := Lookup(obj, attrs...)
	return attribute
}

func SetAttr(obj any, attr string, value any) bool {
	attrs := strings.Split(attr, ".")

	attrValue, found := lookupValue(obj, attrs...)
	if !found {
		return false
	}

	if !attrValue.CanSet() {
		return false
	}

	valueValue := reflect.ValueOf(value)
	if !valueValue.Type().AssignableTo(attrValue.Type()) {
		return false
	}

	attrValue.Set(valueValue)

	return true
}

func structMatchInternal(objValue reflect.Value, queryType reflect.Type) bool {
	if queryType.Field(0).Type != objValue.Type() {
		return false
	}

	for i := 1; i < queryType.NumField(); i++ {
		field := queryType.Field(i)

		attr := objValue.FieldByName(field.Name)
		if attr == *new(reflect.Value) {
			return false
		}

		if field.Type.Name() != "" {
			if field.Type != attr.Type() {
				return false
			} else {
				continue
			}
		}
		if !structMatchInternal(attr, field.Type) {
			return false
		}
	}
	return true
}

func StructMatch[Query any](obj any) bool {
	if obj == nil {
		return false
	}
	objValue := reflect.ValueOf(obj)
	queryType := reflect.TypeOf(*new(Query))

	return structMatchInternal(objValue, queryType)
}

func structQueryInternal(objValue reflect.Value, queryValue reflect.Value) bool {
	queryType := queryValue.Type()
	if queryValue.Type().Field(0).Type != objValue.Type() && deinterfaceValue(objValue).Type() != queryValue.Type().Field(0).Type {
		return false
	}

	typeField := queryType.Field(0)
	if typeField.Name != "_" {
		queryValue.Field(0).Set(deinterfaceValue(objValue))
	}

	for i := 1; i < queryType.NumField(); i++ {
		field := queryType.Field(i)

		attr := derefValue(objValue).FieldByName(field.Name)
		if attr == *new(reflect.Value) {
			return false
		}

		if field.Type.Name() != "" || field.Type.Kind() == reflect.Slice {
			if field.Type != attr.Type() {
				return false
			} else {
				// Let's do some assigning!
				queryValue.Field(i).Set(attr)
				continue
			}
		}
		if !structQueryInternal(attr, queryValue.Field(i)) {
			return false
		}
	}
	return true
}

func StructQuery[Query any](obj any) *Query {
	if obj == nil {
		return nil
	}

	query := new(Query)
	queryValue := reflect.ValueOf(query)
	objValue := reflect.ValueOf(obj)
	if structQueryInternal(objValue, queryValue.Elem()) {
		return query
	}
	return nil
}
