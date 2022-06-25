# Deckwreck

Deckwreck is a library that helps with the difficult parts of writing a
recursive descent parser. At present it provides

* fast maps from keyword strings to integers; and
* a precedence-based expression parser with O(n) performance.

## Current status

The keyword map code is stable and well tested. The expression parsing code
is more experimental at present.

## Keyword testing

Many parsers need to check identifiers against a small list of reserved
keywords. The Go parser, for example, [uses a map to perform this
check](https://github.com/golang/go/blob/527ace0ffa81d59698d3a78ac3545de7295ea76b/src/go/token/token.go#L282).

Deckwreck provides fast trie-based maps from strings to integers. These are
around 1.5-2 times faster than a hash map for typical use cases.

## Expression parsing

Parsing expressions with operators of different precedence levels is one of the
more challenging aspects of writing a recursive descent parser. Deckwreck's
expression parser uses an iterative algorithm and does not recurse. The
expression parser constructs a shadow parse tree from a reusable pool of nodes.
The shadow parse tree is then walked to construct the real parse tree. Parse
tree construction is controlled by user-supplied interface implementations.

Notable features of Deckwreck's expression parser:

* O(n) in the number of expressions and operators, as confirmed by benchmarks.
* Fast for small expressions (the most common case in typical source code).
* Works with existing parse tree data structures via a generic interface. 
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
