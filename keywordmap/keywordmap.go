// Package keywordmap provides a trie-based implemetation of a map from strings
// to indices. Its intended use is mapping strings to indices into a list of
// keywords. Strings are considered as sequences of bytes. Both the number of
// strings and number of indices must be 'small'. Indices must be less than
// 65535, and the concatenation of the string keys must have less than 65536
// characters in total in the worst case. These restrictions are unproblematic
// for the package's intended use case, but make it wholly unsuitable as a
// general replacement for map[string]int.
//
// Internally, keywordmap constructs a compact trie backed by an array of 16 bit
// unsigned integers.
//
// For typical keyword sets, membership tests using keywordmap are about 1.5 – 2
// times faster than tests using map[string]int (as benchmarked on an M1
// MacBook Air).
package keywordmap

type uintType = uint16

const maxEntries = int(^uintType(0)) + 1
const nodeSize = 17

// Trie represents a small set of strings (such as the set of a programming
// language's keywords). It should be constructed only via MakeTrie.
type Trie struct {
	// Keywords are broken into sequences of 4 bit 'nibbles', with high nibbles
	// coming before low nibbles. Each node in the trie therefore has at most 16
	// children. This makes it feasible to use an array to store the indices of
	// child nodes.
	//
	// The root of the trie is placed at the beginning of the array. Each trie
	// node consists of (in order):
	//
	//     * 16 child node indices for every possible following nibble. 0 is used
	//       if there is no child for the relevant nibble. (As the trie is a
	//       tree, the root node cannot be the child of any node, so 0 could not
	//       be a valid child index.)
	//
	//     * 0 if no keyword terminates at this node, or 1 + i for the index i of
	//       the relevant keyword.
	//
	// The total size of each trie node is therefore 17 uintType words.
	backingSlice []uintType
	next         int
}

// ByteIndexable is a string or byte slice
type ByteIndexable interface {
	string | []byte
}

// MakeTrie constructs a trie from a set of keywords. Each keyword is considered
// as a sequence of bytes. If your keywords have multiple possible encodings,
// you will need to add each encoding to the trie. The second return value is
// true if a suitable trie could be constructed, or false otherwise. In the
// latter case, the returned trie is empty. A trie can fail to be constructed if
// the set of keywords is too large. Generally speaking, construction should
// succeed if you do not have more than a few hundred keywords.
func MakeTrie[T ByteIndexable](keywords []T) (Trie, bool) {
	if len(keywords) == 0 {
		return makeEmptyTrie(), true
	}

	maxLen := 0
	totalLen := 0
	for _, k := range keywords {
		if len(k) > maxLen {
			maxLen = len(k)
		}

		totalLen += len(k)
	}

	// worst case estimation
	size := nodeSize*(totalLen*2+len(keywords)) + nodeSize
	// if worst case estimation is greater than maxEntries, we might still be able to
	// build the trie, so worth a try
	if size > maxEntries*nodeSize {
		size = maxEntries * nodeSize
	}

	var trie Trie
	trie.backingSlice = make([]uintType, size)
	trie.next = 2

	for wi, k := range keywords {
		if !addToTrie(&trie, k, wi) {
			return makeEmptyTrie(), false
		}
	}

	return trie, true
}

func makeEmptyTrie() Trie {
	return Trie{make([]uintType, nodeSize*2), 1}
}

func addToTrie[T ByteIndexable](trie *Trie, word T, wordIndex int) bool {
	if wordIndex >= maxEntries-1 {
		return false
	}

	// first node in array is a dummy node that leads nowhere. it's useful for
	// slightly reducing branching in the traversal code.

	off := 1
	last := len(word)*2 - 1
	for i := 0; i < len(word)*2; i++ {
		b := (int(word[i/2]) >> (4 * ((i % 2) ^ 1))) & 0xF

		childIndexI := (off * nodeSize) + b

		if childIndexI >= maxEntries*nodeSize {
			return false
		}

		if trie.backingSlice[childIndexI] == 0 {
			if trie.next >= maxEntries {
				return false
			}

			trie.backingSlice[childIndexI] = uintType(trie.next)
			off = trie.next
			trie.next++
		} else {
			off = int(trie.backingSlice[childIndexI])
		}

		if i == last {
			trie.backingSlice[off*nodeSize+nodeSize-1] = uintType(wordIndex + 1)
		}
	}

	return true
}

// KeywordIndex returns the index of word in the list of keywords passed to
// MakeTrie, or -1 if it is not present.
func KeywordIndex[T ByteIndexable](trie *Trie, word T) int {
	ba := trie.backingSlice

	off := 1

	for i := 0; i < len(word); i++ {
		// loop is not speed critical in 'addToTrie', but unrolling the high and low
		// nibbles here does appreciably increase performance.

		b := int(word[i])

		b1 := b >> 4

		childIndexI := (off * nodeSize) + b1
		off = int(ba[childIndexI])
		// if off == 0 {
		// 	return -1
		// }
		// ^^^
		// if off is zero, then the next lookup will go to the dummy initial node
		// with no children, so the lookup will fail anyway. this let's us avoid
		// branching here on off == 0

		b2 := b & 0xF
		childIndexI = (off * nodeSize) + b2
		off = int(ba[childIndexI])

		if off == 0 {
			return -1
		}
	}

	return int(ba[off*nodeSize+nodeSize-1]) - 1
}

// GetBackingSlice returns a copy of the Trie's backing slice. This value (or
// another slice with the same contents) can be passed to
// MakeTrieFromBackingArray. It serves no other purpose and has no defined
// interpretation.
func GetBackingSlice(trie *Trie) []uintType {
	a := make([]uintType, len(trie.backingSlice))
	copy(a, trie.backingSlice)
	return a
}

// MakeTrieFromBackingSlice can be used for minimal-cost construction of a Trie.
// Follow these steps:
//   * Construct your desired Trie in some scratch code using MakeTrie.
//   * Use GetBackingSlice to get the contents of the backing slice.
//   * Copy the contents of the backing slice into your code as a constant.
//   * Use this constant as the argument to MakeTrieFromBackingSlice.
// It should rarely (if ever) be necessary to initialize a Trie using this function,
// as MakeTrie is not at all expensive.
func MakeTrieFromBackingSlice(slice []uintType) Trie {
	return Trie{
		backingSlice: slice,
		next:         len(slice),
	}
}
