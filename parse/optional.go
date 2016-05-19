package parse

import "reflect"

// It's given a non-pointer type, and returns a pointer.
func parseOptional(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	value, err := parseIntoType(state, into, tag)
	if err != nil {
		return reflect.Zero(reflect.PtrTo(into)).Interface(), nil
	}
	pointer := reflect.New(into)
	pointer.Elem().Set(reflect.ValueOf(value))
	return pointer.Interface(), nil
}
