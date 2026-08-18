package tree

import "interval-tree/internal/interval"

// Iterator provides in-order traversal of the interval tree.
type Iterator struct {
	stack []*node
	cur   *node
}

// Iter returns an iterator for in-order traversal of all intervals.
func (t *Tree) Iter() *Iterator {
	it := &Iterator{}
	it.pushLeft(t.root)
	return it
}

// Valid reports whether the iterator points to a valid node.
func (it *Iterator) Valid() bool {
	return len(it.stack) > 0
}

// Next advances the iterator and returns the current interval.
func (it *Iterator) Next() interval.Interval {
	if !it.Valid() {
		return interval.Interval{}
	}
	n := it.stack[len(it.stack)-1]
	it.stack = it.stack[:len(it.stack)-1]
	it.pushLeft(n.right)
	return n.iv
}

// Collect returns all remaining intervals from the iterator.
func (it *Iterator) Collect() []interval.Interval {
	var result []interval.Interval
	for it.Valid() {
		result = append(result, it.Next())
	}
	return result
}

func (it *Iterator) pushLeft(n *node) {
	for n != nil {
		it.stack = append(it.stack, n)
		n = n.left
	}
}
