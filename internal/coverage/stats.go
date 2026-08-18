package coverage

// Percent computes coverage as a percentage given covered and total points.
func Percent(covered, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total) * 100
}

// Ratio computes coverage as a ratio [0, 1].
func Ratio(covered, total int64) float64 {
	if total == 0 {
		return 0
	}
	return float64(covered) / float64(total)
}

// Grade returns a letter grade based on coverage percentage.
func Grade(pct float64) string {
	switch {
	case pct >= 90:
		return "A"
	case pct >= 80:
		return "B"
	case pct >= 70:
		return "C"
	case pct >= 60:
		return "D"
	default:
		return "F"
	}
}

// Gap computes the uncovered points given covered and total.
func Gap(covered, total int64) int64 {
	if covered > total {
		return 0
	}
	return total - covered
}
