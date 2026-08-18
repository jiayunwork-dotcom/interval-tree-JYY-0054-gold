package interval

import "math/rand"

// Random generates a random valid interval within [lo, hi].
func Random(rng *rand.Rand, lo, hi int64) Interval {
	if hi <= lo {
		return Interval{Start: lo, End: lo}
	}
	span := hi - lo
	a := lo + rng.Int63n(span)
	b := lo + rng.Int63n(span)
	if a > b {
		a, b = b, a
	}
	return Interval{Start: a, End: b}
}

// RandomN generates n random intervals within [lo, hi].
func RandomN(rng *rand.Rand, n int, lo, hi int64) []Interval {
	out := make([]Interval, n)
	for i := range out {
		out[i] = Random(rng, lo, hi)
	}
	return out
}

// RandomFixedLength generates a random interval of exact length.
func RandomFixedLength(rng *rand.Rand, length, lo, hi int64) Interval {
	if length <= 0 {
		length = 1
	}
	maxStart := hi - length + 1
	if maxStart < lo {
		maxStart = lo
	}
	start := lo
	if maxStart > lo {
		start = lo + rng.Int63n(maxStart-lo)
	}
	return Interval{Start: start, End: start + length - 1}
}

// Shuffle randomly reorders a slice of intervals in place.
func Shuffle(rng *rand.Rand, ivs []Interval) {
	rng.Shuffle(len(ivs), func(i, j int) {
		ivs[i], ivs[j] = ivs[j], ivs[i]
	})
}
