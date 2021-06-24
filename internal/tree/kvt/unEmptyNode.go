// Generated by gen-kv.pl. DO NOT EDIT.
package kvt

import (
	"github.com/arr-ai/frozen/pkg/kv"
)

type unEmptyNode struct{}

var _ unNode = unEmptyNode{}

func (e unEmptyNode) Add(args *CombineArgs, v kv.KeyValue, depth int, h hasher, matches *int) unNode {
	l := newUnLeaf()
	return l.Add(args, v, depth, h, matches)
}

func (unEmptyNode) appendTo(dest []kv.KeyValue) []kv.KeyValue {
	return dest
}

func (unEmptyNode) Freeze() node {
	return leaf(nil)
}

func (e unEmptyNode) Get(args *EqArgs, v kv.KeyValue, h hasher) *kv.KeyValue {
	return nil
}

func (e unEmptyNode) Remove(_ *EqArgs, _ kv.KeyValue, _ int, _ hasher, _ *int) unNode {
	return e
}