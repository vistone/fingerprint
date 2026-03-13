//go:build ignore

package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// BenchmarkResult represents a single benchmark result
type BenchmarkResult struct {
	Name        string
	Iterations  int64
	TimeNS      float64 // ns/op
	MemoryB     float64 // B/op
	Allocations float64 // allocs/op
}

func main() {
	baseFile := flag.String("base", "benchmark_baseline.txt", "Baseline benchmark file")
	currentFile := flag.String("current", "benchmark_current.txt", "Current benchmark file")
	threshold := flag.Float64("threshold", 10.0, "Performance regression threshold in percent")
	flag.Parse()

	baseResults := parseBenchmarkFile(*baseFile)
	currentResults := parseBenchmarkFile(*currentFile)

	fmt.Println("=== Benchmark Comparison ===")
	fmt.Printf("%-40s %12s %12s %12s\n", "Benchmark", "Base", "Current", "Change")
	fmt.Println(strings.Repeat("-", 80))

	regressions := 0

	for name, base := range baseResults {
		current, ok := currentResults[name]
		if !ok {
			fmt.Printf("%-40s %s\n", name, "MISSING")
			continue
		}

		// Compare time per operation
		change := ((current.TimeNS - base.TimeNS) / base.TimeNS) * 100
		status := "OK"
		if change > *threshold {
			status = "REGRESSION"
			regressions++
		} else if change < -*threshold {
			status = "IMPROVED"
		}

		fmt.Printf("%-40s %10.2fns %10.2fns %+9.2f%% %s\n",
			name, base.TimeNS, current.TimeNS, change, status)
	}

	fmt.Println(strings.Repeat("-", 80))
	if regressions > 0 {
		fmt.Printf("\n❌ Found %d performance regressions\n", regressions)
		os.Exit(1)
	}
	fmt.Println("\n✅ No significant performance regressions")
}

func parseBenchmarkFile(filename string) map[string]BenchmarkResult {
	results := make(map[string]BenchmarkResult)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	// Benchmark pattern: BenchmarkName-N    123456    123.4 ns/op    56 B/op    2 allocs/op
	// Using strict patterns for parsing
	benchPattern := regexp.MustCompile(`^(Benchmark\S+)\s+(\d+)\s+([\d.]+)\s+ns/op`)
	memPattern := regexp.MustCompile(`([\d.]+)\s+B/op`)
	allocPattern := regexp.MustCompile(`([\d.]+)\s+allocs/op`)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		matches := benchPattern.FindStringSubmatch(line)
		if matches == nil {
			continue
		}

		name := matches[1]
		iterations, _ := strconv.ParseInt(matches[2], 10, 64)
		timeNS, _ := strconv.ParseFloat(matches[3], 64)

		result := BenchmarkResult{
			Name:       name,
			Iterations: iterations,
			TimeNS:     timeNS,
		}

		if memMatches := memPattern.FindStringSubmatch(line); memMatches != nil {
			result.MemoryB, _ = strconv.ParseFloat(memMatches[1], 64)
		}

		if allocMatches := allocPattern.FindStringSubmatch(line); allocMatches != nil {
			result.Allocations, _ = strconv.ParseFloat(allocMatches[1], 64)
		}

		results[name] = result
	}

	return results
}
