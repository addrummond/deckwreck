// Package keywordmap provides a trie-based implemetation of a map from strings
// to indices. Its intended use is mapping strings to indices into a list of
// keywords. Strings are considered as sequences of bytes.
//
// For typical keyword sets, membership tests using keywordmap are about 1.5 – 2
// times faster than tests using map[string]int (as benchmarked on an M1 MacBook
// Air).
//
// Internally, keywordmap constructs a compact trie backed by an array of
// unsigned integers. The default (and recommended) Trie type uses 16-bit
// integers. This is sufficient for realistically sized sets of keywords. You
// may use GenericTrie[uint32] for larger tries. However, this package is not
// optimized for dealing with large sets of keywords.
package keywordmap

import "golang.org/x/exp/constraints"

//const maxEntries = int(^uintType(0)) + 1
const nodeSize = 17

// Trie is the recommended instantiation of GenericTrie. A backing array of
// uint16 suffices for sets of keywords with no more than a few hundred members.
type Trie = GenericTrie[uint16]

// GenericTrie represents a set of strings (such as the set of a programming
// language's keywords). It should be constructed only via MakeGenericTrie (or
// MakeTrie when I=uint16). The I type parameter specifies the types of the
// elements of the backing array.
type GenericTrie[I constraints.Unsigned] struct {
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
	// The total size of each trie node is therefore 17 I-sized words.
	backingSlice []I
}

// ByteIndexable is a string or byte slice
type ByteIndexable interface {
	string | []byte
}

// MakeTrie calls MakeGenericTrie with the I type parameter set to uint16 (the
// recommended default).
func MakeTrie[T ByteIndexable](keywords []T) (Trie, bool) {
	return MakeGenericTrie[uint16](keywords)
}

// MakeGenericTrie constructs a trie from a set of keywords. Each keyword is considered
// as a sequence of bytes. If your keywords have multiple possible encodings,
// you will need to add each encoding to the trie. The second return value is
// true if a suitable trie could be constructed, or false otherwise. In the
// latter case, the returned trie is empty. A trie can fail to be constructed if
// the set of keywords is too large.
func MakeGenericTrie[I constraints.Unsigned, T ByteIndexable](keywords []T) (GenericTrie[I], bool) {
	if len(keywords) == 0 {
		return MakeEmptyTrie[I](), true
	}

	maxLen := 0
	totalLen := 0
	for _, k := range keywords {
		if len(k) > maxLen {
			maxLen = len(k)
		}

		totalLen += len(k)
	}

	var trie GenericTrie[I]
	trie.backingSlice = make([]I, nodeSize*2)

	for wi, k := range keywords {
		if !AddToTrie(&trie, k, wi) {
			return MakeEmptyTrie[I](), false
		}
	}

	return trie, true
}

// MakeEmptyTrie returns an empty trie.
func MakeEmptyTrie[I constraints.Unsigned]() GenericTrie[I] {
	return GenericTrie[I]{make([]I, nodeSize*2)}
}

// AddToTrie adds word to the trie and associates it with the index wordIndex.
// It returns true if the word was successfully added to the trie, or false
// otherwise. A word can fail to be added to the trie if the trie becomes too big.
// In the case where AddToTrie returns false, part of the word may have been
// added to the trie.
//
// It is usually better to construct tries using MakeTrie. AddToTrie is useful
// if there are gaps in the sequence of indices associated with each keyword.
func AddToTrie[T ByteIndexable, I constraints.Unsigned](trie *GenericTrie[I], word T, wordIndex int) bool {
	// == MIN(maximum positive value of I, maximum positive value of int)
	max := int(^I(0))

	if wordIndex+1 >= max {
		return false
	}

	// first node in array is a dummy node that leads nowhere. it's useful for
	// slightly reducing branching in the traversal code.

	off := 1
	last := len(word)*2 - 1
	for i := 0; i < len(word)*2; i++ {
		b := (int(word[i/2]) >> (4 * ((i % 2) ^ 1))) & 0xF

		childIndexI := (off * nodeSize) + b

		if childIndexI/nodeSize >= max {
			return false
		}

		if trie.backingSlice[childIndexI] == 0 {
			if len(trie.backingSlice) >= max {
				return false
			}

			trie.backingSlice[childIndexI] = I(len(trie.backingSlice) / nodeSize)
			off = len(trie.backingSlice) / nodeSize
			trie.backingSlice = append(trie.backingSlice, make([]I, nodeSize)...)
		} else {
			off = int(trie.backingSlice[childIndexI])
		}

		if i == last {
			trie.backingSlice[off*nodeSize+nodeSize-1] = I(wordIndex + 1)
		}
	}

	return true
}

// KeywordIndex returns the index of word in the list of keywords passed to
// MakeTrie/MakeGenericTrie, or -1 if it is not present.
func KeywordIndex[T ByteIndexable, I constraints.Unsigned](trie GenericTrie[I], word T) int {
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
		// with no children, so the lookup will fail anyway. This let's us avoid
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

// GetBackingSlice returns a copy of the trie's backing slice. This value (or
// another slice with the same contents) can be passed to
// MakeTrieFromBackingArray. It serves no other purpose and has no defined
// interpretation.
func GetBackingSlice[I constraints.Unsigned](trie GenericTrie[I]) []I {
	a := make([]I, len(trie.backingSlice))
	copy(a, trie.backingSlice)
	return a
}

// MakeTrieFromBackingSlice can be used for minimal-cost construction of a trie.
// Follow these steps:
//   * Construct your desired trie in some scratch code using MakeTrie/MakeGenericTrie.
//   * Use GetBackingSlice to get the contents of the backing slice.
//   * Copy the contents of the backing slice into your code as a constant.
//   * Use this constant as the argument to MakeTrieFromBackingSlice.
// It should rarely (if ever) be necessary to initialize a trie using this function,
// as MakeTrie/MakeGenericTrie are not at all expensive.
func MakeTrieFromBackingSlice[I constraints.Unsigned](slice []I) GenericTrie[I] {
	return GenericTrie[I]{slice}
}
