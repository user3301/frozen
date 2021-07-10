package tree

import (
	"github.com/arr-ai/frozen/internal/depth"
)

var (
	// DefaultNPEqArgs provides default equality with non-parallel behaviour.
	DefaultNPEqArgs = NewDefaultEqArgs(depth.NonParallel)

	// DefaultNPCombineArgs provides default combiner with non-parallel
	// behaviour.
	DefaultNPCombineArgs = NewCombineArgs(DefaultNPEqArgs, UseRHS)
)

type NodeArgs struct {
	depth.Gauge
}

func NewNodeArgs(gauge depth.Gauge) NodeArgs {
	return NodeArgs{
		Gauge: gauge,
	}
}

type CombineArgs struct {
	*EqArgs

	f func(a, b elementT) elementT

	flipped *CombineArgs
}

func NewCombineArgs(ea *EqArgs, combine func(a, b elementT) elementT) *CombineArgs {
	return &CombineArgs{EqArgs: ea, f: combine}
}

func (a *CombineArgs) Flip() *CombineArgs {
	if a.flipped == nil {
		f := a.f
		a.flipped = &CombineArgs{
			EqArgs:  a.EqArgs.Flip(),
			f:       func(a, b elementT) elementT { return f(b, a) },
			flipped: a,
		}
	}
	return a.flipped
}

type EqArgs struct {
	NodeArgs

	eq func(a, b elementT) bool
	// TODO
	lhash, rhash func(a elementT, seed uintptr) uintptr

	flipped *EqArgs
}

func NewEqArgs(
	gauge depth.Gauge,
	eq func(a, b elementT) bool,
	lhash, rhash func(a elementT, seed uintptr) uintptr,
) *EqArgs {
	na := NewNodeArgs(gauge)
	return &EqArgs{
		NodeArgs: na,
		eq:       eq,
		lhash:    lhash,
		rhash:    rhash,
	}
}

func NewDefaultEqArgs(gauge depth.Gauge) *EqArgs {
	return NewEqArgs(gauge, elementEqual, hashValue, hashValue)
}

func (a *EqArgs) Flip() *EqArgs {
	if a.flipped == nil {
		eq := a.eq
		a.flipped = &EqArgs{
			NodeArgs: a.NodeArgs,
			eq:       func(a, b elementT) bool { return eq(b, a) },
			lhash:    a.rhash,
			rhash:    a.lhash,
			flipped:  a,
		}
	}
	return a.flipped
}

type WhereArgs struct {
	NodeArgs

	Pred func(elem elementT) bool
}
