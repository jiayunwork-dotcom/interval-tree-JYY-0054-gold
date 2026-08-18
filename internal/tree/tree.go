// Package tree implements an augmented binary search tree of intervals.
//
// The tree is keyed by the interval's Start endpoint (ties broken by End),
// which makes it a proper BST over interval.Interval. Every node additionally
// stores the maximum End value found anywhere in its subtree (the
// "maxEnd" augmentation). This single extra field lets the two query
// operations - Stab (find every interval covering a point) and Overlapping
// (find every interval overlapping a query interval) run in O(log n + k)
// average time instead of scanning the whole tree, because subtrees whose
// maxEnd cannot reach the query are pruned away.
//
// The underlying BST is kept balanced with AVL rotations, so Insert and
// Delete stay O(log n) even under adversarial input order. The maxEnd field
// is recomputed from the children on every structural change, which is what
// keeps the augmentation correct after insertions and deletions.
package tree

import (
	"errors"
	"interval-tree/internal/interval"
	"math"
	"sort"
)

// ErrNotFound is returned by Delete when the requested interval is not
// present in the tree.
var ErrNotFound = errors.New("tree: interval not found")

// ErrInvalid is returned when an operation receives an interval that fails
// interval.Interval.Valid (End < Start).
var ErrInvalid = errors.New("tree: invalid interval (End < Start)")

// node is a single element of the AVL tree.
//
// Invariants:
//   - the subtree rooted at n is a BST under CompareKey
//   - n.height is the height of the subtree (0 for a leaf's nil child)
//   - n.maxEnd equals max(n.iv.End, maxEnd(left), maxEnd(right))
//   - n.count is the multiplicity of n.iv (how many Insert calls landed here)
type node struct {
	iv     interval.Interval
	count  int
	height int8
	maxEnd int64
	left   *node
	right  *node
}

// Tree is an augmented interval BST.
//
// The zero value is an empty, ready to use tree. The type is not safe for
// concurrent use; callers that need that should guard with a mutex.
type Tree struct {
	root *node
	size int // total number of stored intervals (counts duplicated keys)
}

func maxI64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

// height returns the cached height of a subtree (0 for nil).
func height(n *node) int {
	if n == nil {
		return 0
	}
	return int(n.height)
}

// maxEnd returns the maximum End in a subtree, or math.MinInt64 when the
// subtree is empty. Using the sentinel keeps the augmentation math branch
// free and correct.
func maxEnd(n *node) int64 {
	if n == nil {
		return math.MinInt64
	}
	return n.maxEnd
}

// update recomputes height and maxEnd from the node's children. It must be
// called whenever the subtree under n changes.
func (n *node) update() {
	h := height(n.left)
	if hr := height(n.right); hr > h {
		h = hr
	}
	n.height = int8(h + 1)

	m := n.iv.End
	if ml := maxEnd(n.left); ml > m {
		m = ml
	}
	if mr := maxEnd(n.right); mr > m {
		m = mr
	}
	n.maxEnd = m
}

// balance performs the AVL rebalancing step for the subtree rooted at n and
// returns the (possibly new) root of that subtree. It assumes n.update has
// not yet been applied for the latest change, so it updates first.
func balance(n *node) *node {
	n.update()
	b := height(n.left) - height(n.right)
	if b > 1 {
		// Left heavy.
		if height(n.left.left) < height(n.left.right) {
			n.left = rotateLeft(n.left)
		}
		return rotateRight(n)
	}
	if b < -1 {
		// Right heavy.
		if height(n.right.right) < height(n.right.left) {
			n.right = rotateRight(n.right)
		}
		return rotateLeft(n)
	}
	return n
}

func rotateRight(x *node) *node {
	y := x.left
	t2 := y.right
	y.right = x
	x.left = t2
	x.update()
	y.update()
	return y
}

func rotateLeft(x *node) *node {
	y := x.right
	t2 := y.left
	y.left = x
	x.right = t2
	x.update()
	y.update()
	return y
}

// Insert adds iv to the tree. Duplicate intervals (same Start and End) are
// stored with an incremented multiplicity rather than as separate nodes.
// Inserting an invalid interval returns ErrInvalid and leaves the tree
// unchanged.
func (t *Tree) Insert(iv interval.Interval) error {
	if !iv.Valid() {
		return ErrInvalid
	}
	t.root = t.insert(t.root, iv)
	t.size++
	return nil
}

func (t *Tree) insert(n *node, iv interval.Interval) *node {
	if n == nil {
		return &node{iv: iv, count: 1, height: 1, maxEnd: iv.End}
	}
	switch cmp := iv.CompareKey(n.iv); {
	case cmp < 0:
		n.left = t.insert(n.left, iv)
	case cmp > 0:
		n.right = t.insert(n.right, iv)
	default:
		n.count++
		n.update()
		return n
	}
	return balance(n)
}

// Delete removes one occurrence of iv from the tree. When the interval's
// multiplicity drops to zero the node is physically removed (using the
// in-order successor for the two-children case, which transfers the
// successor's remaining multiplicity to the node being kept). Deleting an
// interval that is not present returns ErrNotFound; an invalid interval
// returns ErrInvalid. In both error cases the tree is left unchanged.
func (t *Tree) Delete(iv interval.Interval) error {
	if !iv.Valid() {
		return ErrInvalid
	}
	if t.root == nil {
		return ErrNotFound
	}
	if !t.contains(t.root, iv) {
		return ErrNotFound
	}
	t.root = t.delete(t.root, iv)
	t.size--
	return nil
}

func (t *Tree) contains(n *node, iv interval.Interval) bool {
	for n != nil {
		switch cmp := iv.CompareKey(n.iv); {
		case cmp < 0:
			n = n.left
		case cmp > 0:
			n = n.right
		default:
			return true
		}
	}
	return false
}

func (t *Tree) delete(n *node, iv interval.Interval) *node {
	if n == nil {
		return nil
	}
	switch cmp := iv.CompareKey(n.iv); {
	case cmp < 0:
		n.left = t.delete(n.left, iv)
	case cmp > 0:
		n.right = t.delete(n.right, iv)
	default:
		n.count--
		if n.count > 0 {
			n.update()
			return n
		}
		if n.left == nil {
			return n.right
		}
		if n.right == nil {
			return n.left
		}
		var succ *node
		n.right, succ = removeMin(n.right)
		n.iv = succ.iv
		n.count = succ.count
	}
	return balance(n)
}

// removeMin detaches and returns the node with the smallest key in the
// subtree rooted at n, returning the rebalanced remainder as the first
// result.
func removeMin(n *node) (*node, *node) {
	if n.left == nil {
		return n.right, n
	}
	var min *node
	n.left, min = removeMin(n.left)
	return balance(n), min
}

// Len returns the total number of intervals stored, counting duplicates.
func (t *Tree) Len() int {
	return t.size
}

// MaxEnd returns the largest End value among all stored intervals, or
// math.MinInt64 when the tree is empty.
func (t *Tree) MaxEnd() int64 {
	return maxEnd(t.root)
}

// Contains reports whether iv is present in the tree.
func (t *Tree) Contains(iv interval.Interval) bool {
	return t.contains(t.root, iv)
}

// All returns every stored interval, sorted by Start then End. Duplicates
// appear according to their multiplicity. The returned slice is a fresh
// allocation; mutating it does not affect the tree.
func (t *Tree) All() []interval.Interval {
	out := make([]interval.Interval, 0, t.size)
	t.collect(t.root, &out)
	sort.Slice(out, func(i, j int) bool { return out[i].CompareKey(out[j]) < 0 })
	return out
}

func (t *Tree) collect(n *node, out *[]interval.Interval) {
	if n == nil {
		return
	}
	t.collect(n.left, out)
	for i := 0; i < n.count; i++ {
		*out = append(*out, n.iv)
	}
	t.collect(n.right, out)
}

// Stab returns every interval that covers the point p, i.e. Start <= p <=
// End. The result is sorted by Start then End and may contain duplicates
// matching the stored multiplicity. An empty tree yields a non-nil empty
// slice.
func (t *Tree) Stab(p int64) []interval.Interval {
	out := make([]interval.Interval, 0)
	t.stab(t.root, p, &out)
	sort.Slice(out, func(i, j int) bool { return out[i].CompareKey(out[j]) < 0 })
	return out
}

func (t *Tree) stab(n *node, p int64, out *[]interval.Interval) {
	if n == nil {
		return
	}
	// No interval in the left subtree can cover p unless its maxEnd reaches p.
	if maxEnd(n.left) >= p {
		t.stab(n.left, p, out)
	}
	if n.iv.Covers(p) {
		*out = append(*out, n.iv)
	}
	// Every interval in the right subtree has Start > n.iv.Start. If n.iv
	// already starts after p, the right subtree cannot contain a covering
	// interval either.
	if n.iv.Start <= p {
		t.stab(n.right, p, out)
	}
}

// Overlapping returns every interval that shares at least one integer with
// the query interval q. The result is sorted by Start then End and reflects
// multiplicities. An empty tree yields a non-nil empty slice.
func (t *Tree) Overlapping(q interval.Interval) []interval.Interval {
	out := make([]interval.Interval, 0)
	t.overlap(t.root, q, &out)
	sort.Slice(out, func(i, j int) bool { return out[i].CompareKey(out[j]) < 0 })
	return out
}

func (t *Tree) overlap(n *node, q interval.Interval, out *[]interval.Interval) {
	if n == nil {
		return
	}
	// An interval overlapping q must have End >= q.Start; prune the left
	// subtree when its maxEnd cannot reach that far.
	if maxEnd(n.left) >= q.Start {
		t.overlap(n.left, q, out)
	}
	if n.iv.Overlaps(q) {
		for i := 0; i < n.count; i++ {
			*out = append(*out, n.iv)
		}
	}
	// An interval overlapping q must have Start <= q.End; the right subtree
	// only holds larger starts, so prune when n.iv.Start is already past q.End.
	if n.iv.Start <= q.End {
		t.overlap(n.right, q, out)
	}
}

// Clear removes every interval, leaving the tree empty and ready for reuse.
func (t *Tree) Clear() {
	t.root = nil
	t.size = 0
}
