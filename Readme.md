# Deckwreck

Deckwreck is a library that helps with the difficult parts of writing a
recursive descent parser. At present it provides the following:

* Fast maps from strings to integers that are useful for mapping strings to
  keywords
* A precedence-based expression parser with O(n) performance.

## Current status

The keyword testing code is stable and well tested. The expression parsing code
is more experimental at present.

## Keyword testing

Many parsers need to check identifiers against a small list of reserved
keywords. The Go parser, for example, [uses a map to perform this
check](https://github.com/golang/go/blob/527ace0ffa81d59698d3a78ac3545de7295ea76b/src/go/token/token.go#L282).

Deckwreck provides an fast implementation of maps from strings to integers that
can perform these checks around 1.5-2 times faster than a hash map.

## Expression parsing

Parsing expressions with operators of different precedence levels is one of the
more challenging aspects of writing a recursive descent parser. Deckwreck's
expression parser uses an iterative algorithm and does not recurse. It
constructs a shadow parse tree from a reusable pool of nodes (with the effect
that the parser itself does not allocate on most invocations). The shadow parse
tree is then walked to construct the real parse tree via user-supplied interface
implementations.

Notable features of Deckwreck's expression parser:

* O(n) in the number of expressions and operators, as confirmed by benchmarks.
* Fast for small expressions (usually the most common case).
* Works with your existing parse tree data structures via a generic interface. 
* Handles unary and binary operators with a unified set of precedence levels (a
  unary operator can have lower precedence than certain binary operators, or
  vice versa).
* Handles optionally binary/unary operators (e.g. `-` in C and Javascript).
* Handles operators that can either be prefix or postfix (e.g. `++` and `--` in
  C and Javascript).
* Handles expression combination via juxtaposition (as in e.g. ML, Haskell,
  Elm). Juxtaposition can be given a precedence level just like any other
  operator.
* Handles parenthetical operators (e.g. the `[...]` indexation operator in C and
  Javascript).
* Handles matching of parentheses of different types (e.g. `[]` and `()`).
* Operators can be defined as either left or right associative.
* Generates a partial parse tree for erroneous inputs.
