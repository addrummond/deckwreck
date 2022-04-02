package keywordmap

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
)

func TestMakeTrieWithStrings(t *testing.T) {
	trie, ok := MakeTrie([]string{"debu", "with", "and", "for", "case", "to", "form"})
	if !ok {
		t.Errorf("Expecting trie to be constructed successfuly.")
	}

	if KeywordIndex(&trie, "debu") != 0 {
		t.Errorf("Expecting 'debug' to be in trie")
	}
	if KeywordIndex(&trie, "with") != 1 {
		t.Errorf("Expecting 'with' to be in trie")
	}
	if KeywordIndex(&trie, "and") != 2 {
		t.Errorf("Expecting 'and' to be in trie")
	}
	if KeywordIndex(&trie, "for") != 3 {
		t.Errorf("Expecting 'for' to be in trie")
	}
	if KeywordIndex(&trie, "case") != 4 {
		t.Errorf("Expecting 'case' to be in trie")
	}
	if KeywordIndex(&trie, "to") != 5 {
		t.Errorf("Expecting 'to' to be in trie")
	}
	if KeywordIndex(&trie, "form") != 6 {
		t.Errorf("Expecting 'form' to be in trie")
	}

	var s string

	for a := 0; a < 27; a++ {
		if a == 26 {
			s = ""
			expectStringNotInTrie(t, &trie, s)
			break
		}

		for b := 0; b < 27; b++ {
			if b == 26 {
				s = fmt.Sprintf("%c", rune('a'+a))
				expectStringNotInTrie(t, &trie, s)
				break
			}

			for c := 0; c < 27; c++ {
				if c == 26 {
					s = fmt.Sprintf("%c%c", rune('a'+a), rune('a'+b))
					expectStringNotInTrie(t, &trie, s)
					break
				}

				for d := 0; d < 27; d++ {
					if d == 26 {
						s = fmt.Sprintf("%c%c%c", rune('a'+a), rune('a'+b), rune('a'+c))
						expectStringNotInTrie(t, &trie, s)
						break
					}

					s = fmt.Sprintf("%c%c%c%c", rune('a'+a), rune('a'+b), rune('a'+c), rune('a'+d))
					expectStringNotInTrie(t, &trie, s)
				}
			}
		}
	}
}

func expectStringNotInTrie(t *testing.T, trie *Trie, s string) {
	if s != "debu" && s != "with" && s != "and" && s != "for" && s != "case" && s != "to" && s != "form" {
		if KeywordIndex(trie, s) != -1 {
			t.Errorf("Did not expect to find '%v' in trie", s)
		}
	}
}

func TestMakeTrieWithByteArrays(t *testing.T) {
	trie, ok := MakeTrie([][]byte{[]byte("debu"), []byte("with"), []byte("and"), []byte("for"), []byte("case"), []byte("to"), []byte("form")})
	if !ok {
		t.Errorf("Expecting trie to be constructed successfuly.")
	}

	if KeywordIndex(&trie, []byte("debu")) != 0 {
		t.Errorf("Expecting 'debug' to be in trie")
	}
	if KeywordIndex(&trie, []byte("with")) != 1 {
		t.Errorf("Expecting 'with' to be in trie")
	}
	if KeywordIndex(&trie, []byte("and")) != 2 {
		t.Errorf("Expecting 'and' to be in trie")
	}
	if KeywordIndex(&trie, []byte("for")) != 3 {
		t.Errorf("Expecting 'for' to be in trie")
	}
	if KeywordIndex(&trie, []byte("case")) != 4 {
		t.Errorf("Expecting 'case' to be in trie")
	}
	if KeywordIndex(&trie, []byte("to")) != 5 {
		t.Errorf("Expecting 'to' to be in trie")
	}
	if KeywordIndex(&trie, []byte("form")) != 6 {
		t.Errorf("Expecting 'form' to be in trie")
	}

	var ba []byte

	for a := 0; a < 27; a++ {
		if a == 26 {
			ba = []byte{}
			expectByteArrayNotInTrie(t, &trie, ba)
			break
		}

		for b := 0; b < 27; b++ {
			if b == 26 {
				ba = []byte(fmt.Sprintf("%c", rune('a'+a)))
				expectByteArrayNotInTrie(t, &trie, ba)
				break
			}

			for c := 0; c < 27; c++ {
				if c == 26 {
					ba = []byte(fmt.Sprintf("%c%c", rune('a'+a), rune('a'+b)))
					expectByteArrayNotInTrie(t, &trie, ba)
					break
				}

				for d := 0; d < 27; d++ {
					if d == 26 {
						ba = []byte(fmt.Sprintf("%c%c%c", rune('a'+a), rune('a'+b), rune('a'+c)))
						expectByteArrayNotInTrie(t, &trie, ba)
						break
					}

					ba = []byte(fmt.Sprintf("%c%c%c%c", rune('a'+a), rune('a'+b), rune('a'+c), rune('a'+d)))
					expectByteArrayNotInTrie(t, &trie, ba)
				}
			}
		}
	}
}

func expectByteArrayNotInTrie(t *testing.T, trie *Trie, ba []byte) {
	if string(ba) != "debu" && string(ba) != "with" && string(ba) != "and" && string(ba) != "for" && string(ba) != "case" && string(ba) != "to" && string(ba) != "form" {
		if KeywordIndex(trie, ba) != -1 {
			t.Errorf("Did not expect to find '%v' in trie", string(ba))
		}
	}
}

func TestTooBigTrie(t *testing.T) {
	var keywords []string
	for i := 0; i < 90000; i++ {
		keywords = append(keywords, fmt.Sprintf("%v", i))
	}
	trie, ok := MakeTrie(keywords)
	if ok {
		t.Errorf("Expecting trie to fail to be constructed")
	}

	// trie returned should be a dummy empty trie
	if KeywordIndex(&trie, "foo") != -1 || KeywordIndex(&trie, "bar") != -1 {
		t.Errorf("Expecting empty trie")
	}
}

func TestEmptyTrie(t *testing.T) {
	var keywords []string
	trie, ok := MakeTrie(keywords)
	if !ok {
		t.Errorf("Expecting trie to be constructed successfully")
	}

	if KeywordIndex(&trie, "") != -1 {
		t.Errorf("Expecting empty string to be absent")
	}
	if KeywordIndex(&trie, []byte{}) != -1 {
		t.Errorf("Expecting empty byte array to be absent")
	}

	for i := 0; i < 256; i++ {
		if KeywordIndex(&trie, []byte{byte(i)}) != -1 {
			t.Errorf("Expecting one element byte array to be absent")
		}
		if KeywordIndex(&trie, string([]byte{byte(i)})) != -1 {
			t.Errorf("Expecting one byte string to be absent")
		}
	}
}

func BenchmarkTrie(b *testing.B) {
	trie, ok := MakeTrie([]string{"debug", "with", "and", "for", "case", "to", "form"})
	if !ok {
		b.Errorf("Expecting trie to be constructed successfuly.")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if KeywordIndex(&trie, "cape") != -1 {
			panic("Internal error [1] in benchmark")
		}
		if KeywordIndex(&trie, "dooby") != -1 {
			panic("Internal error [2] in benchmark")
		}
		if KeywordIndex(&trie, "fudge") != -1 {
			panic("Internal error [3] in benchmark")
		}
		if KeywordIndex(&trie, "case") == -1 {
			panic("Internal error [4] in benchmark")
		}
		if KeywordIndex(&trie, "debug") == -1 {
			panic("Internal error [5] in benchmark")
		}
		if KeywordIndex(&trie, "for") == -1 {
			panic("Internal error [6] in benchmark")
		}
		if KeywordIndex(&trie, "form") == -1 {
			panic("Internal error [7] in benchmark")
		}
	}
}

func BenchmarkHash(b *testing.B) {
	keywords := map[string]int{
		"debug": 0,
		"with":  1,
		"and":   2,
		"for":   3,
		"case":  4,
		"to":    5,
		"form":  6,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, ok := keywords["cape"]; ok {
			panic("Internal error [8] in benchmark")
		}
		if _, ok := keywords["dooby"]; ok {
			panic("Internal error [9] in benchmark")
		}
		if _, ok := keywords["fudge"]; ok {
			panic("Internal error [10] in benchmark")
		}
		if i, ok := keywords["case"]; !ok || i != 4 {
			panic("Internal error [11] in benchmark")
		}
		if i, ok := keywords["debug"]; !ok || i != 0 {
			panic("Internal error [12] in benchmark")
		}
		if i, ok := keywords["for"]; !ok || i != 3 {
			panic("Internal error [13] in benchmark")
		}
		if i, ok := keywords["form"]; !ok || i != 6 {
			panic("Internal error [14] in benchmark")
		}
	}
}

type TestData struct {
	Keywords []string
	ToTest   []string
	InTrie   []bool
}

func BenchmarkRandomTrie(b *testing.B) {
	td := getRandomTestData()
	trie, ok := MakeTrie(td.Keywords)
	if !ok {
		b.Errorf("Expecting trie to be constructed successfuly.")
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j, w := range td.ToTest {
			if td.InTrie[j] {
				if KeywordIndex(&trie, w) == -1 {
					panic("Internal error [15] in benchmark")
				}
			} else {
				if KeywordIndex(&trie, w) != -1 {
					panic("Internal error [16] in benchmark")
				}
			}
		}
	}
}

func BenchmarkRandomHash(b *testing.B) {
	td := getRandomTestData()
	keywords := make(map[string]int)
	for i, k := range td.Keywords {
		keywords[k] = i
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for j, w := range td.ToTest {
			if td.InTrie[j] {
				if _, ok := keywords[w]; !ok {
					panic("Internal error [17] in benchmark")
				}
			} else {
				if _, ok := keywords[w]; ok {
					panic("Internal error [18] in benchmark")
				}
			}
		}
	}
}

// With current seed produces 236 keywords and 76 strings to test (including the
// 236 keywords).
func getRandomTestData() (td TestData) {
	rand.Seed(423423484)
	seen := make(map[string]struct{})

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")
	for i := 0; i < 200; i++ {
		var s string

		for {
			length := rand.Intn(9) + 1
			var sb strings.Builder
			for j := 0; j < length; j++ {
				sb.WriteRune(letterRunes[rand.Intn(len(letterRunes))])
			}

			s = sb.String()
			if _, ok := seen[s]; !ok {
				seen[s] = struct{}{}
				break
			}
		}

		if i%5 == 0 {
			td.Keywords = append(td.Keywords, s)
			td.ToTest = append(td.ToTest, s)
			td.InTrie = append(td.InTrie, true)

			// Add some prefixes to the trie
			if len(s) >= 3 {
				prefixLen := rand.Intn(len(s)-1) + 1
				prefix := s[0:prefixLen]

				if _, ok := seen[prefix]; !ok {
					td.Keywords = append(td.Keywords, prefix)
					td.ToTest = append(td.ToTest, prefix)
					td.InTrie = append(td.InTrie, true)
					seen[prefix] = struct{}{}
				}
			}
		} else {
			td.ToTest = append(td.ToTest, s)
			td.InTrie = append(td.InTrie, false)
		}
	}

	return
}
