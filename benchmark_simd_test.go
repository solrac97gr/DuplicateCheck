package duplicatecheck

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

// TestSIMDBenchmarkComparison runs detailed performance comparisons
func TestSIMDBenchmarkComparison(t *testing.T) {
	config := DefaultSIMDConfig()
	config.Enabled = false // Test scalar path

	type testCase struct {
		name        string
		lengthA     int
		lengthB     int
		iterations  int
		description string
	}

	testCases := []testCase{
		{"Very short strings", 5, 5, 10000, "apple vs apple"},
		{"Short strings", 20, 20, 5000, "iPhone 13 Pro vs iPhone 13 Pro"},
		{"Medium strings", 100, 100, 1000, "Product with ~100 chars"},
		{"Long strings", 500, 500, 100, "Product desc with ~500 chars"},
		{"Very long strings", 1000, 1000, 50, "Full product desc ~1000 chars"},
	}

	// Helper to create test strings of given length
	createTestString := func(length int) string {
		pattern := "The quick brown fox jumps over the lazy dog. "
		if length == 0 {
			return ""
		}
		full := strings.Repeat(pattern, (length/len(pattern))+1)
		return full[:length]
	}

	t.Logf("\n=== SIMD Performance Benchmark ===\n")
	t.Logf("%-25s | %-10s | %-10s | %-12s | %-15s | %-15s\n",
		"Test Case", "Length", "Iterations", "Time (ms)", "Time/Op (ns)", "Throughput")
	t.Logf("%s\n", strings.Repeat("-", 110))

	for _, tc := range testCases {
		s := createTestString(tc.lengthA)
		testStr := createTestString(tc.lengthB)

		start := time.Now()
		for i := 0; i < tc.iterations; i++ {
			_ = ComputeDistanceOptimized(s, testStr, config)
		}
		elapsed := time.Since(start)

		totalNs := elapsed.Nanoseconds()
		perOpNs := totalNs / int64(tc.iterations)
		totalMs := float64(totalNs) / 1_000_000
		throughput := float64(tc.iterations) / elapsed.Seconds()

		t.Logf("%-25s | %-10d | %-10d | %-12.2f | %-15d | %-15.0f ops/sec\n",
			tc.name, tc.lengthA, tc.iterations, totalMs, perOpNs, throughput)
	}

	t.Logf("\n")
}

// TestBenchmarkDetailedOutput provides detailed performance analysis
func TestBenchmarkDetailedOutput(t *testing.T) {
	t.Logf("\n=== Detailed Levenshtein Performance Analysis ===\n")

	// Test 1: Short strings (typical product names)
	t.Logf("SHORT STRINGS (Product Names):\n")
	shortTests := map[string]struct {
		s1, s2 string
	}{
		"Identical": {"iPhone 13", "iPhone 13"},
		"1 diff": {"iPhone 13", "iPhone 12"},
		"2 diffs": {"iPhone 13", "Samsung 21"},
		"Completely different": {"Apple", "Sony"},
	}

	for name, test := range shortTests {
		start := time.Now()
		dist := levenshteinDistanceScalar(test.s1, test.s2)
		elapsed := time.Since(start)
		t.Logf("  %-25s: distance=%2d, time=%8d ns\n", name, dist, elapsed.Nanoseconds())
	}

	// Test 2: Medium strings (typical product descriptions start)
	t.Logf("\nMEDIUM STRINGS (100 chars):\n")
	mediumStr := "Apple iPhone 13 Pro with A15 Bionic chip, 6.1-inch display, advanced camera system"
	mediumStr2 := "Apple iPhone 14 Pro with A16 Bionic chip, 6.1-inch display, advanced camera system"

	for i := 0; i < 3; i++ {
		start := time.Now()
		dist := levenshteinDistanceScalar(mediumStr, mediumStr2)
		elapsed := time.Since(start)
		t.Logf("  Iteration %d: distance=%2d, time=%6d ns\n", i+1, dist, elapsed.Nanoseconds())
	}

	// Test 3: Long strings (full product descriptions)
	t.Logf("\nLONG STRINGS (500+ chars):\n")
	longStr := strings.Repeat("This is a longer product description with multiple sentences. ", 10)
	longStr2 := strings.Repeat("This is a longer product description with multiple sentences. ", 10)
	// Make a small change
	runes := []rune(longStr2)
	if len(runes) > 50 {
		runes[50] = 'X'
	}
	longStr2 = string(runes)

	for i := 0; i < 3; i++ {
		start := time.Now()
		dist := levenshteinDistanceScalar(longStr, longStr2)
		elapsed := time.Since(start)
		t.Logf("  Iteration %d: distance=%2d, time=%6d ns\n", i+1, dist, elapsed.Nanoseconds())
	}

	t.Logf("\nNote: Times may vary due to CPU frequency scaling and system load\n")
}

// TestSIMDMemoryProfile checks memory usage patterns
func TestSIMDMemoryProfile(t *testing.T) {
	config := DefaultSIMDConfig()

	sizes := []int{10, 50, 100, 500, 1000}

	t.Logf("\nMemory Usage vs String Length:\n")
	t.Logf("%-10s | %-20s | %-20s\n", "Length", "Time/Op (ns)", "Iterations")
	t.Logf("%s\n", strings.Repeat("-", 55))

	for _, size := range sizes {
		testStr := strings.Repeat("x", size)

		// Measure scalar version
		start := time.Now()
		iterations := 100
		for i := 0; i < iterations; i++ {
			_ = ComputeDistanceOptimized(testStr, testStr, config)
		}
		elapsed := time.Since(start)
		timePerOp := elapsed.Nanoseconds() / int64(iterations)

		t.Logf("%-10d | %-20d | %-20d\n", size, timePerOp, iterations)
	}
}

// BenchmarkProductNameComparison simulates real e-commerce product comparisons
func BenchmarkProductNameComparison(b *testing.B) {
	config := DefaultSIMDConfig()
	config.Enabled = false

	productPairs := []struct {
		name string
		s    string
		t    string
	}{
		{"Exact match", "Apple iPhone 13 Pro Max", "Apple iPhone 13 Pro Max"},
		{"One char diff", "Apple iPhone 13 Pro Max", "Apple iPhone 13 Pro Max"},
		{"Brand + Model", "Samsung Galaxy S21 Ultra", "Samsung Galaxy S22 Ultra"},
		{"Different brands", "iPhone 13", "Samsung Galaxy S21"},
		{"Long desc",
			"High-performance laptop with Intel i7, 16GB RAM, 512GB SSD, 15.6 inch display",
			"High-performance laptop with Intel i7, 16GB RAM, 512GB SSD, 15.6 inch display"},
	}

	for _, pair := range productPairs {
		b.Run(pair.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ComputeDistanceOptimized(pair.s, pair.t, config)
			}
			b.ReportAllocs()
		})
	}
}

// BenchmarkDifferentLengths shows performance across various string lengths
func BenchmarkDifferentLengths(b *testing.B) {
	config := DefaultSIMDConfig()
	config.Enabled = false

	testLengths := []int{10, 50, 100, 200, 500, 1000}

	for _, length := range testLengths {
		pattern := "The quick brown fox jumps over the lazy dog. "
		s := strings.Repeat(pattern, (length/len(pattern))+1)[:length]
		testStr := strings.Repeat(pattern, (length/len(pattern))+1)[:length]

		b.Run(fmt.Sprintf("Length_%d", length), func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = ComputeDistanceOptimized(s, testStr, config)
			}
			b.ReportAllocs()
		})
	}
}
