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
## Details

### Sequence

When parsing into a struct, the struct's fields will be parsed sequentially. If any of these fields fail to parse, then the struct also fails to parse. If all of the fields succeed parsing, then the struct succeeds parsing.

### Alternation

If the first field of a struct is the special type

```
type parser.Choice struct {
    Choice string
    Number int
}
```

then the struct will instead be parsed as an alternation. When parsing into the type, each field will be attempted in sequence. If any of these succeed, then parsing the struct succeeds. The `Choice` field will be updated to indicate which field was successful, and all other fields will have the zero value.

### Optionals

Parsing a `*T` is optional. It will attempt to parse a `T` instead. If successful, a pointer to the parsed value will be produced. Otherwise, the pointer will be `nil`, but parsing will succeed.

### Repetitions

Parsing a `[]T` will parse `T` as many times as possible. The result will be a slice of each result (in sequence) that parsed successfully. If none parsed successfully, it will be empty or `nil`.

### Forward Lookahead

Parsing a `<-chan T` will parse a `T`, then back up (without consuming any input). If a `T` fails to parse, then so will the `<-chan T`.

If successful, the resulting value will be a channel of capacity 1 containing the parsed value.


### Negative Lookahead

Parsing a `chan<- T` will try to parse a `T`. If this succeeds, then the lookahead fails. If it fails, then this succeeds.

If successful, the result is `nil`.

### Verification

Sometimes you want to do additional (programmatic) checking of a parsed structured after it's done (because the grammar is more permissive than semantically is reasonable).

If a value (or a pointer to it) implements `Verify() error`, it will be called after successful parsing. If it returns an error, then the value will be assumed to have failed will the corresponding error.

### Custom Parsing

Sometimes, you want to parse from configuration information, but you want to store an entirely different type in your tree. If your type implements a method:

```
func (target* Type) From(other OtherType) error {
}
```

or

```
func (target* Type) From(other OtherType) {
}
```

then the parser will instead attempt to parse an `OtherType` and if successful will use `From` to instantiate your value.

### Custom Parsing

:TODO DOCUMENTATION:


