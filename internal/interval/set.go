package interval

import "sort"

// Set is an ordered collection of unique intervals.
type Set struct {
	items []Interval
}

// NewSet creates a set from intervals, deduplicating.
func NewSet(ivs []Interval) *Set {
	s := &Set{}
	seen := make(map[Interval]struct{})
	for _, iv := range ivs {
		if _, ok := seen[iv]; !ok {
			seen[iv] = struct{}{}
			s.items = append(s.items, iv)
		}
	}
	sort.Slice(s.items, func(i, j int) bool {
		return s.items[i].CompareKey(s.items[j]) < 0
	})
	return s
}

// Add inserts an interval if not already present.
func (s *Set) Add(iv Interval) bool {
	for _, existing := range s.items {
		if existing.Equal(iv) {
			return false
		}
	}
	s.items = append(s.items, iv)
	sort.Slice(s.items, func(i, j int) bool {
		return s.items[i].CompareKey(s.items[j]) < 0
	})
	return true
}

// Remove removes an interval if present.
func (s *Set) Remove(iv Interval) bool {
	for i, existing := range s.items {
		if existing.Equal(iv) {
			s.items = append(s.items[:i], s.items[i+1:]...)
			return true
		}
	}
	return false
}

// Contains reports whether iv is in the set.
func (s *Set) Contains(iv Interval) bool {
	for _, existing := range s.items {
		if existing.Equal(iv) {
			return true
		}
	}
	return false
}

// Len returns the number of unique intervals.
func (s *Set) Len() int { return len(s.items) }

// Items returns a copy of all intervals in order.
func (s *Set) Items() []Interval {
	out := make([]Interval, len(s.items))
	copy(out, s.items)
	return out
}

// Clear removes all intervals.
func (s *Set) Clear() {
	s.items = nil
}

// Union returns a new set containing all intervals from both.
func (s *Set) Union(other *Set) *Set {
	all := make([]Interval, 0, s.Len()+other.Len())
	all = append(all, s.items...)
	all = append(all, other.items...)
	return NewSet(all)
}

// Intersection returns intervals present in both sets.
func (s *Set) Intersection(other *Set) *Set {
	var common []Interval
	for _, iv := range s.items {
		if other.Contains(iv) {
			common = append(common, iv)
		}
	}
	return NewSet(common)
}

// Difference returns intervals in s but not in other.
func (s *Set) Difference(other *Set) *Set {
	var diff []Interval
	for _, iv := range s.items {
		if !other.Contains(iv) {
			diff = append(diff, iv)
		}
	}
	return NewSet(diff)
}
