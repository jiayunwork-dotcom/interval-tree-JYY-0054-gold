// Command interval-tree is a small command line tool built on top of the
// interval, tree and coverage packages. It loads a set of inclusive integer
// intervals (one "start end" per line) and answers a handful of questions
// about them: how many there are, the size of their union, the uncovered
// ranges inside a window, and which intervals cover or overlap a point.
//
// All of the real logic lives in the internal packages; this file only
// parses flags, reads the input file and prints results.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"interval-tree/internal/coverage"
	"interval-tree/internal/interval"
	"interval-tree/internal/tree"
)

// reorder moves every flag token (and, for non boolean flags, its immediate
// value) ahead of the positional tokens. The standard library flag package
// stops parsing at the first positional argument, so a subcommand form such
// as `interval-tree stab 5 -file data.txt` would otherwise never see the
// -file flag. Reordering to `-file data.txt stab 5` lets fs.Parse consume
// the flag first and leave "5" as a positional argument.
func reorder(args []string) []string {
	var flags, pos []string
	for i := 0; i < len(args); i++ {
		a := args[i]
		if strings.HasPrefix(a, "-") && a != "-" && a != "--" {
			flags = append(flags, a)
			if strings.IndexByte(a, '=') >= 0 {
				// Value attached via '=', nothing extra to grab.
				continue
			}
			// Grab the next token as the flag's value unless it is itself a flag.
			if i+1 < len(args) && !strings.HasPrefix(args[i+1], "-") {
				flags = append(flags, args[i+1])
				i++
			}
			continue
		}
		pos = append(pos, a)
	}
	return append(flags, pos...)
}

// loadIntervals reads intervals from path. When path is "-" or empty the
// caller must have supplied the intervals inline via extra, which is treated
// as a flat list of "start end" tokens (used for quick one off queries).
func loadIntervals(path string, extra []string) ([]interval.Interval, error) {
	if path != "" && path != "-" {
		f, err := os.Open(path)
		if err != nil {
			return nil, fmt.Errorf("open %q: %w", path, err)
		}
		defer f.Close()
		return readIntervals(f)
	}
	if len(extra) == 0 {
		return nil, errors.New("no intervals provided (use -file or pass \"start end\" tokens)")
	}
	if len(extra)%2 != 0 {
		return nil, errors.New("inline intervals must come in start/end pairs")
	}
	out := make([]interval.Interval, 0, len(extra)/2)
	for i := 0; i < len(extra); i += 2 {
		s, err := strconv.ParseInt(extra[i], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad start %q: %w", extra[i], err)
		}
		e, err := strconv.ParseInt(extra[i+1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("bad end %q: %w", extra[i+1], err)
		}
		iv := interval.Interval{Start: s, End: e}
		if !iv.Valid() {
			return nil, fmt.Errorf("invalid interval %v", iv)
		}
		out = append(out, iv)
	}
	return out, nil
}

// readIntervals parses "start end" lines, skipping blank lines and lines
// whose first non blank character is '#'.
func readIntervals(r io.Reader) ([]interval.Interval, error) {
	var out []interval.Interval
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	line := 0
	for sc.Scan() {
		line++
		text := strings.TrimSpace(sc.Text())
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		iv, err := interval.Parse(text)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", line, err)
		}
		out = append(out, iv)
	}
	if err := sc.Err(); err != nil {
		return nil, fmt.Errorf("reading intervals: %w", err)
	}
	return out, nil
}

// run implements the command logic. It returns an error for any user facing
// problem; main is the only place that converts that into an exit code.
func run(args []string) error {
	if len(args) == 0 {
		return errors.New("usage: interval-tree <count|union|gaps|overlaps|stab> [-file PATH] [args]")
	}
	cmd := args[0]
	rest := reorder(args[1:])

	fs := flag.NewFlagSet(cmd, flag.ContinueOnError)
	file := fs.String("file", "", "path to intervals file (one \"start end\" per line); use \"-\" for stdin")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	pos := fs.Args()

	ivs, err := loadIntervals(*file, pos)
	if err != nil {
		return err
	}

	switch cmd {
	case "count":
		fmt.Println(len(ivs))
		return nil

	case "union":
		a, err := coverage.New(ivs)
		if err != nil {
			return err
		}
		fmt.Println(a.UnionLength())
		return nil

	case "gaps":
		if len(pos) != 2 {
			return errors.New("gaps requires two arguments: <min> <max>")
		}
		min, err := strconv.ParseInt(pos[0], 10, 64)
		if err != nil {
			return fmt.Errorf("bad min %q: %w", pos[0], err)
		}
		max, err := strconv.ParseInt(pos[1], 10, 64)
		if err != nil {
			return fmt.Errorf("bad max %q: %w", pos[1], err)
		}
		a, err := coverage.New(ivs)
		if err != nil {
			return err
		}
		for _, g := range a.Gaps(min, max) {
			fmt.Println(g.String())
		}
		return nil

	case "overlaps":
		if len(pos) != 1 {
			return errors.New("overlaps requires one argument: <point>")
		}
		point, err := strconv.ParseInt(pos[0], 10, 64)
		if err != nil {
			return fmt.Errorf("bad point %q: %w", pos[0], err)
		}
		var tr tree.Tree
		for _, iv := range ivs {
			if err := tr.Insert(iv); err != nil {
				return err
			}
		}
		for _, iv := range tr.Overlapping(interval.Interval{Start: point, End: point}) {
			fmt.Println(iv.String())
		}
		return nil

	case "stab":
		if len(pos) != 1 {
			return errors.New("stab requires one argument: <point>")
		}
		point, err := strconv.ParseInt(pos[0], 10, 64)
		if err != nil {
			return fmt.Errorf("bad point %q: %w", pos[0], err)
		}
		var tr tree.Tree
		for _, iv := range ivs {
			if err := tr.Insert(iv); err != nil {
				return err
			}
		}
		for _, iv := range tr.Stab(point) {
			fmt.Println(iv.String())
		}
		return nil

	default:
		return fmt.Errorf("unknown command %q", cmd)
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "interval-tree:", err)
		os.Exit(1)
	}
}
