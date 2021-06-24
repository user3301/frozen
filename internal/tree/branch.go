package tree

import (
	"fmt"
	"log"
	"strings"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/fmtutil"
	"github.com/arr-ai/frozen/internal/iterator"
)

const fanout = 1 << fanoutBits

var (
	// UseRHS returns its RHS arg.
	UseRHS = func(_, b interface{}) interface{} { return b }

	// UseLHS returns its LHS arg.
	UseLHS = func(a, _ interface{}) interface{} { return a }
)

type branch struct {
	p packer
}

func (b *branch) Canonical(_ int) node {
	var buf [maxLeafLen]interface{}
	if data := b.CopyTo(buf[:0]); data != nil {
		return append(leaf(nil), data...)
	}
	return b
}

func (b *branch) Combine(args *CombineArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case leaf:
		result := node(b)
		for _, e := range n {
			result = result.With(args, e, depth, newHasher(e, depth), matches)
		}
		return result
	case *branch:
		var allMatches [fanout]int
		result := &branch{
			p: b.p.TransformPair(n.p, b.p.mask|n.p.mask, args.Parallel(depth),
				func(m masker, x, y node) node {
					return x.Combine(args, y, depth+1, &allMatches[m.index()])
				}),
		}
		for _, m := range allMatches {
			*matches += m
		}
		return result

	default:
		panic(errors.WTF)
	}
}

func (b *branch) CopyTo(dest []interface{}) []interface{} {
	for _, child := range b.p.data {
		if dest = child.CopyTo(dest); dest == nil {
			break
		}
	}
	return dest
}

func (b *branch) Defrost() unNode {
	u := newUnBranch()
	for m := b.p.mask; m != 0; m = m.next() {
		u.p[m.index()] = b.p.data[b.p.mask.offset(m)].Defrost()
	}
	return u
}

func (b *branch) Difference(args *EqArgs, n node, depth int, removed *int) node {
	switch n := n.(type) {
	case leaf:
		result := node(b)
		for _, e := range n {
			result = result.Without(args, e, depth, newHasher(e, depth), removed)
		}
		return result
	case *branch:
		var allMatches [fanout]int
		result := &branch{p: b.p.TransformPair(n.p, b.p.mask, args.Parallel(depth),
			func(m masker, x, y node) node {
				return x.Difference(args, y, depth+1, &allMatches[m.index()])
			})}
		for _, m := range allMatches {
			*removed += m
		}
		return result.Canonical(depth)
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Empty() bool {
	return false
}

func (b *branch) Equal(args *EqArgs, n node, depth int) bool {
	if n, is := n.(*branch); is {
		return b.p.All(n.p, b.p.mask|n.p.mask, args.Parallel(depth),
			func(m masker, a, b node) bool {
				return a.Equal(args, b, depth+1)
			})
	}
	return false
}

func (b *branch) Get(args *EqArgs, v interface{}, h hasher) *interface{} {
	return b.p.Get(newMasker(h.hash())).Get(args, v, h.next())
}

func (b *branch) Intersection(args *EqArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case leaf:
		return n.Intersection(args.flip, b, depth, matches)
	case *branch:
		var allMatches [fanout]int
		result := &branch{p: b.p.TransformPair(n.p, b.p.mask&n.p.mask, args.Parallel(depth),
			func(m masker, a, b node) node {
				return a.Intersection(args, b, depth+1, &allMatches[m.index()])
			})}
		for _, m := range allMatches {
			*matches += m
		}
		return result.Canonical(depth)
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Iterator(buf []packer) iterator.Iterator {
	return b.p.Iterator(buf)
}

func (b *branch) Reduce(args NodeArgs, depth int, r func(values ...interface{}) interface{}) interface{} {
	var results [fanout]interface{}
	b.p.All(packer{}, b.p.mask, args.Parallel(depth), func(m masker, child, _ node) bool {
		results[m.index()] = child.Reduce(args, depth+1, r)
		return true
	})
	m := b.p.mask
	acc := results[m.index()]
	for m = m.next(); m != 0; m = m.next() {
		acc = r(acc, results[m.index()])
	}
	return acc
}

func (b *branch) SubsetOf(args *EqArgs, n node, depth int) bool {
	switch n := n.(type) {
	case leaf:
		return false
	case *branch:
		return b.p.All(n.p, b.p.mask, args.Parallel(depth), func(m masker, a, b node) bool {
			return a.SubsetOf(args, b, depth+1)
		})
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Transform(args *CombineArgs, depth int, count *int, f func(v interface{}) interface{}) node {
	var results [fanout]node
	var counts [fanout]int
	b.p.All(packer{}, b.p.mask, args.Parallel(depth), func(m masker, child, _ node) bool {
		i := m.index()
		results[i] = child.Transform(args, depth+1, &counts[i], f)
		return true
	})
	m := b.p.mask
	acc := results[m.index()]
	var duplicates int
	for m = m.next(); m != 0; m = m.next() {
		acc = acc.Combine(args, results[m.index()], 0, &duplicates)
	}

	for _, c := range counts {
		*count += c
	}
	*count -= duplicates

	return acc
}

func (b *branch) Where(args *WhereArgs, depth int, matches *int) node {
	var allMatches [fanout]int
	var nodes [fanout]node
	b.p.All(packer{}, b.p.mask, args.Parallel(depth), func(m masker, a, _ node) bool {
		nodes[m.index()] = a.Where(args, depth+1, &allMatches[m.index()])
		return true
	})

	for _, m := range allMatches {
		*matches += m
	}
	return (&branch{p: packerFromNodes(&nodes)}).Canonical(depth)
}

func (b *branch) With(args *CombineArgs, v interface{}, depth int, h hasher, matches *int) node {
	i := newMasker(h.hash())
	return &branch{p: b.p.With(i, b.p.Get(i).With(args, v, depth+1, h.next(), matches))}
}

func (b *branch) Without(args *EqArgs, v interface{}, depth int, h hasher, matches *int) node {
	i := newMasker(h.hash())
	child := b.p.Get(i).Without(args, v, depth+1, h.next(), matches)
	return (&branch{p: b.p.With(i, child)}).Canonical(depth)
}

func (b *branch) Format(f fmt.State, _ rune) {
	s := b.String()
	fmt.Fprint(f, s)
	fmtutil.PadFormat(f, len(s))
}

var branchStringIndices = []string{
	"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹",
	"¹⁰", "¹¹", "¹²", "¹³", "¹⁴", "¹⁵",
}

func (b *branch) String() string {
	var buf [20]interface{}
	deep := b.CopyTo(buf[:]) != nil

	var sb strings.Builder

	sb.WriteRune('⁅')
	if deep {
		sb.WriteString("\n")
	}

	m := b.p.mask

	defer func() {
		if recover() != nil {
			log.Print(m, b.p)
		}
	}()

	for i, child := range b.p.data {
		index := branchStringIndices[m.index()]
		m = m.next()
		if deep {
			fmt.Fprintf(&sb, "   %s%s\n", index, fmtutil.IndentBlock(child.String()))
		} else {
			if i > 0 {
				sb.WriteByte(' ')
			}
			fmt.Fprintf(&sb, "%s%v", index, child)
		}
	}
	sb.WriteRune('⁆')
	return sb.String()
}