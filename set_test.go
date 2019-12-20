package frozen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetEmpty(t *testing.T) {
	t.Parallel()

	var s Set
	assert.True(t, s.IsEmpty())
	assertSetEqual(t, Set{}, s)
}

func TestSetWith(t *testing.T) {
	t.Parallel()

	var s Set
	arr := []interface{}{}
	for i := 0; i < 1000; i++ {
		assertSetEqual(t, NewSet(arr...), s, "i=%v", i)
		assert.Equal(t, i, s.Count(), "i=%v", i)
		assert.False(t, s.Has(i), "i=%v", i)
		s = s.With(i)
		assert.True(t, s.Has(i), "i=%v", i)
		assert.False(t, s.IsEmpty(), "i=%v", i)
		arr = append(arr, i)
	}
}

func TestSetWithout(t *testing.T) {
	t.Parallel()

	var s Set
	arr := []interface{}{}
	const N = 1000
	for i := 0; i < N; i++ {
		s = s.With(i)
		arr = append(arr, i)
	}
	for i := 0; i < N; i++ {
		assertSetEqual(t, NewSet(arr...), s, "i=%v", i)
		assert.False(t, s.IsEmpty(), "i=%v", i)
		assert.True(t, s.Has(i), "i=%v", i)
		s = s.Without(i)
		assert.False(t, s.Has(i), "i=%v", i)
		arr = arr[1:]
	}
	assert.True(t, s.IsEmpty())
}

func TestSetAny(t *testing.T) {
	t.Parallel()

	var s Set
	assert.Panics(t, func() { s.Any() })
	s = s.With(1)
	assert.Equal(t, 1, s.Any())
	s = s.With(2)
	assert.Contains(t, []int{1, 2}, s.Any())
}

func TestSetHash(t *testing.T) {
	t.Parallel()

	maps := []Set{
		{},
		NewSet(1, 2),
		NewSet(1, 3),
		NewSet(3, 4),
		NewSet(3, 5),
		NewSet(1, 3, 4),
		NewSet(1, 3, 5),
		NewSet(1, 2, 3, 4),
		NewSet(1, 2, 3, 5),
		NewSet(NewMap(KV("cc", NewSet(NewMap(KV("c", 1)))))),
		NewSet(NewMap(KV("cc", NewSet(NewMap(KV("c", 2)))))),
	}
	for i, a := range maps {
		for j, b := range maps {
			assert.Equal(t, i == j, a.Equal(b), "i=%v j=%v a=%v b=%v", i, j, a, b)
			assert.Equal(t, i == j, a.Hash(0) == b.Hash(0),
				"i=%v j=%v a=%v b=%v a.Hash()=%v b.Hash()=%v", i, j, a, b, a.Hash(0), b.Hash(0))
		}
		assert.False(t, a.Equal(42))
	}
}

func TestSetEqual(t *testing.T) {
	t.Parallel()

	sets := []Set{
		{},
		NewSet(1),
		NewSet(2),
		NewSet(1, 2),
		NewSet(1, 2, 3, 4, 5, 6, 7, 8, 9, 10),
	}
	for i, a := range sets {
		for j, b := range sets {
			assert.Equal(t, i == j, a.Equal(b))
		}
		assert.False(t, a.Equal(42))
	}
}

func TestSetIsSubsetOf(t *testing.T) {
	for i := uint64(0); i < 100; i++ {
		a := NewSetFromMask64(i)
		for j := uint64(0); j < 100; j++ {
			b := NewSetFromMask64(j)
			assert.Equal(t, i&^j == 0, a.IsSubsetOf(b), "i=%b j=%b\na=%v\nb=%v", i, j, a.root, b.root)
		}
	}
}

func TestSetString(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "{}", Set{}.String())
	assert.Equal(t, "{1}", NewSet(1).String())
	assert.Contains(t, []string{"{1, 2}", "{2, 1}"}, NewSet(1, 2).String())
	assert.Contains(t, []string{"{1, 2}", "{2, 1}"}, NewSet(2, 1).String())
}

func TestSetWhere(t *testing.T) {
	t.Parallel()

	s := Iota2(1, 10)
	multipleOf := func(n int) func(e interface{}) bool {
		return func(e interface{}) bool { return e.(int)%n == 0 }
	}
	not := func(f func(e interface{}) bool) func(e interface{}) bool {
		return func(e interface{}) bool { return !f(e) }
	}
	assertSetEqual(t, Iota3(2, 10, 2), s.Where(multipleOf(2)))
	assertSetEqual(t, Iota3(1, 10, 2), s.Where(not(multipleOf(2))))
	assertSetEqual(t, Iota3(3, 10, 3), s.Where(multipleOf(3)))
	assertSetEqual(t, NewSet(1, 2, 4, 5, 7, 8), s.Where(not(multipleOf(3))))
	assertSetEqual(t, NewSet(6), s.Where(multipleOf(2)).Where(multipleOf(3)))
	assertSetEqual(t, NewSet(6), s.Where(multipleOf(3)).Where(multipleOf(2)))
	assertSetEqual(t, NewSet(3, 9), s.Where(not(multipleOf(2))).Where(multipleOf(3)))
	assertSetEqual(t, NewSet(3, 9), s.Where(multipleOf(3)).Where(not(multipleOf(2))))
	assertSetEqual(t, NewSet(2, 4, 8), s.Where(multipleOf(2)).Where(not(multipleOf(3))))
	assertSetEqual(t, NewSet(2, 4, 8), s.Where(not(multipleOf(3))).Where(multipleOf(2)))
	assertSetEqual(t, NewSet(1, 5, 7), s.Where(not(multipleOf(2))).Where(not(multipleOf(3))))
	assertSetEqual(t, NewSet(1, 5, 7), s.Where(not(multipleOf(3))).Where(not(multipleOf(2))))
}

func TestSetMap(t *testing.T) {
	t.Parallel()

	square := func(e interface{}) interface{} { return e.(int) * e.(int) }
	div2 := func(e interface{}) interface{} { return e.(int) / 2 }
	assertSetEqual(t, Set{}, Set{}.Map(square))
	assertSetEqual(t, NewSet(1, 4, 9, 16, 25), NewSet(1, 2, 3, 4, 5).Map(square))
	assertSetEqual(t, NewSet(0, 1, 3), NewSet(1, 2, 3, 6).Map(div2))
	assert.Equal(t, 3, NewSet(1, 2, 3, 6).Map(div2).Count())
}

func TestSetReduce(t *testing.T) {
	t.Parallel()

	sum := func(acc, b interface{}) interface{} { return acc.(int) + b.(int) }
	product := func(acc, b interface{}) interface{} { return acc.(int) * b.(int) }
	assert.Equal(t, 0, Set{}.Reduce(sum, 0))
	assert.Equal(t, 1, Set{}.Reduce(product, 1))
	assert.Equal(t, 55, Iota2(1, 11).Reduce(sum, 0))
	assert.Equal(t, 720, Iota2(2, 7).Reduce(product, 1))
}

func testSetBinaryOperator(t *testing.T, bitop func(a, b uint64) uint64, setop func(a, b Set) Set) {
	t.Parallel()

	m := map[uint64]struct{}{
		0b000000000000000: {},
		0b000000000000001: {},
		0b000000000000010: {},
		0b000001000010000: {},
		0b111111111111111: {},
		0b010101010101010: {},
		0b100100100100100: {},
		0b000100010001000: {},
		0b100001000010000: {},
	}
	for i := 0; i < 100; i++ {
		m[uint64(i)] = struct{}{}
	}
	for i := 100; i < 10_000; i += 100 {
		m[uint64(i)] = struct{}{}
	}
	for i := 10_000; i < 1_000_000; i += 10_000 {
		m[uint64(i)] = struct{}{}
	}
	for i := 1_000_000; i < 100_000_000; i += 1_000_000 {
		m[uint64(i)] = struct{}{}
	}
	sets := make([]uint64, 0, len(m))
	for set := range m {
		sets = append(sets, set)
	}

	for _, x := range sets {
		for _, y := range sets {
			sx := NewSetFromMask64(x)
			sy := NewSetFromMask64(y)
			sxy := NewSetFromMask64(bitop(x, y))
			assertSetEqual(t, sxy, setop(sx, sy), "sx=%v sy=%v", sx, sy)
		}
	}
}

func TestSetIntersection(t *testing.T) {
	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a & b },
		func(a, b Set) Set { return a.Intersection(b) },
	)
}

func TestSetUnion(t *testing.T) {
	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a | b },
		func(a, b Set) Set { return a.Union(b) },
	)
}

func TestSetMinus(t *testing.T) {
	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a &^ b },
		func(a, b Set) Set { return a.Minus(b) },
	)
}

func TestSetSymmetricDifference(t *testing.T) {
	testSetBinaryOperator(t,
		func(a, b uint64) uint64 { return a ^ b },
		func(a, b Set) Set { return a.SymmetricDifference(b) },
	)
}

func TestSetGroupBy(t *testing.T) {
	t.Parallel()

	const N = 100
	const D = 7
	group := Iota(N).GroupBy(func(el interface{}) interface{} {
		return el.(int) % D
	})
	var b MapBuilder
	for i := 0; i < D; i++ {
		b.Put(i, Iota3(i, N, D))
	}
	assertMapEqual(t, b.Finish(), group)
}

func TestSetRange(t *testing.T) {
	t.Parallel()

	mask := uint64(0)
	for i := Iota(64).Range(); i.Next(); {
		mask |= uint64(1) << i.Value().(int)
	}
	assert.Equal(t, ^uint64(0), mask)
}

func TestSetRangeBy(t *testing.T) {
	t.Parallel()

	output := []int{}
	for i := Iota(10).RangeBy(func(a, b interface{}) bool { return a.(int) < b.(int) }); i.Next(); {
		output = append(output, i.Value().(int))
	}
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}, output)

	output = output[:0]
	for i := Iota(10).RangeBy(func(a, b interface{}) bool { return a.(int) > b.(int) }); i.Next(); {
		output = append(output, i.Value().(int))
	}
	assert.Equal(t, []int{9, 8, 7, 6, 5, 4, 3, 2, 1, 0}, output)
}

var prepopSetInt = memoizePrepop(func(n int) interface{} {
	m := map[int]struct{}{}
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInt(b *testing.B, n int) {
	m := prepopSetInt(n).(map[int]struct{})
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = struct{}{}
	}
}

func BenchmarkInsertSetInt0(b *testing.B) {
	benchmarkInsertSetInt(b, 0)
}

func BenchmarkInsertSetInt1k(b *testing.B) {
	benchmarkInsertSetInt(b, 1<<10)
}

func BenchmarkInsertSetInt1M(b *testing.B) {
	benchmarkInsertSetInt(b, 1<<20)
}

var prepopSetInterface = memoizePrepop(func(n int) interface{} {
	m := map[interface{}]struct{}{}
	for i := 0; i < n; i++ {
		m[i] = struct{}{}
	}
	return m
})

func benchmarkInsertSetInterface(b *testing.B, n int) {
	m := prepopSetInterface(n).(map[interface{}]struct{})
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		m[i] = struct{}{}
	}
}

func BenchmarkInsertSetInterface0(b *testing.B) {
	benchmarkInsertSetInterface(b, 0)
}

func BenchmarkInsertSetInterface1k(b *testing.B) {
	benchmarkInsertSetInterface(b, 1<<10)
}

func BenchmarkInsertSetInterface1M(b *testing.B) {
	benchmarkInsertSetInterface(b, 1<<20)
}

var prepopFrozenSet = memoizePrepop(func(n int) interface{} {
	var s Set
	for i := 0; i < n; i++ {
		s = s.With(i)
	}
	return s
})

func benchmarkInsertFrozenSet(b *testing.B, n int) {
	s := prepopFrozenSet(n).(Set)
	b.ResetTimer()
	for i := n; i < n+b.N; i++ {
		s.With(i)
	}
}

func BenchmarkInsertFrozenSet0(b *testing.B) {
	benchmarkInsertFrozenSet(b, 0)
}

func BenchmarkInsertFrozenSet1k(b *testing.B) {
	benchmarkInsertFrozenSet(b, 1<<10)
}

func BenchmarkInsertFrozenSet1M(b *testing.B) {
	benchmarkInsertFrozenSet(b, 1<<20)
}