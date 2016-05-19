package parse

import (
	"fmt"
	"reflect"
)

// parseAlternation is given a struct type whose first element is a Choice.
func parseAlternation(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	name := tag.Get("name")
	if name == "" {
		name = into.Field(0).Tag.Get("name")
	}
	if name == "" {
		panic(fmt.Sprintf("cannot parse alternative %+v that has no name (either annotated as tag where used as a field, or on `Choice` field.)", into))
	}
	for i := 1; i < into.NumField(); i++ {
		result, err := parseIntoTypeRaw(state, into.Field(i).Type, into.Field(i).Tag)
		if err == nil {
			value := reflect.New(into).Elem()
			value.Field(0).Set(reflect.ValueOf(Choice{into.Field(i).Name, i}))
			value.Field(i).Set(reflect.ValueOf(result))
			return value.Interface(), nil
		}
	}
	return nil, fmt.Errorf("Expected %s at %s", into.Field(0).Tag.Get("name"), state.Location())
}
