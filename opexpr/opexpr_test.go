package opexpr

import (
	"math/rand"
	"strings"
	"testing"
)

const (
	WithJuxtaposition = iota
	WithoutJuxtaposition
	Both
)

const (
	JuxLeftAssoc = iota
	JuxRightAssoc
)

func testParse(t *testing.T, jux, juxAssoc, nErrors int, input, output string) {
	t.Logf("Input: %v\n", input)

	pool := MakeNodePool[SimpleNode[StringElement], StringElement](32)

	if jux == WithJuxtaposition || jux == Both {
		var jux StringElement
		if juxAssoc == JuxRightAssoc {
			jux = StringElement("/{")
		} else {
			jux = StringElement("/")
		}
		wJuxRoot, wJuxErrs := ParseSliceWithJuxtaposition(MakeStringElements(input), &jux, pool)
		wJuxOutput := ShowSimpleNode(wJuxRoot)

		if len(wJuxErrs) != nErrors {
			t.Errorf("Expected %v (with jux) errors, got %v\n", nErrors, len(wJuxErrs))
		}
		if wJuxOutput != output {
			t.Errorf("Expected output (with jux): %v\nGot: %v\n", output, wJuxOutput)
		}
	}
	if jux == WithoutJuxtaposition || jux == Both {
		woJuxRoot, woJuxErrs := ParseSlice(MakeStringElements(input), pool)
		woJuxOutput := ShowSimpleNode(woJuxRoot)

		if len(woJuxErrs) != nErrors {
			t.Errorf("Expected %v (without jux) errors, got %v\n", nErrors, len(woJuxErrs))
		}
		if woJuxOutput != output {
			t.Errorf("Expected output (without jux): %v\nGot: %v\n", output, woJuxOutput)
		}
	}
}

func TestSimpleLeftAssocExpression(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "1 + 2 + 3", "⎡⎡1 + 2⎦ + 3⎦")
}

func TestSimpleRightAssocExpression(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "1 +{ 2 +{ 3", "⎡1 +{ ⎡2 +{ 3⎦⎦")
}

func TestSimpleExpressionWithUnaryPrefix(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "+' 1 + 2", "⎡⎡+'1⎦ + 2⎦")
}

func TestEmptyExpression(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "", "")
}

func TestSimpleExpression(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "1", "1")
}

func TestSimpleExpressionWithLowPrecUnaryPrefix(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "++' 1 + 2", "⎡++'⎡1 + 2⎦⎦")
}

func TestSimpleJux(t *testing.T) {
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "1 2", "⎡1 / 2⎦")
}

func TestMultiJux(t *testing.T) {
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "1 2 3 4 5 6 7", "⎡⎡⎡⎡⎡⎡1 / 2⎦ / 3⎦ / 4⎦ / 5⎦ / 6⎦ / 7⎦")
}

func TestMultiRightJux(t *testing.T) {
	testParse(t, WithJuxtaposition, JuxRightAssoc, 0, "1 2 3 4 5 6 7", "⎡1 /{ ⎡2 /{ ⎡3 /{ ⎡4 /{ ⎡5 /{ ⎡6 /{ 7⎦⎦⎦⎦⎦⎦")
}

func TestJuxWithPrefixOps(t *testing.T) {
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "! 1 ! 2", "⎡⎡!1⎦ / ⎡!2⎦⎦")
}

func TestJuxWithLowPrecPrefixOp(t *testing.T) {
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "!! 1 ! 2", "⎡!!⎡1 / ⎡!2⎦⎦⎦")
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "! 1 !! 2", "⎡⎡!1⎦ / ⎡!!2⎦⎦")
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "! 1 ! !! 2", "⎡⎡!1⎦ / ⎡!⎡!!2⎦⎦⎦")
	testParse(t, WithJuxtaposition, JuxLeftAssoc, 0, "! 1 ! ! 2", "⎡⎡!1⎦ / ⎡!⎡!2⎦⎦⎦")
}

func TestJuxWithJuxDisabled(t *testing.T) {
	testParse(t, WithoutJuxtaposition, JuxLeftAssoc, 1, "1 err", "⎡1 @error:ParseErrorUnexpectedValue@err err⎦")
	testParse(t, WithoutJuxtaposition, JuxLeftAssoc, 1, "! 1 ! 2", "⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!2⎦⎦")
	testParse(t, WithoutJuxtaposition, JuxLeftAssoc, 1, "!! 1 ! 2", "⎡⎡!!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!2⎦⎦")
	testParse(t, WithoutJuxtaposition, JuxLeftAssoc, 1, "! 1 !! 2", "⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@!! ⎡!!2⎦⎦")
	testParse(t, WithoutJuxtaposition, JuxLeftAssoc, 1, "! 1 ! !! 2", "⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!⎡!!2⎦⎦⎦")
	testParse(t, WithoutJuxtaposition, JuxLeftAssoc, 1, "! 1 ! ! 2", "⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!⎡!2⎦⎦⎦")
}

func TestComplexExpressions(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "! 1 * ( 2 + 3 ) ++ 4 ::{ 9 ::{ 10 ::{ nil", "⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡4 ::{ ⎡9 ::{ ⎡10 ::{ nil⎦⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "! 1 * ( 2 + 3 ) ++ ( ( ( 4 ::{ 9 ) ) ) ::{ 10 ::{ nil", "⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡(((⎡4 ::{ 9⎦))) ::{ ⎡10 ::{ nil⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +{ 2 ++{ 3 ++{ 4 ++{ 5", "⎡⎡1 +{ 2⎦ ++{ ⎡3 ++{ ⎡4 ++{ 5⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +{ 2 ++{ 3 ++{ 4 ++{ 5 +++{ 9", "⎡⎡⎡1 +{ 2⎦ ++{ ⎡3 ++{ ⎡4 ++{ 5⎦⎦⎦ +++{ 9⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +{ 2 ++{ 3 --{ 4 **{ 5 +++{ 9", "⎡⎡⎡1 +{ 2⎦ ++{ ⎡3 --{ ⎡4 **{ 5⎦⎦⎦ +++{ 9⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 + 2 ++ 2 + 3", "⎡⎡1 + 2⎦ ++ ⎡2 + 3⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 ++[ 2 ++[ 3 ++[ 4 * 5", "⎡1 ++[ ⎡2 ++[ ⎡3 ++[ ⎡4 * 5⎦⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 ++ 2 ++ 3 ++ 4 * 5", "⎡⎡⎡1 ++ 2⎦ ++ 3⎦ ++ ⎡4 * 5⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 ++ 2 -- 3 ++ 4 * 5", "⎡⎡⎡1 ++ 2⎦ -- 3⎦ ++ ⎡4 * 5⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +++{ 2 +++{ 3 ++{ 4 ++{ 5 + 6 + 7 + 8", "⎡1 +++{ ⎡2 +++{ ⎡3 ++{ ⎡4 ++{ ⎡⎡⎡5 + 6⎦ + 7⎦ + 8⎦⎦⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 1, "! 1 * ( 2 + 3 ) ++ * 4 ::{ 9 ::{ 10 ::{ nil", "⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡⎡@error:ParseErrorUnexpectedOperator@* * 4⎦ ::{ ⎡9 ::{ ⎡10 ::{ nil⎦⎦⎦⎦")
}

func TestSimpleParens(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "( 1 + 2 )", "(⎡1 + 2⎦)")
	testParse(t, Both, JuxLeftAssoc, 0, "( [ 1 + 2 ] )", "([⎡1 + 2⎦])")
	testParse(t, Both, JuxLeftAssoc, 1, "( ( 1 + 2 )", "⎡((⎡1 + 2⎦))@error:ParseErrorMissingClosingParen@)⎦")
	testParse(t, Both, JuxLeftAssoc, 1, "( 1 + 2", "⎡(⎡1 + 2⎦)@error:ParseErrorMissingClosingParen@2⎦")
	testParse(t, Both, JuxLeftAssoc, 1, "( (1 + 2", "⎡(⎡(1 + 2⎦)@error:ParseErrorMissingClosingParen@2⎦")
	testParse(t, Both, JuxLeftAssoc, 1, "( 1 + 2 ]", "(⎡⎡1 + 2⎦@error:ParseErrorWrongKindOfClosingParen@]⎦)")
}

func TestParentheticalOp(t *testing.T) {
	testParse(t, Both, JuxLeftAssoc, 0, "1 [[ 2 + 3 ]", "⎡1 [[ ⎡2 + 3⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 [[ 2 + 3 ] [[ 4 + 5 ]", "⎡⎡1 [[ ⎡2 + 3⎦⎦ [[ ⎡4 + 5⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 [[ 2 + 3 ] [[ 4 ]", "⎡⎡1 [[ ⎡2 + 3⎦⎦ [[ 4⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +{ ( 2 + 3 ) +{ ( 4 )", "⎡1 +{ ⎡(⎡2 + 3⎦) +{ (4)⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +{ ( 2 + 3 ) +{ ( 4 )", "⎡1 +{ ⎡(⎡2 + 3⎦) +{ (4)⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 + 2 [[ 2 + 3 ]", "⎡⎡1 + 2⎦ [[ ⎡2 + 3⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 ++ 2 [[ 2 + 3 ]", "⎡1 ++ ⎡2 [[ ⎡2 + 3⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 + 2 [[[ 2 + 3 ]", "⎡⎡1 + 2⎦ [[[ ⎡2 + 3⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 +++ 2 [[ 2 + 3 ]", "⎡1 +++ ⎡2 [[ ⎡2 + 3⎦⎦⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 [[ 2 ] [[ 3 ]", "⎡⎡1 [[ 2⎦ [[ 3⎦")
	testParse(t, Both, JuxLeftAssoc, 0, "1 [[[ 2 + 3 ] + 4", "⎡⎡1 [[[ ⎡2 + 3⎦⎦ + 4⎦")
}

func benchmarkRightAssoc(b *testing.B, nArgs int) {
	var sb strings.Builder
	sb.WriteString("1")
	for i := 0; i < nArgs; i++ {
		sb.WriteString(" +{ 1")
	}
	input := sb.String()
	elems := MakeStringElements(input)

	b.ResetTimer()

	pool := MakeNodePool[SimpleNode[StringElement], StringElement](64)

	for i := 0; i < b.N; i++ {
		r, errs := ParseSlice(elems, pool)
		if len(errs) > 0 {
			b.Errorf("Not expecting to get any errors: %+v\n", r)
		}
	}
}

func BenchmarkRightAssoc0(b *testing.B) {
	benchmarkRightAssoc(b, 0)
}
func BenchmarkRightAssoc1(b *testing.B) {
	benchmarkRightAssoc(b, 1)
}
func BenchmarkRightAssoc20(b *testing.B) {
	benchmarkRightAssoc(b, 20)
}
func BenchmarkRightAssoc40(b *testing.B) {
	benchmarkRightAssoc(b, 40)
}
func BenchmarkRightAssoc60(b *testing.B) {
	benchmarkRightAssoc(b, 60)
}
func BenchmarkRightAssoc80(b *testing.B) {
	benchmarkRightAssoc(b, 80)
}
func BenchmarkRightAssoc100(b *testing.B) {
	benchmarkRightAssoc(b, 100)
}

func BenchmarkRightAssoc120(b *testing.B) {
	benchmarkRightAssoc(b, 120)
}

func BenchmarkRightAssoc140(b *testing.B) {
	benchmarkRightAssoc(b, 140)
}

func BenchmarkRightAssoc160(b *testing.B) {
	benchmarkRightAssoc(b, 160)
}

func BenchmarkRightAssoc180(b *testing.B) {
	benchmarkRightAssoc(b, 180)
}

func BenchmarkRightAssoc200(b *testing.B) {
	benchmarkRightAssoc(b, 200)
}

func bencharkLeftAssocInsideRightAssoc(b *testing.B, nArgs int) {
	var sb strings.Builder
	sb.WriteString("1")
	for i := 0; i < nArgs; i++ {
		sb.WriteString(" +++{ 1")
	}
	for i := 0; i < nArgs; i++ {
		sb.WriteString(" **{ 1")
	}
	// If we have to descend down the entire right branching structure
	// node-by-node for each of these higher-precedence left associative
	// operators, then we won't get O(n) performance.
	for i := 0; i < nArgs; i++ {
		sb.WriteString(" * 1")
	}
	input := sb.String()
	elems := MakeStringElements(input)

	b.ResetTimer()

	pool := MakeNodePool[SimpleNode[StringElement], StringElement](64)

	for i := 0; i < b.N; i++ {
		r, errs := ParseSlice(elems, pool)

		if len(errs) > 0 {
			b.Errorf("Not expecting to get any errors: %+v\n", r)
		}
	}
}

func BenchmarkLeftAssocInsideRightAssoc0(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 0)
}

func BenchmarkLeftAssocInsideRightAssoc1(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 1)
}

func BenchmarkLeftAssocInsideRightAssoc20(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 20)
}

func BenchmarkLeftAssocInsideRightAssoc40(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 40)
}

func BenchmarkLeftAssocInsideRightAssoc60(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 60)
}

func BenchmarkLeftAssocInsideRightAssoc80(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 80)
}

func BenchmarkLeftAssocInsideRightAssoc100(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 100)
}

func BenchmarkLeftAssocInsideRightAssoc120(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 120)
}

func BenchmarkLeftAssocInsideRightAssoc140(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 140)
}

func BenchmarkLeftAssocInsideRightAssoc160(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 160)
}

func BenchmarkLeftAssocInsideRightAssoc180(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 180)
}

func BenchmarkLeftAssocInsideRightAssoc200(b *testing.B) {
	bencharkLeftAssocInsideRightAssoc(b, 200)
}

// TestFuzz generates a bunch of random nonsense inputs and checks that none of
// them give rise to runtime errors.
func TestFuzz(t *testing.T) {
	source := rand.NewSource(12345)
	rand := rand.New(source)

	pool := MakeNodePool[SimpleNode[StringElement], StringElement](64)
	for i := 0; i < 10000; i++ {
		seq := randomStringElementSequence(rand)
		ParseSliceWithJuxtaposition(seq, nil, pool)
	}
}

func randomStringElementSequence(r *rand.Rand) []StringElement {
	l := r.Intn(20)
	elems := make([]StringElement, 0)
	for i := 0; i < l; i++ {
		elems = append(elems, randomStringElement(r))
	}
	return elems
}

func randomStringElement(r *rand.Rand) StringElement {
	chars := "+-/*{}[]()!&#$1234567890"
	l := r.Intn(9) + 1
	var sb strings.Builder
	for i := 0; i < l; i++ {
		c := chars[r.Intn(len(chars))]
		sb.WriteByte(c)
	}
	return StringElement(sb.String())
}
