package duplicatecheck

import (
	"strings"
	"testing"
)

func TestSIMDConfigDefaults(t *testing.T) {
	config := DefaultSIMDConfig()

	if config.Enabled {
		t.Errorf("SIMD should be disabled by default for compatibility")
	}

	if config.MinStringLength < 50 || config.MinStringLength > 200 {
		t.Errorf("MinStringLength %d should be reasonable (50-200)", config.MinStringLength)
	}

	if config.Architecture == "" {
		t.Errorf("Architecture should be detected")
	}
}

func TestComputeDistanceOptimizedDisabled(t *testing.T) {
	config := DefaultSIMDConfig()
	config.Enabled = false

	tests := []struct {
		name     string
		s        string
		t        string
		expected int
	}{
		{"Identical", "apple", "apple", 0},
		{"One char diff", "apple", "aple", 1},
		{"Completely different", "abc", "xyz", 3},
		{"Empty strings", "", "", 0},
		{"One empty", "hello", "", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeDistanceOptimized(tt.s, tt.t, config)
			if result != tt.expected {
				t.Errorf("ComputeDistanceOptimized(%q, %q) = %d, want %d",
					tt.s, tt.t, result, tt.expected)
			}
		})
	}
}

func TestComputeDistanceOptimizedLongStrings(t *testing.T) {
	config := DefaultSIMDConfig()
	config.Enabled = false // Disable SIMD for this test - test the scalar path
	config.MinStringLength = 10

	// Test with strings longer than MinStringLength
	tests := []struct {
		name     string
		s        string
		t        string
		expected int
	}{
		{"Long identical", "a long product name", "a long product name", 0},
		{"Long similar", "apple iphone", "apple iphon", 1},
		{"Long different", "Samsung Galaxy S21", "Google Pixel 6", 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ComputeDistanceOptimized(tt.s, tt.t, config)
			if result != tt.expected {
				t.Errorf("ComputeDistanceOptimized(%q, %q) = %d, want %d",
					tt.s, tt.t, result, tt.expected)
			}
		})
	}
}

func TestLevenshteinDistanceScalar(t *testing.T) {
	tests := []struct {
		name     string
		s        string
		t        string
		expected int
	}{
		{"Identical", "test", "test", 0},
		{"Empty vs non-empty", "", "abc", 3},
		{"Both empty", "", "", 0},
		{"One char", "a", "b", 1},
		{"Insertion", "abc", "abbc", 1},
		{"Deletion", "abbc", "abc", 1},
		{"Substitution", "cat", "car", 1},
		{"Multiple edits", "kitten", "sitting", 3},
		{"Transposition", "ab", "ba", 2},
		{"Case sensitivity", "ABC", "abc", 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := levenshteinDistanceScalar(tt.s, tt.t)
			if result != tt.expected {
				t.Errorf("levenshteinDistanceScalar(%q, %q) = %d, want %d",
					tt.s, tt.t, result, tt.expected)
			}
		})
	}
}

func TestScalarVsOptimized(t *testing.T) {
	// Verify that scalar and optimized versions produce identical results
	testCases := []struct {
		s string
		t string
	}{
		{"apple", "aple"},
		{"Samsung Galaxy", "Samsung Galxy"},
		{"iPhone 13 Pro Max", "iPhone 13 Pro"},
		{"", "test"},
		{"test", ""},
		{"x", "y"},
		{"a very long product description with many words", "a very different product description"},
	}

	config := DefaultSIMDConfig()
	config.Enabled = false

	for _, tc := range testCases {
		scalar := levenshteinDistanceScalar(tc.s, tc.t)
		optimized := ComputeDistanceOptimized(tc.s, tc.t, config)

		if scalar != optimized {
			t.Errorf("Mismatch for (%q, %q): scalar=%d, optimized=%d",
				tc.s, tc.t, scalar, optimized)
		}
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name string
		s    string
		t    string
	}{
		{"Very long identical", "a" + strings.Repeat("b", 1000) + "c", "a" + strings.Repeat("b", 1000) + "c"},
		{"Unicode", "café", "cafe"},
		{"Numbers", "123", "124"},
		{"Special chars", "hello!", "hello?"},
		{"Spaces", "hello world", "helloworld"},
		{"Tabs", "hello\tworld", "hello world"},
		{"Newlines", "hello\nworld", "hello world"},
	}

	config := DefaultSIMDConfig()
	config.Enabled = false

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify it doesn't panic and returns reasonable value
			result := ComputeDistanceOptimized(tt.s, tt.t, config)
			if result < 0 {
				t.Errorf("Negative distance for (%q, %q): %d", tt.s, tt.t, result)
			}
		})
	}
}

func TestIsSIMDAvailable(t *testing.T) {
	// This test just verifies the function doesn't panic
	// The actual availability depends on build flags and CPU features
	available := IsSIMDAvailable()
	// Log for informational purposes
	t.Logf("SIMD available: %v", available)
}

func TestSIMDConfigMinLength(t *testing.T) {
	config := DefaultSIMDConfig()
	config.Enabled = true
	config.MinStringLength = 100

	// Strings shorter than MinStringLength should use scalar
	short1 := "short"
	short2 := "sword"
	result := ComputeDistanceOptimized(short1, short2, config)

	// Should still get correct result (scalar fallback)
	expected := levenshteinDistanceScalar(short1, short2)
	if result != expected {
		t.Errorf("Short string with SIMD enabled: got %d, want %d", result, expected)
	}

	// Strings longer than MinStringLength may use SIMD if available
	long1 := "this is a longer product name for testing simd performance"
	long2 := "this is a longer product name for testing simd performanc"
	result = ComputeDistanceOptimized(long1, long2, config)

	expected = levenshteinDistanceScalar(long1, long2)
	if result != expected {
		t.Errorf("Long string: got %d, want %d", result, expected)
	}
}

func BenchmarkScalarLevenshtein(b *testing.B) {
	s := "apple iphone 13 pro max with a15 bionic chip"
	t := "apple iphone 14 pro max with a16 bionic chip"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = levenshteinDistanceScalar(s, t)
	}
}

func BenchmarkOptimizedLevenshtein(b *testing.B) {
	s := "apple iphone 13 pro max with a15 bionic chip"
	t := "apple iphone 14 pro max with a16 bionic chip"
	config := DefaultSIMDConfig()
	config.Enabled = false // Use scalar path for baseline

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeDistanceOptimized(s, t, config)
	}
}

func BenchmarkLongStringScalar(b *testing.B) {
	sBytes := make([]byte, 500)
	tBytes := make([]byte, 500)
	// Fill with pattern
	for i := 0; i < len(sBytes); i++ {
		sBytes[i] = 'a' + byte(i%26)
		tBytes[i] = 'a' + byte((i+1)%26)
	}
	s := string(sBytes)
	t := string(tBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = levenshteinDistanceScalar(s, t)
	}
}

func BenchmarkLongStringOptimized(b *testing.B) {
	s := string(make([]byte, 500))
	t := string(make([]byte, 500))
	config := DefaultSIMDConfig()
	config.Enabled = false

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ComputeDistanceOptimized(s, t, config)
	}
}
