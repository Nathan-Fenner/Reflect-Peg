package parse

import (
	"fmt"
	"reflect"
)

// Embedding a Choice in a struct indicates that it's an alternation.
type Choice struct {
	Choice string
	Index  int
}

type ParseInto interface {
	ParseInto(*State, reflect.StructTag) error
}

type ParseCheck interface {
	Verify() error
}

type ParseFail interface {
	Failed()
}

var ParseIntoType = reflect.TypeOf((*ParseInto)(nil)).Elem()
var ParseFailType = reflect.TypeOf((*ParseFail)(nil)).Elem()

func parseIntoType(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	input := Input{state.Position, into}
	if output, ok := state.Memory[input]; ok {
		state.Position = output.Position
		return output.Result, output.Error
	}
	if state.Learning[Input{state.Position, into}] {
		panic(fmt.Sprintf("An infinite loop has occurred- parsing %+v [%s] at %d", into, tag, state.Position))
	}
	state.Learning[Input{state.Position, into}] = true
	oldPosition := state.Position
	value, err := parseIntoTypeCheck(state, into, tag)
	if err != nil {
		// Restore its position.
		state.Position = oldPosition
	}
	state.Memory[input] = Output{value, err, state.Position}
	return value, err
}

type fatalErrorNeedsLocation struct {
	Message interface{}
}

type fatalError struct {
	Location string
	Message  interface{}
}

func Panic(object interface{}) {
	panic(fatalErrorNeedsLocation{
		Message: object,
	})
}

func Panicf(format string, arguments ...interface{}) {
	Panic(fmt.Sprintf(format, arguments...))
}

func parseIntoTypeCheck(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	defer func() {
		value := recover()
		if value != nil {
			if fatal, ok := value.(fatalErrorNeedsLocation); ok {
				panic(fatalError{
					Location: state.Location(),
					Message:  fatal.Message,
				})
			}
			// Lift the error if it's not the one we're looking for.
			panic(value)
		}
	}()
	value, err := parseIntoTypeRaw(state, into, tag)
	if err != nil {
		if into.Implements(ParseFailType) {
			// This allows it to do whatever it needs to.
			// TODO: does it need context somehow?
			reflect.Zero(into).Interface().(ParseFail).Failed()
		}
		return nil, err
	}
	if checker, ok := value.(ParseCheck); ok {
		if err := checker.Verify(); err != nil {
			return nil, err
		}
	}
	return value, nil
}

func parseIntoTypeRaw(state *State, into reflect.Type, tag reflect.StructTag) (interface{}, error) {
	if into == reflect.TypeOf(Choice{}) {
		panic("Asked to parse into parse.Choice: probably a mistake.")
	}
	if pointerInto := reflect.PtrTo(into); pointerInto.Implements(ParseIntoType) {
		value := reflect.New(pointerInto.Elem())
		err := value.Interface().(ParseInto).ParseInto(state, tag)
		if err != nil {
			return nil, err
		}
		return value.Elem().Interface(), nil
	}
	if into.Kind() == reflect.Ptr {
		// Optional
		return parseOptional(state, into.Elem(), tag)
	}
	if into.Kind() == reflect.Slice {
		// Many
		return parseMany(state, into.Elem(), tag)
	}
	if into.Kind() == reflect.Chan {
		switch into.ChanDir() {
		case reflect.RecvDir:
			// Lookahead.
			return parseLookahead(state, into.Elem(), tag)
		case reflect.SendDir:
			return parseNegative(state, into.Elem(), tag)
		case reflect.BothDir:
			panic("Cannot parse into both-way channel (maybe you meant a receive-only channel?)")
		}
	}
	if into.Kind() != reflect.Struct {
		panic(fmt.Sprintf("into type %+v is not a pointer, slice, struct, or ParseInto.", into))
	}
	if into.NumField() > 0 && into.Field(0).Type == reflect.TypeOf(Choice{}) {
		// Alternation
		return parseAlternation(state, into, tag)
	}
	// Sequence
	return parseSequence(state, into, tag)
}

func parseIntoTypeCapture(state *State, into reflect.Type) (outValue interface{}, outErr error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			return
		}
		fatal, ok := recovered.(fatalError)
		if !ok {
			panic(recovered)
		}
		if asString, ok := fatal.Message.(string); ok {
			outErr = fmt.Errorf("%s at %s", asString, fatal.Location)
		}
		if asError, ok := fatal.Message.(error); ok {
			outErr = fmt.Errorf("%s at %s", asError.Error(), fatal.Location)
		}
		// TODO: stringer?
		outErr = fmt.Errorf("%+v at %s", fatal.Message, fatal.Location)
	}()
	return parseIntoType(state, into, "")
}

func Parse(source string, target interface{}) error {
	pointer := reflect.ValueOf(target)
	if pointer.Kind() != reflect.Ptr {
		panic("Parse given non-pointer.")
	}
	state := &State{
		Source:   []byte(source),
		Position: 0,
		Memory:   map[Input]Output{},
		Learning: map[Input]bool{},
	}
	value, err := parseIntoTypeCapture(state, pointer.Type().Elem())
	if err != nil {
		return err
	}
	pointer.Elem().Set(reflect.ValueOf(value))
	return nil
}
