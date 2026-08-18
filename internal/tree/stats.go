package tree

import "interval-tree/internal/interval"

// Stats holds computed statistics about the tree.
type Stats struct {
	Size      int
	Height    int
	MinStart  int64
	MaxEnd    int64
	AvgLength float64
}

// ComputeStats returns structural statistics.
func (t *Tree) ComputeStats() Stats {
	if t.root == nil {
		return Stats{}
	}
	s := Stats{
		Size:   t.size,
		Height: nodeHeight(t.root),
	}
	var sum int64
	count := 0
	s.MinStart = t.root.iv.Start
	s.MaxEnd = t.root.iv.End
	t.walk(t.root, func(n *node) {
		if n.iv.Start < s.MinStart {
			s.MinStart = n.iv.Start
		}
		if n.iv.End > s.MaxEnd {
			s.MaxEnd = n.iv.End
		}
		sum += n.iv.End - n.iv.Start + 1
		count++
	})
	if count > 0 {
		s.AvgLength = float64(sum) / float64(count)
	}
	return s
}

// AllIntervals returns all intervals in sorted order.
func (t *Tree) AllIntervals() []interval.Interval {
	var result []interval.Interval
	t.walk(t.root, func(n *node) {
		for i := 0; i < n.count; i++ {
			result = append(result, n.iv)
		}
	})
	return result
}

// Count returns the total number of intervals (with multiplicities).
func (t *Tree) Count() int {
	return t.size
}

func (t *Tree) walk(n *node, fn func(*node)) {
	if n == nil {
		return
	}
	t.walk(n.left, fn)
	fn(n)
	t.walk(n.right, fn)
}

func nodeHeight(n *node) int {
	if n == nil {
		return 0
	}
	l := nodeHeight(n.left)
	r := nodeHeight(n.right)
	if l > r {
		return l + 1
	}
	return r + 1
}
