package duplicatecheck

import (
	"fmt"
	"math"
	"runtime"
	"sort"
	"testing"
	"time"
)

// QuickBenchResult holds benchmark statistics
type QuickBenchResult struct {
	Name       string
	NumAds     int
	P50        time.Duration
	P95        time.Duration
	P99        time.Duration
	Min        time.Duration
	Max        time.Duration
	MemoryMB   float64
	Throughput int
}

// BenchmarkQuickMatrix runs a fast performance comparison
func BenchmarkQuickMatrix(b *testing.B) {
	b.StopTimer()

	scenarios := []struct {
		name    string
		textLen int
		numAds  int
		iters   int
	}{
		{"100chars_vs_10ads", 100, 10, 20},
		{"100chars_vs_100ads", 100, 100, 15},
		{"100chars_vs_1000ads", 100, 1000, 10},
		{"500chars_vs_10ads", 500, 10, 15},
		{"500chars_vs_100ads", 500, 100, 10},
		{"500chars_vs_1000ads", 500, 1000, 5},
		{"1000chars_vs_10ads", 1000, 10, 10},
		{"1000chars_vs_100ads", 1000, 100, 5},
		{"1000chars_vs_1000ads", 1000, 1000, 3},
	}

	results := make([]QuickBenchResult, len(scenarios))

	fmt.Println("\n🚀 Running Quick Performance Matrix...")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	for i, scenario := range scenarios {
		fmt.Printf("  [%d/%d] %s... ", i+1, len(scenarios), scenario.name)
		results[i] = runQuickBench(scenario.name, scenario.textLen, scenario.numAds, scenario.iters)
		fmt.Printf("✓ P50=%s\n", formatQuickDur(results[i].P50))
	}

	// Print results
	fmt.Println("\n\n=== SUMMARY ===")
	fmt.Println()
	fmt.Printf("%-25s | %8s | %11s | %11s | %11s | %11s | %11s | %10s\n",
		"Test", "Num Ads", "P50 (med)", "P95", "P99", "Min", "Max", "Memory")
	fmt.Println("------------------------------------------------------------------------------------------------------------------------------")

	for _, r := range results {
		fmt.Printf("%-25s | %8d | %11s | %11s | %11s | %11s | %11s | %8.2fMB\n",
			r.Name, r.NumAds,
			formatQuickDur(r.P50), formatQuickDur(r.P95), formatQuickDur(r.P99),
			formatQuickDur(r.Min), formatQuickDur(r.Max), r.MemoryMB)
	}

	fmt.Println("\n=== ANALYSIS ===")
	fmt.Println()

	for _, r := range results {
		status := getQuickStatus(r.P50)
		fmt.Printf("%-25s: P50=%s, P95=%s, P99=%s, Throughput=%d/sec, Status=%s\n",
			r.Name, formatQuickDur(r.P50), formatQuickDur(r.P95), formatQuickDur(r.P99),
			r.Throughput, status)
	}

	fmt.Println("\n=== ENGINE USAGE ===")
	fmt.Println()
	fmt.Println("  • Configurations with <100 ads:  Levenshtein Engine (Naive)")
	fmt.Println("  • Configurations with ≥100 ads:  Hybrid Engine (MinHash+LSH)")
	fmt.Println()
}

func runQuickBench(name string, textLen, numAds, iters int) QuickBenchResult {
	// Generate test data
	baseText := "Apple iPhone 14 Pro Max with A16 Bionic chip, ProMotion display, and 48MP camera. "
	fullText := ""
	for len(fullText) < textLen {
		fullText += baseText
	}
	fullText = fullText[:textLen]

	query := Product{
		ID:          "QUERY",
		Name:        "Test Product",
		Description: fullText,
	}

	catalog := make([]Product, numAds)
	for i := 0; i < numAds; i++ {
		catalog[i] = Product{
			ID:          fmt.Sprintf("CAT_%d", i),
			Name:        fmt.Sprintf("Product %d", i),
			Description: fullText,
		}
	}

	var durations []time.Duration
	var memStats runtime.MemStats

	if numAds >= 100 {
		// Hybrid engine - now kicks in at 100+ products instead of 500
		engine := NewHybridEngine()
		engine.BuildIndex(catalog)

		runtime.ReadMemStats(&memStats)
		startMem := memStats.TotalAlloc

		for i := 0; i < iters; i++ {
			start := time.Now()
			engine.FindDuplicatesForOne(query, 0.85)
			durations = append(durations, time.Since(start))
		}

		runtime.ReadMemStats(&memStats)
		memMB := float64(memStats.TotalAlloc-startMem) / 1024 / 1024

		return calcQuickStats(name, numAds, durations, memMB)
	} else {
		// Levenshtein engine
		engine := NewLevenshteinEngine()

		runtime.ReadMemStats(&memStats)
		startMem := memStats.TotalAlloc

		for i := 0; i < iters; i++ {
			start := time.Now()
			for _, cat := range catalog {
				engine.Compare(query, cat)
			}
			durations = append(durations, time.Since(start))
		}

		runtime.ReadMemStats(&memStats)
		memMB := float64(memStats.TotalAlloc-startMem) / 1024 / 1024

		return calcQuickStats(name, numAds, durations, memMB)
	}
}

func calcQuickStats(name string, numAds int, durations []time.Duration, memMB float64) QuickBenchResult {
	if len(durations) == 0 {
		return QuickBenchResult{Name: name, NumAds: numAds}
	}

	sort.Slice(durations, func(i, j int) bool {
		return durations[i] < durations[j]
	})

	n := len(durations)
	p50 := durations[n*50/100]
	p95 := durations[int(math.Min(float64(n*95/100), float64(n-1)))]
	p99 := durations[int(math.Min(float64(n*99/100), float64(n-1)))]

	throughput := 0
	if p50 > 0 {
		throughput = int(float64(time.Second) / float64(p50))
	}

	return QuickBenchResult{
		Name:       name,
		NumAds:     numAds,
		P50:        p50,
		P95:        p95,
		P99:        p99,
		Min:        durations[0],
		Max:        durations[n-1],
		MemoryMB:   memMB,
		Throughput: throughput,
	}
}

func formatQuickDur(d time.Duration) string {
	switch {
	case d < time.Microsecond:
		return fmt.Sprintf("%dns", d.Nanoseconds())
	case d < time.Millisecond:
		return fmt.Sprintf("%.2fµs", float64(d.Nanoseconds())/1000)
	case d < time.Second:
		return fmt.Sprintf("%.6fms", float64(d.Nanoseconds())/1000000)
	default:
		return fmt.Sprintf("%.9fs", d.Seconds())
	}
}

func getQuickStatus(p50 time.Duration) string {
	switch {
	case p50 < 1*time.Millisecond:
		return "✅ Excellent"
	case p50 < 50*time.Millisecond:
		return "✅ Good"
	case p50 < 500*time.Millisecond:
		return "⚠️ Slow"
	case p50 < 1*time.Second:
		return "❌ Very Slow"
	default:
		return "💀 Critical"
	}
}
