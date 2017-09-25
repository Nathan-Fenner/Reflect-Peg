package peg

import (
	"fmt"
	"reflect"
)

// TODO: implement alternation
// TODO: memoize parsing to improve speed
// TODO: improve error messages
// TODO: allow commits

// A Location represents a place in the source.
type Location struct {
	Line   int
	Column int
}

// String converts the location into a readable string.
func (l Location) String() string {
	return fmt.Sprintf("%d:%d", l.Line, l.Column)
}

// A Literal expects a tag to indicate its allowable value(s).
type Literal struct {
	Source []byte
	At     Location
}

// ByteParse is used by the peg package to directly extract from the input.
func (literal *Literal) ByteParse(source []byte, here Location, tag []byte) (int, error) {
	if len(source) < len(tag) {
		// TODO: handle whitespace
		return 0, fmt.Errorf("expected `%s` but got `%s` (end of input)", tag, source)
	}
	for i := range tag {
		if source[i] != tag[i] {
			// TODO: handle whitespace
			return 0, fmt.Errorf("expected `%s` but got `%s...`", tag, source[:len(tag)])
		}
	}
	literal.Source = source[:len(tag)]
	literal.At = here
	return len(tag), nil
}

// A ByteParser reads source directly for parsing.
// They receive the tag used to describe them in their parent struct as a third parameter.
type ByteParser interface {
	ByteParse(source []byte, here Location, tag []byte) (int, error)
}

// SomeType is used as a placeholder for any type.
type SomeType struct{}

// A FromParser is a type that can be constructed from an instance of another.
// The type must implement FromParse as a pointer receiver.
// First, an instance of the SomeType will be parsed. Then it will be passed
// as a parameter to this method, which should initialize its own fields. In the event that a
// parsing error occurs, returning an error indicates that the parser should backtrack.
// In addition, the tag for the field will be provided if present.
type FromParser interface {
	FromParse(source SomeType, here Location, tag []byte) error
}

// Context is used to configure the parser.
// Interface alternations are provided through it.
type Context struct {
	Alternates map[interface{}][]interface{}
}

// ParseInto takes a pointer to a value and parses the source provided into it,
// using the shape of the type.
func ParseInto(target interface{}, source []byte, context Context) error {
	_, err := parseIntoField(reflect.ValueOf(target), source, Location{}, nil, context)
	return err
}

// parseIntoField expects a pointer to a value of parseable type.
// If parsing succeeds, then the returned error will be nil and the target will be assigned to the
// parsed value.
func parseIntoField(target reflect.Value, source []byte, here Location, tag []byte, context Context) ([]byte, error) {
	if target.Type().Kind() != reflect.Ptr {
		panic(fmt.Sprintf("ParseInto is called with non-pointer %+v", target))
	}
	if byteParser, ok := target.Interface().(ByteParser); ok {
		n, err := byteParser.ByteParse(source, here, tag)
		if err != nil {
			return nil, err
		}
		return source[n:], nil
	}
	if target.Type().Elem().Kind() != reflect.Interface && target.Type().Elem().Kind() != reflect.Ptr {
		// only makes sense to look up methods on types that can have methods.
		if method, ok := target.Type().MethodByName("FromParse"); ok {
			if method.Func.Type().NumIn() == 4 && method.Func.Type().In(2) == reflect.TypeOf(Location{}) && method.Func.Type().In(3) == reflect.TypeOf([]byte{}) && method.Func.Type().NumOut() == 1 && method.Func.Type().Out(0) == reflect.TypeOf((*error)(nil)).Elem() {
				// Construct an object of the 'from' type, and parse into it.
				fromPointer := reflect.New(method.Func.Type().In(1))
				rest, err := parseIntoField(fromPointer, source, here, nil, context)
				if err != nil {
					return nil, err
				}
				// Use the parsed object as a parameter to the FromParse method on the target.
				if err := method.Func.Call([]reflect.Value{target, fromPointer.Elem(), reflect.ValueOf(here), reflect.ValueOf(tag)})[0].Interface(); err != nil {
					return nil, err.(error)
				}
				return rest, nil
			}
		}
	}
	if target.Type().Elem().Kind() == reflect.Slice {
		collected := reflect.MakeSlice(target.Type().Elem(), 0, 0)
		there := here
		for {
			eachPointer := reflect.New(target.Type().Elem().Elem())
			rest, err := parseIntoField(eachPointer, source, there, tag, context)
			if err != nil {
				break
			}
			collected = reflect.Append(collected, eachPointer.Elem())
			source = rest
		}
		target.Elem().Set(collected)
		return source, nil
	}
	if target.Type().Elem().Kind() == reflect.Ptr {
		optionalPointer := reflect.New(target.Type().Elem().Elem())
		rest, err := parseIntoField(optionalPointer, source, here, tag, context)
		if err == nil {
			optional := reflect.New(target.Type().Elem())
			optional.Elem().Set(optionalPointer.Elem())
			target.Elem().Set(optional)
			return rest, nil
		}
		return source, nil
	}
	if target.Type().Elem().Kind() == reflect.Chan {
		forwardPointer := reflect.New(target.Type().Elem().Elem())
		_, err := parseIntoField(forwardPointer, source, here, tag, context)
		if err == nil {
			forward := reflect.MakeChan(target.Type().Elem(), 1)
			forward.Send(forwardPointer.Elem())
			target.Elem().Set(forward)
		}
		return source, nil
	}
	if target.Type().Elem().Kind() == reflect.Struct {
		result := reflect.New(target.Type().Elem())
		for i := 0; i < target.Type().Elem().NumField(); i++ {
			tag := target.Type().Elem().Field(i).Tag.Get("parse")
			tagBytes := []byte(nil)
			if tag != "" {
				tagBytes = []byte(tag)
			}
			rest, err := parseIntoField(result.Elem().Field(i).Addr(), source, here, tagBytes, context)
			if err != nil {
				return nil, err
			}
			source = rest
		}
		target.Elem().Set(result.Elem())
		return source, nil
	}
	panic(fmt.Sprintf("unsupported parse type %+v of kind %s", target.Type().Elem(), target.Type().Elem().Kind().String()))
}
