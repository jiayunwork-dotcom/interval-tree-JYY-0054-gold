package sweep

import (
	"testing"

	"interval-tree/internal/interval"
)

func iv(s, e int64) interval.Interval {
	return interval.Interval{Start: s, End: e}
}

func TestGenerateEvents(t *testing.T) {
	ivs := []interval.Interval{iv(1, 5), iv(3, 8)}
	events := GenerateEvents(ivs)
	if len(events) != 4 {
		t.Fatalf("expected 4 events, got %d", len(events))
	}
}

func TestMaxCoverage(t *testing.T) {
	ivs := []interval.Interval{iv(1, 5), iv(3, 7), iv(4, 9)}
	max := MaxCoverage(ivs)
	if max != 3 {
		t.Fatalf("expected 3, got %d", max)
	}
}

func TestCoverageProfile(t *testing.T) {
	ivs := []interval.Interval{iv(2, 4), iv(3, 5)}
	profile := CoverageProfile(ivs, 1, 6)
	expected := []int{0, 1, 2, 2, 1, 0}
	for i, v := range expected {
		if profile[i] != v {
			t.Fatalf("pos %d: expected %d, got %d", i+1, v, profile[i])
		}
	}
}

func TestUncoveredPoints(t *testing.T) {
	ivs := []interval.Interval{iv(2, 4), iv(7, 9)}
	uncovered := UncoveredPoints(ivs, 1, 10)
	if len(uncovered) != 4 {
		t.Fatalf("expected 4, got %d", len(uncovered))
	}
}

func TestCoverageRatio(t *testing.T) {
	ivs := []interval.Interval{iv(1, 5)}
	ratio := CoverageRatio(ivs, 1, 10)
	if ratio < 0.49 || ratio > 0.51 {
		t.Fatalf("expected ~0.5, got %f", ratio)
	}
}

func TestMinCoverage(t *testing.T) {
	ivs := []interval.Interval{iv(1, 10), iv(3, 7)}
	min := MinCoverage(ivs)
	if min != 1 {
		t.Fatalf("expected 1, got %d", min)
	}
}
