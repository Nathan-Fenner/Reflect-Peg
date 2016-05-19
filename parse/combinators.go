package parse

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
)

// Literal is annotated with parse:"<literal>"
type Literal struct {
	Contents []byte
	Location string
}

func (l *Literal) ParseInto(state *State, tag reflect.StructTag) error {
	literal := tag.Get("parse")
	if literal == "" {
		panic(fmt.Sprintf("parse.Literal given illegal no 'parse' tag: %q", tag))
	}
	for i := 0; i < len(literal); i++ {
		if i+state.Position >= len(state.Source) || state.Source[i+state.Position] != literal[i] {
			return fmt.Errorf("Expected %q at %s", literal, state.Location())
		}
	}
	l.Contents = []byte(tag.Get("parse"))
	l.Location = state.Location()
	state.Position += len(literal)
	return nil
}

type Number struct {
	Number   float64
	Location string
}

// Matches 4e6, -4.e7, .6E20, -0e0
// Doesn't match .e7, for example.
var numberRegex = regexp.MustCompile(`-?[0-9]+\.?[0-9]*([eE]-?[0-9]+)?|-?[0-9]*\.?[0-9]+([eE]-?[0-9]+)?`)

func (n *Number) ParseInto(state *State, tag reflect.StructTag) error {
	matched := numberRegex.Find(state.Rest())
	if matched == nil {
		return fmt.Errorf("expected number at %s", state.Location())
	}
	number, err := strconv.ParseFloat(string(matched), 64)
	if err != nil {
		return fmt.Errorf("expected number at %s but %s", state.Location(), err.Error())
	}
	n.Number = number
	n.Location = state.Location()
	state.Position += len(matched)
	return nil
}

type Location struct {
	Location string
}

func (l *Location) ParseInto(state *State, tag reflect.StructTag) error {
	l.Location = state.Location()
	return nil
}

type Regex struct {
	Contents []byte
	Location string
}

func (r *Regex) ParseInto(state *State, tag reflect.StructTag) error {
	regex := regexp.MustCompile(tag.Get("regex"))
	matched := regex.Find(state.Rest())
	if matched == nil {
		return fmt.Errorf("expected string to match regex %q at %s", tag.Get("regex"), state.Location())
	}
	if len(matched) == 0 || &matched[0] == &state.Source[state.Position] {
		r.Contents = matched
		state.Position += len(matched)
		return nil
	}
	return fmt.Errorf("expected string to match regex %q at %s", tag.Get("regex"), state.Location())
}

type matching struct{}

type Open struct {
	Literal Literal `parse:"("`
}

func (o Open) Annotate(m matching) error {
	return fmt.Errorf("expected `)` to match `(` opened at %s", o.Literal.Location)
}

type Close struct {
	Literal Literal `parse:")"`
}

func (c Close) Failed() {
	Panic(matching{})
}
