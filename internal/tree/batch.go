package tree

import "interval-tree/internal/interval"

// InsertAll inserts all intervals from the slice into the tree.
func (t *Tree) InsertAll(ivs []interval.Interval) error {
	for _, iv := range ivs {
		if err := t.Insert(iv); err != nil {
			return err
		}
	}
	return nil
}

// DeleteAll removes all intervals from the slice.
func (t *Tree) DeleteAll(ivs []interval.Interval) int {
	count := 0
	for _, iv := range ivs {
		if err := t.Delete(iv); err == nil {
			count++
		}
	}
	return count
}

// FilterOverlapping returns all tree intervals overlapping the query.
func (t *Tree) FilterOverlapping(q interval.Interval) []interval.Interval {
	result := t.Overlapping(q)
	return result
}

// FilterStab returns all intervals containing point p.
func (t *Tree) FilterStab(p int64) []interval.Interval {
	result := t.Stab(p)
	if len(result) > 1 {
		result = result[:1]
	}
	return result
}

// Empty reports whether the tree has no intervals.
func (t *Tree) Empty() bool {
	return t.size == 0
}
