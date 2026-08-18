package query

import (
	"testing"

	"interval-tree/internal/interval"
)

func iv(s, e int64) interval.Interval {
	return interval.Interval{Start: s, End: e}
}

func TestStabPoint(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(1, 5), iv(3, 8), iv(10, 12)})
	result := c.StabPoint(4)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestOverlapping(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(1, 3), iv(5, 7), iv(10, 15)})
	result := c.Overlapping(iv(2, 6))
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestContaining(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(1, 10), iv(3, 7), iv(20, 30)})
	result := c.Containing(iv(4, 6))
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestContainedBy(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(1, 3), iv(5, 7), iv(2, 8)})
	result := c.ContainedBy(iv(1, 10))
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
}

func TestNearest(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(1, 3), iv(10, 12), iv(20, 22)})
	near, ok := c.Nearest(9)
	if !ok {
		t.Fatal("expected result")
	}
	if near.Start != 10 {
		t.Fatalf("expected [10,12], got [%d,%d]", near.Start, near.End)
	}
}

func TestGaps(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(3, 5), iv(8, 10)})
	gaps := c.Gaps(1, 12)
	if len(gaps) != 3 {
		t.Fatalf("expected 3 gaps, got %d", len(gaps))
	}
}

func TestSpan(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(5, 10), iv(1, 3), iv(20, 25)})
	span, ok := c.Span()
	if !ok || span.Start != 1 || span.End != 25 {
		t.Fatalf("unexpected span: [%d,%d]", span.Start, span.End)
	}
}

func TestMaxGap(t *testing.T) {
	c := NewCollection([]interval.Interval{iv(1, 3), iv(10, 12), iv(20, 22)})
	g := c.MaxGap()
	if g != 7 {
		t.Fatalf("expected 7, got %d", g)
	}
}
