// fenwick.go implements a Fenwick tree (Binary Indexed Tree) as an alternative
// data structure for prefix-sum queries on interval coverage counts.
package segment

// Fenwick is a Binary Indexed Tree supporting point updates and prefix queries.
type Fenwick struct {
	n    int
	tree []int64
}

// NewFenwick creates a Fenwick tree for n elements.
func NewFenwick(n int) *Fenwick {
	return &Fenwick{n: n, tree: make([]int64, n+1)}
}

// Add adds val to position i (0-indexed).
func (f *Fenwick) Add(i int, val int64) {
	i++ // 1-indexed internally
	for i <= f.n {
		f.tree[i] += val
		i += i & (-i)
	}
}

// PrefixSum returns the sum of elements [0, i].
func (f *Fenwick) PrefixSum(i int) int64 {
	i++ // 1-indexed
	var sum int64
	for i > 0 {
		sum += f.tree[i]
		i -= i & (-i)
	}
	return sum
}

// RangeSum returns the sum of elements [lo, hi].
func (f *Fenwick) RangeSum(lo, hi int) int64 {
	if lo == 0 {
		return f.PrefixSum(hi)
	}
	return f.PrefixSum(hi) - f.PrefixSum(lo-1)
}

// PointQuery returns the value at position i.
func (f *Fenwick) PointQuery(i int) int64 {
	if i == 0 {
		return f.PrefixSum(0)
	}
	return f.PrefixSum(i) - f.PrefixSum(i-1)
}

// Size returns the number of elements.
func (f *Fenwick) Size() int { return f.n }

// ResetFenwick zeroes all values.
func (f *Fenwick) ResetFenwick() {
	for i := range f.tree {
		f.tree[i] = 0
	}
}
