// Package segment implements a segment tree for efficient range queries on
// intervals. Unlike the interval tree (which answers "which intervals overlap
// a query"), the segment tree answers aggregate queries like "how many intervals
// cover each point" or "what is the sum/min/max over a range".
package segment

// Tree is a segment tree supporting range updates and point queries.
// It uses lazy propagation for efficient range additions.
type Tree struct {
	n    int
	tree []int64
	lazy []int64
}

// New creates a segment tree for the range [0, n).
func New(n int) *Tree {
	size := 4 * n
	return &Tree{
		n:    n,
		tree: make([]int64, size),
		lazy: make([]int64, size),
	}
}

// Update adds value to all positions in [lo, hi].
func (t *Tree) Update(lo, hi int, value int64) {
	t.update(1, 0, t.n-1, lo, hi, value)
}

// Query returns the value at position pos.
func (t *Tree) Query(pos int) int64 {
	return t.query(1, 0, t.n-1, pos)
}

// RangeSum returns the sum of values in [lo, hi].
func (t *Tree) RangeSum(lo, hi int) int64 {
	var sum int64
	for i := lo; i <= hi; i++ {
		sum += t.Query(i)
	}
	return sum
}

// RangeMax returns the maximum value in [lo, hi].
func (t *Tree) RangeMax(lo, hi int) int64 {
	max := t.Query(lo)
	for i := lo + 1; i <= hi; i++ {
		v := t.Query(i)
		if v > max {
			max = v
		}
	}
	return max
}

// RangeMin returns the minimum value in [lo, hi].
func (t *Tree) RangeMin(lo, hi int) int64 {
	min := t.Query(lo)
	for i := lo + 1; i <= hi; i++ {
		v := t.Query(i)
		if v < min {
			min = v
		}
	}
	return min
}

// Size returns the range size.
func (t *Tree) Size() int { return t.n }

// Reset zeroes all values.
func (t *Tree) Reset() {
	for i := range t.tree {
		t.tree[i] = 0
	}
	for i := range t.lazy {
		t.lazy[i] = 0
	}
}

// Snapshot returns the value at every position.
func (t *Tree) Snapshot() []int64 {
	out := make([]int64, t.n)
	for i := 0; i < t.n; i++ {
		out[i] = t.Query(i)
	}
	return out
}

func (t *Tree) update(node, start, end, lo, hi int, val int64) {
	if lo > end || hi < start {
		return
	}
	if lo <= start && end <= hi {
		t.tree[node] += val
		t.lazy[node] += val
		return
	}
	mid := (start + end) / 2
	t.pushDown(node)
	t.update(2*node, start, mid, lo, hi, val)
	t.update(2*node+1, mid+1, end, lo, hi, val)
}

func (t *Tree) query(node, start, end, pos int) int64 {
	if start == end {
		return t.tree[node]
	}
	t.pushDown(node)
	mid := (start + end) / 2
	if pos <= mid {
		return t.query(2*node, start, mid, pos)
	}
	return t.query(2*node+1, mid+1, end, pos)
}

func (t *Tree) pushDown(node int) {
	if t.lazy[node] != 0 {
		t.tree[2*node] += t.lazy[node]
		t.lazy[2*node] += t.lazy[node]
		t.tree[2*node+1] += t.lazy[node]
		t.lazy[2*node+1] += t.lazy[node]
		t.lazy[node] = 0
	}
}
