package parse

import (
	"fmt"
	"reflect"
)

// It's given a non-channel type, and returns a sending-channel.
func parseNegative(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	if tag.Get(`name`) == "" {
		panic(fmt.Sprintf("Field %+v should have `name` tag: %q", into, tag))
	}
	// Negative lookahead.
	_, err := parseIntoType(state, into.Elem(), tag)
	if err == nil {
		return nil, fmt.Errorf("expected %s to fail", tag.Get(`name`))
	}
	return reflect.Zero(into).Interface(), nil
}
