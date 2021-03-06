package peg

import "testing"

func TestPass(t *testing.T) {
	type ExampleAB struct {
		A Literal `parse:"A"`
		B Literal `parse:"B"`
	}

	var exampleAB ExampleAB
	err := ParseInto(&exampleAB, []byte("AB"), Context{})
	if err != nil {
		t.Errorf("error ``%s'' unexpected", err)
	}

	type ExampleAAABC struct {
		A []Literal `parse:"A"`
		B Literal   `parse:"B"`
		C Literal   `parse:"C"`
	}
	var exampleAAABC ExampleAAABC
	err = ParseInto(&exampleAAABC, []byte("AAAAAABC"), Context{})
	if err != nil {
		t.Errorf("error ``%s'' unexpected", err)
	}
	var exampleBC ExampleAAABC
	err = ParseInto(&exampleBC, []byte("BC"), Context{})
	if err != nil {
		t.Errorf("error ``%s'' unexpected", err)
	}
}

func TestReject(t *testing.T) {
	type ExampleAB struct {
		A Literal `parse:"A"`
		B Literal `parse:"B"`
	}

	var exampleAB ExampleAB
	err := ParseInto(&exampleAB, []byte("A"), Context{})
	if err == nil {
		t.Errorf("error expected")
	}

	type ExampleAAABC struct {
		A []Literal `parse:"A"`
		B Literal   `parse:"B"`
		C Literal   `parse:"C"`
	}
	var exampleAAABC ExampleAAABC
	err = ParseInto(&exampleAAABC, []byte("AAAAAACB"), Context{})
	if err == nil {
		t.Errorf("error expected")
	}
}

func TestAlternation(t *testing.T) {
	type EitherAB interface {
	}

	type A struct {
		A Literal `parse:"A"`
	}
	type B struct {
		B Literal `parse:"B"`
	}

	type Example struct {
		Sequence []EitherAB
	}

	var example Example

	err := ParseInto(&example, []byte("ABBBABAB"), Context{
		Alternates: AlternateMap{
			new(EitherAB): {new(A), new(B)},
		},
	})

	if err != nil {
		t.Errorf("error ``%s'' unexpected", err.Error())
	}
}
