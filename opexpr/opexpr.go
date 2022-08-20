package opexpr

import (
	"fmt"
	"strings"
)

// ExpressionKind represents the 'kind' of an expression, i.e. whether it is a
// value or operator of a certain arity and associativity.
type ExpressionKind int

const (
	isRightAssoc    ExpressionKind = 1
	hasLeftArg      ExpressionKind = 2
	hasRightArg     ExpressionKind = 4
	isParen         ExpressionKind = 8
	isCloseParen    ExpressionKind = 16
	isCloseAllParen ExpressionKind = 32
)

const (
	// BinaryLeftAssoc is the ExpressionKind for a left associative binary operator.
	BinaryLeftAssoc ExpressionKind = hasLeftArg | hasRightArg
	// BinaryRightAssoc is the ExpressionKind for a right associative binary operator.
	BinaryRightAssoc ExpressionKind = hasLeftArg | hasRightArg | isRightAssoc
	// Prefix is the ExpressionKind for a prefix operator.
	Prefix ExpressionKind = hasRightArg
	// Postfix is the ExpressionKind for a postfix operator.
	Postfix ExpressionKind = hasLeftArg
	// OpenParen is the ExpressionKind for an opening parenthesis
	OpenParen ExpressionKind = isParen
	// CloseParen is the ExpressionKind for a closing parenthesis
	CloseParen ExpressionKind = isParen | isCloseParen
	// CloseAllParens is the ExpressionKind for a closing parenthesis that closes all
	// currently open parentheses.
	CloseAllParens ExpressionKind = isParen | isCloseParen | isCloseAllParen
	// Parenthetical is the expression kind for a parenthetical operator (such as
	// C/Javscript's [...] indexation operator).
	Parenthetical ExpressionKind = isParen | hasLeftArg | hasRightArg
	// Value is the ExpressionKind for a value.
	Value ExpressionKind = 0
)

func (k ExpressionKind) String() string {
	switch k {
	case BinaryLeftAssoc:
		return "BinaryLeftAssoc"
	case BinaryRightAssoc:
		return "BinaryRightAssoc"
	case Prefix:
		return "Prefix"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case CloseAllParens:
		return "CloseAllParens"
	case Postfix:
		return "Postfix"
	case Value:
		return "Value"
	case Parenthetical:
		return "Parenthetical"
	default:
		panic(fmt.Sprintf("Unrecognized OperatorKind %v", int(k)))
	}
}

// Element should be implemented for pointers to the elements of the input slice
type Element interface {
	// ParenKind returns an integer representing the 'kind' of parenthesis. For
	// example, by assigning different kinds to '()' and '[]' you can require that
	// opening and closing parentheses are appropriately matched. Its value is
	// applicable only for OpenParen, CloseParen and CloseAllParens elements.
	ParenKind() int
	// ExpressionKind returns a value representing the kind of the expression (see
	// docs for ExpressionKind). The boolean argument is true when the expression
	// has a value to its left. This makes it possible to handle operators that
	// can be either prefix or postfix, or operators that can be either binary or
	// unary.
	ExpressionKind(hasExpressionToLeft bool) ExpressionKind
	// Precedence returns an integer representing the operator's precedence
	// relative to other operators. Lower values indicate higher precedence (i.e.
	// operators that bind tighter). The boolean argument is true when the
	// expression has a value to its left. This makes it possible to handle
	// operators that can be either prefix or postfix, or operators that can be
	// either binary or unary.
	Precedence(hasExpressionToLeft bool) int
}

type Stream[E any] interface {
	Next() (E, bool)
}

type ParseErrorKind int

const (
	// ParseErrorUnexpectedOperator occurs when an operator is found in a position
	// where it cannot be incorprated into a valid parse.
	ParseErrorUnexpectedOperator = iota
	// ParseErrorUnexpectedValue occurs when a value is found in a position where
	// it cannot be incorporated into a valid parse.
	ParseErrorUnexpectedValue
	// ParseErrorUnexpectedClosingParen occurs when a closing parenthesis is found
	// with no matching opening parenthesis.
	ParseErrorUnexpectedClosingParen
	// ParseErrorWrongKindOfClosingParena occurs when an opening parenthesis is
	// closed with a parenthesis of a different kind (e.g. '(' is closed with
	// ']')).
	ParseErrorWrongKindOfClosingParen
	// ParseErrorMissingClosingParen occurs when no closing parenthesis is found
	// for an opening parenthesis.
	ParseErrorMissingClosingParen
)

func (k ParseErrorKind) String() string {
	switch k {
	case ParseErrorUnexpectedOperator:
		return "ParseErrorUnexpectedOperator"
	case ParseErrorUnexpectedValue:
		return "ParseErrorUnexpectedValue"
	case ParseErrorUnexpectedClosingParen:
		return "ParseErrorUnexpectedClosingParen"
	case ParseErrorWrongKindOfClosingParen:
		return "ParseErrorWrongKindOfClosingParen"
	case ParseErrorMissingClosingParen:
		return "ParseErrorMissingClosingParen"
	default:
		panic("Unrecognized ParseErrorKind")
	}
}

// ParseError represents a parse error. Elem is either to a pointer to an
// element or nil if Kind is ParseErrorMissingClosingParen.
type ParseError[EP Element] struct {
	Kind ParseErrorKind
	Elem EP
}

func ShowParseError[EP Element](pe *ParseError[EP]) string {
	return fmt.Sprintf("%v@%v", pe.Kind, pe.Elem)
}

// TreeBuilder defines methods for building a tree out of input elements
type TreeBuilder[T any, E Element] interface {
	Element
	// Given that the receiver is an operator or an opening parenthesis, make a
	// corresponding parse tree node with the specified children. Either or both
	// of leftArg and rightArg may be nil depending on the arity and
	// prefix/postfix nature of the operator. If the receiver is an opening
	// parenthesis then leftArg is non-nil and rightArg is nil.
	MakeNode(leftArg, rightArg *T) *T
	// MakeErrorNode makes an error node given a parse error and left/right
	// children (either or both of which may be nil).
	MakeErrorNode(pe *ParseError[E], leftChild, rightChild *T) *T
}

// SimpleNode has a trivial implementation of the TreeBuilder interface that is
// useful for debugging and testing
type SimpleNode[T Element] struct {
	Left, Right *SimpleNode[T]
	Value       T
}

type nodePool[T any, E TreeBuilder[T, E]] struct {
	nodes      []node[T, E]
	parenRoots []**node[T, E]
	parenKinds []int
	stack      []*node[T, E]
}

// MakeNodePool makes a pool of shadow parse tree nodes for a given tree node
// type and element type.
func MakeNodePool[T any, E TreeBuilder[T, E]](capacity int) *nodePool[T, E] {
	return &nodePool[T, E]{make([]node[T, E], capacity), make([]**node[T, E], capacity/4), make([]int, capacity/4), make([]*node[T, E], capacity/4)}
}

type node[T any, EP Element] struct {
	elem        EP
	left, right *node[T, EP]
	treeNode    *T
	err         *ParseError[EP]

	// This field is used to memoize traversal of the tree when finding the
	// appropriate level to insert a right-associative operator.
	bottom *node[T, EP]
}

func zeroNode[T any, EP Element](n *node[T, EP]) {
	n.left = nil
	n.right = nil
	n.treeNode = nil
	n.err = nil
	n.bottom = nil
}

// ParseStream parses a stream of input elements that implement the
// ElementStream interface. It returns pointer to the root parse tree node, or
// nil if the input stream is empty. It should only be necessary to provide the
// first type parameter (the type of the nodes in the resulting parse tree). The
// existingNodePool argument should be the return value of MakeNodePool().
func ParseStream[T any, E TreeBuilder[T, E], S Stream[E]](stream S, pool *nodePool[T, E]) (*T, []*ParseError[E]) {
	return ParseStreamWithJuxtaposition(stream, nil, pool)
}

// ParseStreamWithJuxtaposition works like ParseStream except that it takes an
// additional operator element. This element is used to combine juxtaposed
// values.
func ParseStreamWithJuxtaposition[T any, E TreeBuilder[T, E], S Stream[E]](stream S, juxtapositionElement *E, pool *nodePool[T, E]) (*T, []*ParseError[E]) {
	const stackAllocSize = 128
	stackElems := make([]E, 0, stackAllocSize)
	elems := &stackElems
	for {
		e, ok := stream.Next()
		if !ok {
			break
		}
		if elems == &stackElems && len(*elems) == stackAllocSize {
			heapElems := make([]E, stackAllocSize, stackAllocSize*2)
			copy(heapElems, stackElems)
			elems = &heapElems
		}
		*elems = append(*elems, e)
	}

	return ParseSliceWithJuxtaposition(*elems, juxtapositionElement, pool)
}

// ParseSlice parses a slice of input elements that implement the Element
// interface. It returns pointer to the root parse tree node, or nil if the
// input slice is empty. It should only be necessary to provide the first type
// parameter (the type of the nodes in the resulting parse tree). The
// existingNodePool argument should be the return value of MakeNodePool().
func ParseSlice[T any, E TreeBuilder[T, E]](elements []E, pool *nodePool[T, E]) (*T, []*ParseError[E]) {
	return ParseSliceWithJuxtaposition(elements, nil, pool)
}

// ParseSliceWithJuxtaposition works like ParseSlice except that it takes a
// pointer to an additional operator element. If the pointer is non-nil, the
// element is used to combine juxtaposed values.
func ParseSliceWithJuxtaposition[T any, E TreeBuilder[T, E]](elements []E, juxtapositionElement *E, pool *nodePool[T, E]) (*T, []*ParseError[E]) {
	var errors []*ParseError[E]

	// Shortcut for the very common case of a single expression
	if len(elements) == 1 {
		e := elements[0]
		if e.ExpressionKind(false) == Value {
			return e.MakeNode(nil, nil), errors
		}
	}

	poolSize := len(elements)*2 + 1
	if len(pool.nodes) < poolSize {
		pool.nodes = append(pool.nodes, make([]node[T, E], poolSize-len(pool.nodes))...)
	}
	poolI := 0

	var root *node[T, E]
	hole := &root

	// for each paren level.
	pool.parenRoots = pool.parenRoots[:1] // reset to length 1 while leaving existing capacity
	pool.parenKinds = pool.parenKinds[:0] // reset to empty while leaving exisring capacity
	pool.parenRoots[0] = &root

	// used later to alloc an approtiately-sized stack for traversing the
	// temporary parse tree.
	depth := 1

	for i := range elements {
		e := elements[i]

		ekind := e.ExpressionKind(hole == nil)

		parenRootP := pool.parenRoots[len(pool.parenRoots)-1]

		if ekind&isCloseParen != 0 {
			// closing parens
			if len(pool.parenRoots) <= 1 {
				pe := &ParseError[E]{ParseErrorUnexpectedClosingParen, e}
				errors = append(errors, pe)
				node := &pool.nodes[poolI]
				poolI++
				zeroNode(node)
				node.left = root
				node.err = pe
				root = node
			} else if e.ParenKind() != pool.parenKinds[len(pool.parenKinds)-1] {
				pe := &ParseError[E]{ParseErrorWrongKindOfClosingParen, e}
				errors = append(errors, pe)
				node := &pool.nodes[poolI]
				poolI++
				zeroNode(node)
				rt := pool.parenRoots[len(pool.parenRoots)-1]
				node.left = *rt
				node.err = pe
				*rt = node
			} else if ekind&isCloseAllParen != 0 {
				pool.parenRoots = pool.parenRoots[0:1]
				pool.parenKinds = pool.parenKinds[0:0]
			} else {
				pool.parenRoots = pool.parenRoots[:len(pool.parenRoots)-1]
				pool.parenKinds = pool.parenKinds[:len(pool.parenKinds)-1]
			}
		} else if ekind&hasLeftArg != 0 {
			// postfix op or bin op

			if hole != nil {
				pe := &ParseError[E]{ParseErrorUnexpectedOperator, e}
				errors = append(errors, pe)
				errorNode := &pool.nodes[poolI]
				poolI++
				zeroNode(errorNode)
				errorNode.elem = e
				errorNode.err = pe
				*hole = errorNode
				hole = nil
				depth++
			}

			n := findOpLevel(e, parenRootP, hole)

			opNode := &pool.nodes[poolI]
			poolI++
			zeroNode(opNode)
			opNode.elem = e
			opNode.left = *n

			if ekind&hasRightArg != 0 {
				// it's a bin op
				// is it also an opening paren?
				if ekind&isParen != 0 && ekind&isCloseParen == 0 {
					depth++
					pool.parenRoots = append(pool.parenRoots, &opNode.right)
					pool.parenKinds = append(pool.parenKinds, e.ParenKind())
				}
				hole = &opNode.right
			}

			*n = opNode
			depth++
		} else if ekind&hasRightArg != 0 && ekind&hasLeftArg == 0 {
			// prefix op

			if hole == nil {
				if juxtapositionElement == nil {
					pe := &ParseError[E]{ParseErrorUnexpectedOperator, e}
					errors = append(errors, pe)
					errorNode := &pool.nodes[poolI]
					poolI++
					zeroNode(errorNode)
					errorNode.elem = e
					errorNode.left = *parenRootP
					errorNode.err = pe
					hole = &errorNode.right
					*parenRootP = errorNode
				} else {
					n := findOpLevel(*juxtapositionElement, parenRootP, hole)
					opNode := &pool.nodes[poolI]
					poolI++
					zeroNode(opNode)
					opNode.elem = *juxtapositionElement
					opNode.left = *n
					hole = &opNode.right
					*n = opNode
				}
				depth++
			}

			opNode := &pool.nodes[poolI]
			poolI++
			zeroNode(opNode)
			opNode.elem = e
			*hole = opNode
			hole = &opNode.right
			depth++
		} else {
			// value

			valueNode := &pool.nodes[poolI]
			poolI++
			zeroNode(valueNode)
			valueNode.elem = e

			if hole == nil {
				if juxtapositionElement == nil {
					pe := &ParseError[E]{ParseErrorUnexpectedValue, e}
					errors = append(errors, pe)
					errorNode := &pool.nodes[poolI]
					poolI++
					zeroNode(errorNode)
					errorNode.elem = e
					errorNode.left = *parenRootP
					errorNode.err = pe
					hole = &errorNode.right
					*parenRootP = errorNode
				} else {
					n := findOpLevel(*juxtapositionElement, parenRootP, hole)
					opNode := &pool.nodes[poolI]
					poolI++
					zeroNode(opNode)
					opNode.elem = *juxtapositionElement
					opNode.left = *n
					hole = &opNode.right
					*n = opNode
				}
				depth++
			}

			*hole = valueNode

			// If it's an opening paren, add a level.
			if ekind == OpenParen {
				depth++
				pool.parenRoots = append(pool.parenRoots, &valueNode.right)
				pool.parenKinds = append(pool.parenKinds, e.ParenKind())
				hole = &valueNode.right
			} else {
				hole = nil
			}
		}
	}

	if hole != nil && len(elements) > 0 {
		pe := &ParseError[E]{ParseErrorUnexpectedOperator, elements[len(elements)-1]}
		errors = append(errors, pe)
		errorNode := &pool.nodes[poolI]
		poolI++
		zeroNode(errorNode)
		errorNode.elem = elements[len(elements)-1]
		errorNode.err = pe
		*hole = errorNode
		hole = nil
	}

	// Wrap with an error node if there are missing closing parens and we don't
	// already have a 'wrong kind' error.
	if len(pool.parenRoots) > 1 {
		last := pool.parenRoots[len(pool.parenRoots)-1]
		if !(last != nil && (*last).err != nil && (*last).err.Kind == ParseErrorWrongKindOfClosingParen) {
			pe := &ParseError[E]{ParseErrorMissingClosingParen, elements[len(elements)-1]}
			errors = append(errors, pe)
			node := &pool.nodes[poolI]
			// no need to increment poolI as we won't be using the pool again
			zeroNode(node)
			node.left = root
			node.err = pe
			root = node
		}
	}

	rr, errs := buildTree(root, elements, juxtapositionElement, depth, pool), errors

	return rr, errs
}

func findOpLevel[T any, E TreeBuilder[T, E]](e E, root **node[T, E], hole **node[T, E]) **node[T, E] {
	exprToLeft := hole == nil
	precedenceCmp := e.Precedence(exprToLeft)
	n := root

	if e.ExpressionKind(exprToLeft)&isRightAssoc == 0 {
		precedenceCmp++
	}

	prec := (*n).elem.Precedence(exprToLeft)
	for {
		if *n == nil || (*n).err != nil {
			return n
		}
		ek := (*n).elem.ExpressionKind(exprToLeft)
		if ek&hasRightArg == 0 {
			break
		}
		// don't descend into foo[bar]
		// not necessary to check ek&isCloseParen == 0 because ')' wouldnt have ek&(hasLeftArg|hasRightArg)!= 0
		if ek&isParen != 0 && ek&(hasLeftArg|hasRightArg) != 0 {
			break
		}
		if prec < precedenceCmp {
			break
		}

		if (*n).bottom != nil {
			n = &((*n).bottom.right)
			prec = (*n).elem.Precedence(exprToLeft)
		} else {
			oldn := *n
			n = &((*n).right)
			newprec := (*n).elem.Precedence(exprToLeft)
			if prec == newprec {
				(*root).bottom = oldn
			}
			prec = newprec
		}
	}

	return n
}

func buildTree[T any, E TreeBuilder[T, E]](root *node[T, E], elements []E, juxtapositionElement *E, stackDepth int, pool *nodePool[T, E]) *T {
	if root == nil {
		return nil
	}

	current := root
	if len(pool.stack) < stackDepth {
		pool.stack = append(pool.stack, make([]*node[T, E], stackDepth-len(pool.stack))...)
	}
	pool.stack = pool.stack[:0] // empty slice while leaving capacity

	for {
		for {
			if current.left != nil && current.left.treeNode == nil {
				pool.stack = append(pool.stack, current)
				current = current.left
			} else if current.right != nil && current.right.treeNode == nil {
				pool.stack = append(pool.stack, current)
				current = current.right
			} else {
				break
			}
		}

		ce := current.elem
		if current.err == nil {
			current.treeNode = ce.MakeNode(leftTreeNodeOf(current), rightTreeNodeOf(current))
		} else {
			current.treeNode = ce.MakeErrorNode(current.err, leftTreeNodeOf(current), rightTreeNodeOf(current))
			current.err = nil
		}

		if len(pool.stack) == 0 {
			break
		}

		current, pool.stack = pool.stack[len(pool.stack)-1], pool.stack[:len(pool.stack)-1]
	}

	return root.treeNode
}

func leftTreeNodeOf[T any, E Element](n *node[T, E]) *T {
	if n.left != nil {
		return n.left.treeNode
	}
	return nil
}

func rightTreeNodeOf[T any, E Element](n *node[T, E]) *T {
	if n.right != nil {
		return n.right.treeNode
	}
	return nil
}

// ShowSimpleNode shows the expression rooted at its argument using '⎡' and '⎦'
// to delimit parse tree nodes.
func ShowSimpleNode[E Element](n *SimpleNode[E]) string {
	var o strings.Builder
	showSimpleNodeHelper(&o, n)
	return o.String()
}

func showSimpleNodeHelper[E Element](o *strings.Builder, n *SimpleNode[E]) string {
	const oParens = "([{"
	const cParens = ")]}"
	const weirdOpenBracket = '⎡'
	const weirdCloseBracket = '⎦'

	if n == nil {
		return "@NIL@"
	}

	e := n.Value

	if e.ExpressionKind(true) == OpenParen {
		o.WriteByte(oParens[e.ParenKind()%len(oParens)])
		showSimpleNodeHelper(o, n.Right)
		o.WriteByte(cParens[e.ParenKind()%len(oParens)])
	} else if n.Left == nil && n.Right == nil {
		o.WriteString(fmt.Sprintf("%v", e))
	} else if n.Left != nil && n.Right != nil {
		o.WriteRune(weirdOpenBracket)
		showSimpleNodeHelper(o, n.Left)
		o.WriteRune(' ')
		o.WriteString(fmt.Sprintf("%v", e))
		o.WriteRune(' ')
		showSimpleNodeHelper(o, n.Right)
		o.WriteRune(weirdCloseBracket)
	} else if n.Left != nil && n.Right == nil {
		o.WriteRune(weirdOpenBracket)
		showSimpleNodeHelper(o, n.Left)
		o.WriteString(fmt.Sprintf("%v", e))
		o.WriteRune(weirdCloseBracket)
	} else if n.Right != nil && n.Left == nil {
		o.WriteRune(weirdOpenBracket)
		o.WriteString(fmt.Sprintf("%v", e))
		showSimpleNodeHelper(o, n.Right)
		o.WriteRune(weirdCloseBracket)
	} else {
		panic("Internal error 4 in 'printHelper'")
	}

	return o.String()
}
