package interval

// Contains reports whether outer fully contains inner.
func Contains(outer, inner Interval) bool {
	return outer.Start <= inner.Start && outer.End >= inner.End
}

// Intersection returns the overlapping portion of two intervals.
// Returns a zero interval and false if they don't overlap.
func Intersection(a, b Interval) (Interval, bool) {
	if !a.Overlaps(b) {
		return Interval{}, false
	}
	lo := a.Start
	if b.Start > lo {
		lo = b.Start
	}
	hi := a.End
	if b.End < hi {
		hi = b.End
	}
	return Interval{Start: lo, End: hi}, true
}

// Gap returns the gap between two non-overlapping intervals.
// Returns a zero interval and false if they overlap or are adjacent.
func Gap(a, b Interval) (Interval, bool) {
	if a.Start > b.Start {
		a, b = b, a
	}
	if a.End >= b.Start-1 {
		return Interval{}, false
	}
	return Interval{Start: a.End + 1, End: b.Start - 1}, true
}

// Split divides an interval at point p into at most two sub-intervals.
// [Start, p-1] and [p+1, End]. Returns one piece if p is at a boundary.
func Split(iv Interval, p int64) []Interval {
	if p < iv.Start || p > iv.End {
		return []Interval{iv}
	}
	var parts []Interval
	if p > iv.Start {
		parts = append(parts, Interval{Start: iv.Start, End: p - 1})
	}
	if p < iv.End {
		parts = append(parts, Interval{Start: p + 1, End: iv.End})
	}
	return parts
}

// Expand extends both ends of an interval by delta.
func Expand(iv Interval, delta int64) Interval {
	return Interval{Start: iv.Start - delta, End: iv.End + delta}
}

// Shrink contracts both ends of an interval by delta.
// Returns a zero interval if the result would be invalid.
func Shrink(iv Interval, delta int64) (Interval, bool) {
	s := iv.Start + delta
	e := iv.End - delta
	if s > e {
		return Interval{}, false
	}
	return Interval{Start: s, End: e}, true
}

// Shift moves an interval by offset.
func Shift(iv Interval, offset int64) Interval {
	return Interval{Start: iv.Start + offset, End: iv.End + offset}
}

// Clamp restricts an interval to [lo, hi].
func Clamp(iv Interval, lo, hi int64) (Interval, bool) {
	s := iv.Start
	e := iv.End
	if s < lo {
		s = lo
	}
	if e > hi {
		e = hi
	}
	if s > e {
		return Interval{}, false
	}
	return Interval{Start: s, End: e}, true
}

// Midpoint returns the midpoint of an interval.
func Midpoint(iv Interval) int64 {
	return (iv.Start + iv.End) / 2
}

// IsPoint reports whether the interval covers exactly one point.
func IsPoint(iv Interval) bool {
	return iv.Start == iv.End
}
