package parse

import (
	"fmt"
	"reflect"
)

type Output struct {
	Result   interface{}
	Error    error
	Position int
}

type Input struct {
	Position int
	Type     reflect.Type
}

type State struct {
	Source   []byte
	Position int
	Memory   map[Input]Output
	Learning map[Input]bool
}

func (s *State) Rest() []byte {
	return s.Source[s.Position:]
}

func (s *State) Location() string {
	line := 0
	column := 0
	for i := 0; i < s.Position; i++ {
		switch s.Source[i] {
		case '\r':
			column = 0
		case '\n':
			line++
			column = 0
		case '\t':
			column /= 4
			column++
			column *= 4
		default:
			column++
		}
	}
	return fmt.Sprintf("%d:%d", line+1, column+1)
}
