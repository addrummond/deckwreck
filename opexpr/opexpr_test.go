package opexpr

import (
	"math/rand"
	"strings"
	"testing"
)

func TestFoo(t *testing.T) {
	const (
		WithJuxtaposition = iota
		WithoutJuxtaposition
		Both
	)

	const (
		JuxLeftAssoc = iota
		JuxRightAssoc
	)

	type tst struct {
		jux           int
		juxAssoc      int
		input, output string
		nErrors       int
	}

	tsts := []tst{
		{
			Both,
			JuxLeftAssoc,
			"1 + 2 + 3",
			"⎡⎡1 + 2⎦ + 3⎦",
			0,
		},
		{
			Both,
			JuxLeftAssoc,
			"1 +{ 2 +{ 3",
			"⎡1 +{ ⎡2 +{ 3⎦⎦",
			0,
		},
		{
			Both,
			JuxLeftAssoc,
			"+' 1 + 2",
			"⎡⎡+'1⎦ + 2⎦",
			0,
		},
		{
			Both,
			JuxLeftAssoc,
			"",
			"",
			0,
		},
		{
			Both,
			JuxLeftAssoc,
			"++' 1 + 2",
			"⎡++'⎡1 + 2⎦⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"1 2",
			"⎡1 / 2⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"1 2 3 4 5 6 7",
			"⎡⎡⎡⎡⎡⎡1 / 2⎦ / 3⎦ / 4⎦ / 5⎦ / 6⎦ / 7⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxRightAssoc,
			"1 2 3 4 5 6 7",
			"⎡1 /{ ⎡2 /{ ⎡3 /{ ⎡4 /{ ⎡5 /{ ⎡6 /{ 7⎦⎦⎦⎦⎦⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"! 1 ! 2",
			"⎡⎡!1⎦ / ⎡!2⎦⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"!! 1 ! 2",
			"⎡!!⎡1 / ⎡!2⎦⎦⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"! 1 !! 2",
			"⎡⎡!1⎦ / ⎡!!2⎦⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"! 1 ! !! 2",
			"⎡⎡!1⎦ / ⎡!⎡!!2⎦⎦⎦",
			0,
		},
		{
			WithJuxtaposition,
			JuxLeftAssoc,
			"! 1 ! ! 2",
			"⎡⎡!1⎦ / ⎡!⎡!2⎦⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 err",
			"⎡1 @error:ParseErrorUnexpectedValue@err err⎦",
			1,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"! 1 ! 2",
			"⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!2⎦⎦",
			1,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"!! 1 ! 2",
			"⎡⎡!!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!2⎦⎦",
			1,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"! 1 !! 2",
			"⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@!! ⎡!!2⎦⎦",
			1,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"! 1 ! !! 2",
			"⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!⎡!!2⎦⎦⎦",
			1,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"! 1 ! ! 2",
			"⎡⎡!1⎦ @error:ParseErrorUnexpectedOperator@! ⎡!⎡!2⎦⎦⎦",
			1,
		},
		{
			JuxLeftAssoc,
			Both,
			"! 1 * ( 2 + 3 ) ++ 4 ::{ 9 ::{ 10 ::{ nil",
			"⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡4 ::{ ⎡9 ::{ ⎡10 ::{ nil⎦⎦⎦⎦",
			0,
		},
		{
			JuxLeftAssoc,
			Both,
			"! 1 * ( 2 + 3 ) ++ ( ( ( 4 ::{ 9 ) ) ) ::{ 10 ::{ nil",
			"⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡(((⎡4 ::{ 9⎦))) ::{ ⎡10 ::{ nil⎦⎦⎦",
			0,
		},
		{
			JuxLeftAssoc,
			Both,
			"! 1 * ( 2 + 3 ) ++ * 4 ::{ 9 ::{ 10 ::{ nil",
			"⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡⎡@error:ParseErrorUnexpectedOperator@* * 4⎦ ::{ ⎡9 ::{ ⎡10 ::{ nil⎦⎦⎦⎦",
			1,
		},
		{
			JuxLeftAssoc,
			Both,
			"( 1 + 2 )",
			"(⎡1 + 2⎦)",
			0,
		},
		{
			JuxLeftAssoc,
			Both,
			"( [ 1 + 2 ] )",
			"([⎡1 + 2⎦])",
			0,
		},
		{
			JuxLeftAssoc,
			Both,
			"( ( 1 + 2 )",
			"⎡((⎡1 + 2⎦))@error:ParseErrorMissingClosingParen@)⎦",
			1,
		},
		{
			JuxLeftAssoc,
			Both,
			"( 1 + 2",
			"⎡(⎡1 + 2⎦)@error:ParseErrorMissingClosingParen@2⎦",
			1,
		},
		{
			JuxLeftAssoc,
			Both,
			"( (1 + 2",
			"⎡(⎡(1 + 2⎦)@error:ParseErrorMissingClosingParen@2⎦",
			1,
		},
		{
			JuxLeftAssoc,
			Both,
			"( 1 + 2 ]",
			"(⎡⎡1 + 2⎦@error:ParseErrorWrongKindOfClosingParen@]⎦)",
			1,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 +{ 2 ++{ 3 ++{ 4 ++{ 5",
			"⎡⎡1 +{ 2⎦ ++{ ⎡3 ++{ ⎡4 ++{ 5⎦⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 +{ 2 ++{ 3 ++{ 4 ++{ 5 +++{ 9",
			"⎡⎡⎡1 +{ 2⎦ ++{ ⎡3 ++{ ⎡4 ++{ 5⎦⎦⎦ +++{ 9⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 +{ 2 ++{ 3 --{ 4 **{ 5 +++{ 9",
			"⎡⎡⎡1 +{ 2⎦ ++{ ⎡3 --{ ⎡4 **{ 5⎦⎦⎦ +++{ 9⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 [[ 2 + 3 ]",
			"⎡1 [[ ⎡2 + 3⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 (( 2 + 3 )",
			"⎡1 (( ⎡2 + 3⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 (( 2 + 3 ) (( 4 + 5 )",
			"⎡⎡1 (( ⎡2 + 3⎦⎦ (( ⎡4 + 5⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 [[ 2 + 3 ] [[ 4 ]",
			"⎡⎡1 [[ ⎡2 + 3⎦⎦ [[ 4⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 +{ ( 2 + 3 ) +{ ( 4 )",
			"⎡1 +{ ⎡(⎡2 + 3⎦) +{ (4)⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 + 2 ++ 2 + 3",
			"⎡⎡1 + 2⎦ ++ ⎡2 + 3⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 + 2 (( 2 + 3 )",
			"⎡⎡1 + 2⎦ (( ⎡2 + 3⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 ++ 2 [[ 2 + 3 ]",
			"⎡1 ++ ⎡2 [[ ⎡2 + 3⎦⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 + 2 [[[ 2 + 3 ]",
			"⎡⎡1 + 2⎦ [[[ ⎡2 + 3⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 +++ 2 [[ 2 + 3 ]",
			"⎡1 +++ ⎡2 [[ ⎡2 + 3⎦⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 [[ 2 ] [[ 3 ]",
			"⎡⎡1 [[ 2⎦ [[ 3⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 (( 2 ) (( 3 )",
			"⎡⎡1 (( 2⎦ (( 3⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 ++[ 2 ++[ 3 ++[ 4 * 5",
			"⎡1 ++[ ⎡2 ++[ ⎡3 ++[ ⎡4 * 5⎦⎦⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 ++ 2 ++ 3 ++ 4 * 5",
			"⎡⎡⎡1 ++ 2⎦ ++ 3⎦ ++ ⎡4 * 5⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 ++ 2 -- 3 ++ 4 * 5",
			"⎡⎡⎡1 ++ 2⎦ -- 3⎦ ++ ⎡4 * 5⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 +++{ 2 +++{ 3 ++{ 4 ++{ 5 + 6 + 7 + 8",
			"⎡1 +++{ ⎡2 +++{ ⎡3 ++{ ⎡4 ++{ ⎡⎡⎡5 + 6⎦ + 7⎦ + 8⎦⎦⎦⎦⎦",
			0,
		},
		{
			WithoutJuxtaposition,
			JuxLeftAssoc,
			"1 [[[ 2 + 3 ] + 4",
			"⎡⎡1 [[[ ⎡2 + 3⎦⎦ + 4⎦",
			0,
		},
	}

	pool := MakeNodePool[SimpleNode[StringElement], StringElement](64)

	for _, tst := range tsts {
		t.Logf("Input: %v\n", tst.input)

		if tst.jux == WithJuxtaposition || tst.jux == Both {
			var jux StringElement
			if tst.juxAssoc == JuxRightAssoc {
				jux = StringElement("/{")
			} else {
				jux = StringElement("/")
			}
			wJuxRoot, wJuxErrs := ParseSliceWithJuxtaposition(MakeStringElements(tst.input), &jux, pool)
			wJuxOutput := ShowSimpleNode(wJuxRoot)

			if len(wJuxErrs) != tst.nErrors {
				t.Errorf("Expected %v (with jux) errors, got %v\n", tst.nErrors, len(wJuxErrs))
			}
			if wJuxOutput != tst.output {
				t.Errorf("Expected output (with jux): %v\nGot: %v\n", tst.output, wJuxOutput)
			}
		}
		if tst.jux == WithoutJuxtaposition || tst.jux == Both {
			woJuxRoot, woJuxErrs := ParseSlice(MakeStringElements(tst.input), pool)
			woJuxOutput := ShowSimpleNode(woJuxRoot)

			if len(woJuxErrs) != tst.nErrors {
				t.Errorf("Expected %v (without jux) errors, got %v\n", tst.nErrors, len(woJuxErrs))
			}
			if woJuxOutput != tst.output {
				t.Errorf("Expected output (without jux): %v\nGot: %v\n", tst.output, woJuxOutput)
			}
		}
	}
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
