# Reflect-Peg
A reflect-based PEG parser for Go.

Reflect-PEG is a type-safe parsing library in Go. Instead of using code generation or an embedded language to perform parsing, it just inspects the shape of the types that it's asked to parse in order to build them from source.

It's based on the [parsing expression grammar (PEG)](https://en.wikipedia.org/wiki/Parsing_expression_grammar) semantics.

## Example

Suppose we want to match the toy grammar of "strings of A's and B's":

```

type AOrB struct {
    parse.Choice
    A parse.Literal `parse:"A"`
    B parse.Literal `parse:"B"`
}

type AsOrBs struct {
    AsOrBs []AOrB
}

```

Then we use the `Parse` function:

```
func Parse(source string, target interface{}) error
```

We call it like:

```
var result AsOrBs
err := parse.Parse("AABBA", &result)
if err != nil {
    fmt.Printf("Error: %s\n", err.Error()
    return
}
fmt.Printf("Result: %+v", result)
```

The result will look like:

```
AsOrBs {
    {
      Choice: {"A", 1},
      A: {Location: "1:1"},
    },
    {
      Choice: {"A", 1},
      A: {Location: "1:2"},
    },
    {
      Choice: {"B", 2},
      B: {Location: "1:3"},
    },
    {
      Choice: {"B", 2},
      B: {Location: "1:4"},
    },
    {
      Choice: {"A", 1},
      B: {Location: "1:5"},
    }
}
```

Or, we can use a more typical regex with

```
type AsOrBs struct {
    AsOrBs parse.Regex `regex:"[AB]*"`
}
```

## S-Expressions

Here's a simple way to parse s-expressions with reflect-peg.

```
type Whitespace struct {
    Whitespace parse.Regex `regex:"\s*"`
}
type Name struct {
    Whitespace Whitespace
    Name parse.Regex `regex:"[a-z]+"
}
type SExpression struct {
    Open parse.Open // An open parenthesis
    Function Name
    Arguments []Expression
    Close parse.Close // A close parenthesis - if unmatched, it will immediately error with associated open's location.
}
type Expression struct {
    parse.Choice
    Call SExpression
    Variable Name
}
```





