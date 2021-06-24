package tree

import (
	"fmt"
)

type node interface {
	fmt.Stringer

	Canonical(depth int) node
	Combine(args *CombineArgs, n node, depth int, matches *int) node
	CopyTo(dest []elementT) []elementT
	Defrost() unNode
	Difference(args *EqArgs, n node, depth int, removed *int) node
	Empty() bool
	Equal(args *EqArgs, n node, depth int) bool
	Get(args *EqArgs, v elementT, h hasher) *elementT
	Intersection(args *EqArgs, n node, depth int, matches *int) node
	Iterator(buf [][]node) Iterator
	Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT
	SubsetOf(args *EqArgs, n node, depth int) bool
	Transform(args *CombineArgs, depth int, count *int, f func(v elementT) elementT) node
	Where(args *WhereArgs, depth int, matches *int) node
	With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node
	Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node
}
