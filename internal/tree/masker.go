package tree

import (
	"github.com/arr-ai/frozen/errors"
)

// masker represents a set of one-bits and the ability to enumerate them.
type masker uint16

func newMasker(i int) masker {
	return masker(1) << i
}

// first returns a masker with only the low bit of m.
func (m masker) first() masker {
	return m &^ (m - 1)
}

// firstIsIn returns true if, and only if, the low bit of m is also in n.
func (m masker) firstIsIn(n masker) bool {
	return m.first().subsetOf(n)
}

func (m masker) subsetOf(mask masker) bool {
	return m&^mask == 0
}

func (m masker) String() string {
	panic(errors.Unimplemented)
	// return brailleEncoded(bits.Reverse64(uint64(m)))
}
