package interval

import "fmt"

// Builder provides a fluent interface for constructing interval slices.
type Builder struct {
	intervals []Interval
	err       error
}

// NewBuilder creates a new builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Add appends an interval.
func (b *Builder) Add(start, end int64) *Builder {
	if b.err != nil {
		return b
	}
	if end < start {
		b.err = fmt.Errorf("builder: invalid interval [%d, %d]", start, end)
		return b
	}
	b.intervals = append(b.intervals, Interval{Start: start, End: end})
	return b
}

// AddN appends n intervals starting from base with stride.
func (b *Builder) AddN(n int, base, stride, length int64) *Builder {
	for i := 0; i < n; i++ {
		s := base + int64(i)*stride
		b.Add(s, s+length-1)
	}
	return b
}

// Build returns the accumulated intervals or an error.
func (b *Builder) Build() ([]Interval, error) {
	if b.err != nil {
		return nil, b.err
	}
	out := make([]Interval, len(b.intervals))
	copy(out, b.intervals)
	return out, nil
}

// MustBuild panics on error, otherwise returns intervals.
func (b *Builder) MustBuild() []Interval {
	ivs, err := b.Build()
	if err != nil {
		panic(err)
	}
	return ivs
}

// Len returns intervals added so far.
func (b *Builder) Len() int { return len(b.intervals) }

// Reset clears the builder.
func (b *Builder) Reset() *Builder {
	b.intervals = nil
	b.err = nil
	return b
}

// Range generates a single interval [start, end].
func Range(start, end int64) Interval {
	return Interval{Start: start, End: end}
}

// Point generates a single-point interval [p, p].
func Point(p int64) Interval {
	return Interval{Start: p, End: p}
}

// Spans generates evenly spaced intervals.
func Spans(count int, start, stride, length int64) []Interval {
	out := make([]Interval, count)
	for i := 0; i < count; i++ {
		s := start + int64(i)*stride
		out[i] = Interval{Start: s, End: s + length - 1}
	}
	return out
}
