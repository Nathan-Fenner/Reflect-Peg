package parse

import "reflect"

// It's given a non-slice type, and returns a slice.
func parseMany(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	slice := reflect.Zero(reflect.SliceOf(into))
	for {
		value, err := parseIntoType(state, into, tag)
		if err != nil {
			break
		}
		slice = reflect.Append(slice, reflect.ValueOf(value))
	}
	return slice.Interface(), nil
}
