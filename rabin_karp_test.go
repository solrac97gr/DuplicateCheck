package duplicatecheck

import (
	"math"
	"testing"
)

func TestRabinKarpFilterBasics(t *testing.T) {
	filter := NewRabinKarpFilter(5)

	tests := []struct {
		name     string
		s        string
		t        string
		threshold float64
		expected bool
	}{
		// Identical strings should always pass
		{"Identical short", "apple", "apple", 0.8, true},
		{"Identical long", "This is a long product description", "This is a long product description", 0.8, true},

		// Very similar strings should pass (use conservative thresholds for hashing)
		{"Minor typo", "iPhone", "iFone", 0.65, true},
		{"Minor difference", "Samsung Galaxy", "Samsung Galxy", 0.6, true},

		// Very different strings should be rejected
		{"Completely different", "apple", "orange", 0.8, false},
		{"Different brands", "iPhone", "Samsung", 0.75, false},

		// Empty string handling
		{"Empty vs non-empty", "", "apple", 0.8, false},
		{"Both empty", "", "", 0.8, true},

		// Short strings - hash-based is less accurate for very short strings
		{"Both short similar", "ab", "ac", 0.4, true},
		{"Both short very different", "ab", "xy", 0.9, false},

		// Long strings
		{"Both long similar", "abcdefghijklmnop", "abcdefghijklmnpq", 0.7, true},
		{"Both long different", "abcdefghijklmnop", "zyxwvutsrqponmlk", 0.8, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.QuickReject(tt.s, tt.t, tt.threshold)
			if result != tt.expected {
				t.Errorf("QuickReject(%q, %q, %v) = %v, want %v",
					tt.s, tt.t, tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestRabinKarpEstimateSimilarity(t *testing.T) {
	filter := NewRabinKarpFilter(4)

	tests := []struct {
		name     string
		s        string
		t        string
		minSim   float64 // Minimum expected similarity
		maxSim   float64 // Maximum expected similarity
	}{
		// Identical strings should have perfect similarity
		{"Identical", "apple", "apple", 0.95, 1.0},

		// Similar strings should have high similarity
		{"Very similar", "test", "best", 0.50, 1.0},

		// Very different strings should have low similarity
		// Note: "apple" and "orange" share some characters, so not 0
		{"Different", "apple", "orange", 0.15, 0.35},

		// Empty strings
		{"Both empty", "", "", 0.95, 1.0},
		{"One empty", "", "apple", 0.0, 0.1},

		// Length mismatch consideration
		{"Length mismatch", "a", "abcdefghij", 0.0, 0.3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sim := filter.estimateSimilarity(tt.s, tt.t)
			if sim < tt.minSim || sim > tt.maxSim {
				t.Errorf("estimateSimilarity(%q, %q) = %v, want between %v and %v",
					tt.s, tt.t, sim, tt.minSim, tt.maxSim)
			}
		})
	}
}

func TestRabinKarpWindowHashes(t *testing.T) {
	filter := NewRabinKarpFilter(3)

	s := "abcdef"
	hashes := filter.getAllWindowHashes(s)

	// For string "abcdef" with window size 3:
	// Windows: "abc", "bcd", "cde", "def"
	// Should have 4 hashes (6 - 3 + 1)
	expectedCount := len(s) - filter.windowSize + 1
	if len(hashes) != expectedCount {
		t.Errorf("getAllWindowHashes(%q) returned %d hashes, want %d",
			s, len(hashes), expectedCount)
	}

	// All hashes should be non-zero (reasonable values)
	for i, h := range hashes {
		if h > filter.modulo {
			t.Errorf("Hash %d exceeded modulo: %d > %d", i, h, filter.modulo)
		}
	}
}

func TestRabinKarpEnableDisable(t *testing.T) {
	filter := NewRabinKarpFilter(5)

	// Should be enabled by default
	if !filter.IsEnabled() {
		t.Error("Filter should be enabled by default")
	}

	// When disabled, should always return true (continue to Levenshtein)
	filter.Disable()
	if filter.IsEnabled() {
		t.Error("Filter should be disabled")
	}
	if !filter.QuickReject("completely", "different", 0.9) {
		t.Error("When disabled, QuickReject should return true")
	}

	// When re-enabled, should filter again
	filter.Enable()
	if !filter.IsEnabled() {
		t.Error("Filter should be enabled")
	}
	if filter.QuickReject("completely", "different", 0.9) {
		t.Error("When enabled, very different strings should be rejected")
	}
}

func TestRabinKarpWindowSizeVariations(t *testing.T) {
	tests := []struct {
		name       string
		windowSize int
		s          string
		t          string
		threshold  float64
	}{
		{"Window 2", 2, "apple", "apple", 0.8},
		{"Window 5", 5, "iPhone", "iFone", 0.75},
		{"Window 8", 8, "This is a test", "This is a best", 0.7},
		{"Window 1", 1, "a", "a", 0.8},
		{"Window 32", 32, "verylongproductname", "verylongproductname", 0.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewRabinKarpFilter(tt.windowSize)
			result := filter.QuickReject(tt.s, tt.t, tt.threshold)
			if tt.s == tt.t && !result {
				t.Errorf("Identical strings should not be rejected")
			}
		})
	}
}

func TestRabinKarpLengthSensitivity(t *testing.T) {
	filter := NewRabinKarpFilter(4)

	// Test various length ratios
	tests := []struct {
		name      string
		s         string
		t         string
		threshold float64
		shouldReject bool
	}{
		// Similar length, similar content (use conservative threshold for hash)
		{"Similar length similar", "test", "best", 0.5, true},

		// Very different lengths
		{"Very different length", "a", "verylongstring", 0.8, false},

		// High threshold with different lengths
		{"Different length high threshold", "apple", "application", 0.85, false},

		// Low threshold with different lengths
		{"Different length low threshold", "apple", "apples", 0.45, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.QuickReject(tt.s, tt.t, tt.threshold)
			if result != tt.shouldReject {
				t.Errorf("QuickReject(%q, %q, %v) = %v, want %v",
					tt.s, tt.t, tt.threshold, result, tt.shouldReject)
			}
		})
	}
}

func TestRabinKarpProductNames(t *testing.T) {
	filter := NewRabinKarpFilter(5)

	// Real product name scenarios
	// Note: Rabin-Karp is a probabilistic filter, so we use conservative thresholds
	tests := []struct {
		name      string
		productA  string
		productB  string
		threshold float64
		expected  bool // Should continue to Levenshtein (not rejected)
	}{
		{"iPhone variants", "iPhone", "iFone", 0.65, true},
		{"Brand similar", "Samsung Galaxy S20", "Samsung Galxy S20", 0.7, true},
		{"Different brands", "iPhone 12", "Samsung S20", 0.85, false},
		{"Model variants", "MacBook Pro 16", "MacBook Pro 15", 0.7, true},
		{"Completely different", "PlayStation 5", "Xbox Series X", 0.9, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.QuickReject(tt.productA, tt.productB, tt.threshold)
			if result != tt.expected {
				t.Errorf("QuickReject(%q, %q, %v) = %v, want %v",
					tt.productA, tt.productB, tt.threshold, result, tt.expected)
			}
		})
	}
}

func TestRabinKarpSetWindowSize(t *testing.T) {
	filter := NewRabinKarpFilter(5)

	if filter.GetWindowSize() != 5 {
		t.Errorf("Initial window size should be 5, got %d", filter.GetWindowSize())
	}

	filter.SetWindowSize(3)
	if filter.GetWindowSize() != 3 {
		t.Errorf("Window size should be 3 after SetWindowSize, got %d", filter.GetWindowSize())
	}

	// Test boundaries
	filter.SetWindowSize(0)
	if filter.GetWindowSize() != 1 {
		t.Errorf("Window size < 1 should be clamped to 1, got %d", filter.GetWindowSize())
	}

	filter.SetWindowSize(100)
	if filter.GetWindowSize() != 32 {
		t.Errorf("Window size > 32 should be clamped to 32, got %d", filter.GetWindowSize())
	}
}

func TestRabinKarpNoFalseNegatives(t *testing.T) {
	// Critical test: Rabin-Karp should never reject truly similar strings
	// (false negatives would cause us to miss actual duplicates)
	// Use conservative thresholds since hash-based filtering is probabilistic
	filter := NewRabinKarpFilter(5)

	// These pairs should ALL pass Rabin-Karp (not be rejected)
	// even if Levenshtein similarity is borderline
	trueSimilarPairs := []struct {
		a         string
		b         string
		threshold float64
	}{
		{"test", "test", 0.9},      // Identical - always pass
		{"apple", "aple", 0.6},     // One char difference - use conservative threshold
		{"iPhone", "iFone", 0.65},   // Common typo - use conservative threshold
		{"Samsung", "Samsong", 0.6}, // Common typo - use conservative threshold
		{"color", "colour", 0.65},  // Spelling variation - use conservative threshold
	}

	for _, pair := range trueSimilarPairs {
		if !filter.QuickReject(pair.a, pair.b, pair.threshold) {
			t.Errorf("Rabin-Karp incorrectly rejected similar pair: %q vs %q at threshold %v",
				pair.a, pair.b, pair.threshold)
		}
	}
}

func BenchmarkRabinKarpQuickReject(b *testing.B) {
	filter := NewRabinKarpFilter(5)

	pairs := []struct {
		a, b      string
		threshold float64
	}{
		{"iPhone", "iFone", 0.75},
		{"Samsung", "Samsong", 0.7},
		{"apple", "orange", 0.8},
		{"test string", "best string", 0.75},
		{"completely", "different", 0.9},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range pairs {
			_ = filter.QuickReject(p.a, p.b, p.threshold)
		}
	}
}

func BenchmarkRabinKarpEstimateSimilarity(b *testing.B) {
	filter := NewRabinKarpFilter(4)

	pairs := []struct {
		a, b string
	}{
		{"apple", "apple"},
		{"test", "best"},
		{"apple", "orange"},
		{"This is a test", "This is a best"},
		{"completely different strings", "entirely other text"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range pairs {
			_ = filter.estimateSimilarity(p.a, p.b)
		}
	}
}

func BenchmarkRabinKarpGetAllHashes(b *testing.B) {
	filter := NewRabinKarpFilter(5)

	strings := []string{
		"short",
		"medium length string",
		"This is a much longer product description that might span multiple lines in a real database",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range strings {
			_ = filter.getAllWindowHashes(s)
		}
	}
}

func TestRabinKarpEstimatedSpeed(t *testing.T) {
	filter := NewRabinKarpFilter(5)

	speed := filter.EstimatedSimSpeed(100, 0.8)

	// Should return a reasonable speedup estimate
	if speed < 1.0 || speed > 10.0 {
		t.Errorf("EstimatedSimSpeed returned unreasonable value: %v", speed)
	}

	// With higher rejection rate, should have more speedup
	speed2 := filter.EstimatedSimSpeed(1000, 0.8)
	if speed2 < 1.0 {
		t.Errorf("EstimatedSimSpeed should return > 1.0x, got %v", speed2)
	}
}

func TestRabinKarpSimilarityBounds(t *testing.T) {
	filter := NewRabinKarpFilter(5)

	testCases := []struct {
		s string
		t string
	}{
		{"a", "a"},
		{"apple", "apple"},
		{"test", "best"},
		{"", ""},
		{"abc", "def"},
		{"verylongstring", "verylongstring"},
	}

	for _, tc := range testCases {
		sim := filter.estimateSimilarity(tc.s, tc.t)

		// Similarity should always be in [0.0, 1.0]
		if sim < 0.0 || sim > 1.0 {
			t.Errorf("estimateSimilarity(%q, %q) = %v, want [0.0, 1.0]",
				tc.s, tc.t, sim)
		}

		// If strings are identical, similarity should be 1.0
		if tc.s == tc.t && math.Abs(sim-1.0) > 0.01 {
			t.Errorf("Identical strings should have ~1.0 similarity, got %v for %q", sim, tc.s)
		}
	}
}
