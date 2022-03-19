package modules

import (
	"reflect"
	"strings"

	"github.com/destructiqn/kogtevran/generic"
)

const optionTag = "option"

func GetOptionValue(module generic.Module, name string) (interface{}, bool) {
	_, value, ok := getField(reflect.ValueOf(module), name)
	if ok {
		return value.Interface(), true
	}

	return nil, false
}

func SetOptionValue(module generic.Module, name string, newValue interface{}) bool {
	_, value, ok := getField(reflect.ValueOf(module), name)
	if !ok {
		return false
	}

	value.Set(reflect.ValueOf(newValue))
	return true
}

func getField(value reflect.Value, name string) (reflect.StructField, reflect.Value, bool) {
	name = strings.ToLower(name)

	if value.Kind() != reflect.Ptr {
		panic("not pointer value was passed to getField")
	}

	value = reflect.Indirect(value)
	if value.Kind() != reflect.Struct {
		return reflect.StructField{}, reflect.Value{}, false
	}

	for i := 0; i < value.NumField(); i++ {
		if value.Field(i).Kind() == reflect.Struct {
			innerField, innerValue, ok := getField(value.Field(i).Addr(), name)
			if ok {
				return innerField, innerValue, ok
			}
		}

		field := reflect.TypeOf(value.Interface()).Field(i)
		tag, ok := field.Tag.Lookup(optionTag)
		if !ok {
			continue
		}

		if strings.ToLower(tag) == name {
			return field, value.Field(i), true
		}
	}

	return reflect.StructField{}, reflect.Value{}, false
}
