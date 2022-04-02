package opexpr

import (
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
			"⎡⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ 4⎦ ::{ ⎡9 ::{ ⎡10 ::{ nil⎦⎦⎦",
			0,
		},
		{
			JuxLeftAssoc,
			Both,
			"! 1 * ( 2 + 3 ) ++ ( ( ( 4 ::{ 9 ) ) ) ::{ 10 ::{ nil",
			"⎡⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ (((⎡4 ::{ 9⎦)))⎦ ::{ ⎡10 ::{ nil⎦⎦",
			0,
		},
		{
			JuxLeftAssoc,
			Both,
			"! 1 * ( 2 + 3 ) ++ * 4 ::{ 9 ::{ 10 ::{ nil",
			"⎡⎡⎡⎡!1⎦ * (⎡2 + 3⎦)⎦ ++ ⎡@error:ParseErrorUnexpectedOperator@* * 4⎦⎦ ::{ ⎡9 ::{ ⎡10 ::{ nil⎦⎦⎦",
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
			"1 + 2 [[ 2 + 3 ]",
			"⎡⎡1 + 2⎦ [[ ⎡2 + 3⎦⎦",
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
			"⎡1 [[ ⎡2 [[ 3⎦⎦",
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
	for j := 0; j < nArgs; j++ {
		sb.WriteString(" +{ 1")
	}
	input := sb.String()
	elems := MakeStringElements(input)

	b.ResetTimer()

	pool := MakeNodePool[SimpleNode[StringElement], StringElement](64)

	for i := 0; i < b.N; i++ {
		_, errs := ParseSlice(elems, pool)
		if len(errs) > 0 {
			panic("Not expecting to get any errors")
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
