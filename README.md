# interval-tree

`interval-tree` is a small, dependency-free Go library and command line tool for
working with sets of inclusive integer intervals of the form `[start, end]`. It
covers three concerns:

- **Interval arithmetic** (`internal/interval`): overlap and merge tests, length,
  ordering, sorting and parsing of `"start end"` text.
- **An augmented interval search tree** (`internal/tree`): an AVL-balanced binary
  search tree keyed by each interval's start endpoint, augmented with the maximum
  end value stored in every subtree. This lets it answer *stab* (which intervals
  cover a given point) and *overlap* (which intervals overlap a query interval)
  queries in `O(log n + k)` time instead of scanning everything.
- **Coverage analysis** (`internal/coverage`): over a set of intervals, compute the
  size of their union, the uncovered gaps inside a window, a summarized
  per-integer overlap count, and the coverage percentage.

The command line tool reads intervals (one `"start end"` per line, `#` for
comments) from a file and supports the subcommands `count`, `union`, `gaps <min>
<max>`, `overlaps <point>` and `stab <point>`. All non-trivial logic lives in the
internal packages; the CLI only parses flags and prints results.

## Build

```sh
go build ./...
```

The module targets Go 1.21 and uses only the standard library, so no network
access is required to build or test.

## Test

```sh
go test ./...
go vet ./...
```

## Command line usage

```sh
# Count how many intervals are loaded.
interval-tree count -file intervals.txt

# Total number of integers covered by at least one interval.
interval-tree union -file intervals.txt

# Uncovered integer ranges inside [1, 100].
interval-tree gaps 1 100 -file intervals.txt

# Intervals that overlap the point 42.
interval-tree overlaps 42 -file intervals.txt

# Intervals that cover the point 42.
interval-tree stab 42 -file intervals.txt
```

Intervals may also be supplied inline as `start end` token pairs instead of via
`-file`, for example `interval-tree union 1 5 6 10` (which reports `10`).

## Package layout

| Package              | Responsibility                                                |
|----------------------|--------------------------------------------------------------|
| `internal/interval`  | Single-interval type, predicates, merge, sort, parse.        |
| `internal/tree`      | AVL interval tree with `maxEnd` augmentation; stab/overlap.  |
| `internal/coverage`  | Union length, gaps, overlap counts, coverage percentage.     |

## Notes on semantics

Intervals are inclusive on both ends, so the length of `[a, b]` is `b - a + 1`
and two adjacent intervals such as `[1, 5]` and `[6, 10]` together cover every
integer from 1 to 10 with no gap. The tree stores duplicate intervals with a
multiplicity count, and query results reflect that multiplicity.
