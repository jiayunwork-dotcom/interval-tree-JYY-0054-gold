package merge

import (
	"testing"

	"interval-tree/internal/interval"
)

func iv(s, e int64) interval.Interval {
	return interval.Interval{Start: s, End: e}
}

func TestMergeOverlapping(t *testing.T) {
	in := []interval.Interval{iv(1, 5), iv(3, 8), iv(10, 12)}
	out := Merge(in)
	if len(out) != 2 {
		t.Fatalf("expected 2, got %d", len(out))
	}
	if out[0].Start != 1 || out[0].End != 8 {
		t.Fatalf("unexpected: %v", out[0])
	}
}

func TestMergeAdjacent(t *testing.T) {
	in := []interval.Interval{iv(1, 3), iv(4, 6)}
	out := Merge(in)
	if len(out) != 1 || out[0].End != 6 {
		t.Fatalf("unexpected: %v", out)
	}
}

func TestSubtract(t *testing.T) {
	base := []interval.Interval{iv(1, 10)}
	sub := []interval.Interval{iv(3, 5)}
	result := Subtract(base, sub)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestIntersect(t *testing.T) {
	a := []interval.Interval{iv(1, 5), iv(8, 12)}
	b := []interval.Interval{iv(3, 9)}
	result := Intersect(a, b)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestUnion(t *testing.T) {
	a := []interval.Interval{iv(1, 3)}
	b := []interval.Interval{iv(5, 7)}
	result := Union(a, b)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestCoveredLength(t *testing.T) {
	ivs := []interval.Interval{iv(1, 5), iv(3, 8), iv(10, 12)}
	total := CoveredLength(ivs)
	if total != 11 {
		t.Fatalf("expected 11, got %d", total)
	}
}

func TestSymmetricDifference(t *testing.T) {
	a := []interval.Interval{iv(1, 5)}
	b := []interval.Interval{iv(3, 8)}
	result := SymmetricDifference(a, b)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}
