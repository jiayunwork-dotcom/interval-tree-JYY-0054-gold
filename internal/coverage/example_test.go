package coverage

import (
	"fmt"

	"interval-tree/internal/interval"
)

// ExampleAnalyzer_Gaps shows the uncovered ranges inside a window.
func ExampleAnalyzer_Gaps() {
	a, _ := New([]interval.Interval{
		{Start: 1, End: 5},
		{Start: 7, End: 10},
	})
	for _, g := range a.Gaps(1, 10) {
		fmt.Println(g)
	}
	// Output: 6 6
}

// ExampleAnalyzer_UnionLength shows the size of the union.
func ExampleAnalyzer_UnionLength() {
	a, _ := New([]interval.Interval{
		{Start: 1, End: 5},
		{Start: 6, End: 10},
	})
	fmt.Println(a.UnionLength())
	// Output: 10
}
