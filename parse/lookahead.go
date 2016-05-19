package parse

import "reflect"

// It's given a non-channel type, and returns a receiving-channel.
func parseLookahead(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	oldPosition := state.Position
	value, err := parseIntoType(state, into, tag)
	state.Position = oldPosition
	if err != nil {
		return nil, err
	}
	channel := reflect.MakeChan(reflect.ChanOf(reflect.BothDir, into), 1)
	channel.Send(reflect.ValueOf(value))
	return channel.Interface(), nil
}
