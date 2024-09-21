package opexpr

import (
	"fmt"
	"iter"
	"strings"
)

// StringElement provides a toy implementation of Element useful for writing
// tests.
type StringElement string

func (se StringElement) String() string {
	return string(se)
}

// MakeStringElements constructs a slice of StringElements by splitting the
// input string on the space character.
func MakeStringElements(s string) []StringElement {
	if s == "" {
		return []StringElement{}
	}

	ses := make([]StringElement, 0)
	for _, s := range strings.Split(s, " ") {
		ses = append(ses, StringElement(s))
	}
	return ses
}

// MakeStringSeq constructs an sequence of StringElements by splitting the
// input string on the space character.
func MakeStringSeq(s string) iter.Seq[StringElement] {
	return func(yield func(StringElement) bool) {
		for _, s := range strings.Split(s, " ") {
			if !yield(StringElement(s)) {
				break
			}
		}
	}
}

func isParenSymbol(s string) bool {
	if s == ")$" || s == "]$" || s == "}$" {
		return true
	}
	if s[0] != '(' && s[0] != ')' && s[0] != '[' && s[0] != ']' && s[0] != '{' && s[0] != '}' {
		return false
	}
	if len(s) > 1 {
		i := len(s) - 1
		if s[i] != '(' && s[i] != ')' && s[i] != '[' && s[i] != ']' && s[i] != '{' && s[i] != '}' {
			return false
		}
	}
	return true
}

func (s StringElement) ParenKind() int {
	if isParenSymbol(string(s)) {
		if s[0] == '(' || s[0] == ')' {
			return 0
		}
		if s[0] == '[' || s[0] == ']' {
			return 1
		}
		return 2
	}
	return -1
}

func (s StringElement) ExpressionKind(hasExpressionToLeft bool) ExpressionKind {
	if s == "(" || s == "[" || s == "{" {
		return OpenParen
	}
	if s == ")" || s == "]" || s == "}" {
		return CloseParen
	}
	if s == ")$" || s == "]$" || s == "}$" {
		return CloseAllParens
	}

	if isParenSymbol(string(s)) {
		return BinaryLeftAssoc | isParen
	}

	switch s[len(s)-1] {
	case '{':
		return BinaryRightAssoc
	case '\'':
		if hasExpressionToLeft {
			return BinaryLeftAssoc
		}
		return Prefix
	case '[':
		if hasExpressionToLeft {
			return BinaryRightAssoc
		}
		return Prefix
	default:
		switch s[0] {
		case ':', '+', '-', '*', '/', '<', '>':
			return BinaryLeftAssoc
		case '#', '$':
			return Postfix
		case '!', '&':
			return Prefix
		default:
			return Value
		}
	}
}

func (s StringElement) Precedence(_hasExpressionToLeft bool) int {
	if s[len(s)-1] == '{' || s[len(s)-1] == '[' || s[len(s)-1] == '\'' {
		return len(s) - 1
	}
	return len(s)
}

func (elem StringElement) MakeNode(arg1, arg2 *SimpleNode[StringElement]) *SimpleNode[StringElement] {
	return &SimpleNode[StringElement]{arg1, arg2, elem}
}

func (elem StringElement) MakeErrorNode(e *ParseError[StringElement], arg1, arg2 *SimpleNode[StringElement]) *SimpleNode[StringElement] {
	se := StringElement(fmt.Sprintf("@error:%v", e))
	return &SimpleNode[StringElement]{arg1, arg2, se}
}

func (pe ParseError[T]) String() string {
	return ShowParseError(&pe)
}
