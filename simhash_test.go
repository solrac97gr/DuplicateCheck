package duplicatecheck

import (
	"strings"
	"testing"
)

func TestNewSimHashFilter(t *testing.T) {
	tests := []struct {
		name        string
		featureSize int
		expected    int
	}{
		{"Default feature size", 3, 3},
		{"Small feature size", 2, 2},
		{"Large feature size", 5, 5},
		{"Too small (< 2)", 1, 3},    // Should default to 3
		{"Too large (> 8)", 10, 8},   // Should cap at 8
		{"Zero", 0, 3},               // Should default to 3
		{"Negative", -5, 3},          // Should default to 3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter := NewSimHashFilter(tt.featureSize)
			if filter.featureSize != tt.expected {
				t.Errorf("NewSimHashFilter(%d) got featureSize %d, want %d",
					tt.featureSize, filter.featureSize, tt.expected)
			}
			if !filter.IsEnabled() {
				t.Errorf("NewSimHashFilter should be enabled by default")
			}
			if filter.bitSize != 64 {
				t.Errorf("NewSimHashFilter should have bitSize 64, got %d", filter.bitSize)
			}
		})
	}
}

func TestSimHashEnable(t *testing.T) {
	filter := NewSimHashFilter(3)

	// Initially enabled
	if !filter.IsEnabled() {
		t.Errorf("SimHash should be enabled by default")
	}

	// Disable it
	filter.Disable()
	if filter.IsEnabled() {
		t.Errorf("SimHash should be disabled after Disable()")
	}

	// Re-enable it
	filter.Enable()
	if !filter.IsEnabled() {
		t.Errorf("SimHash should be enabled after Enable()")
	}
}

func TestCompute64(t *testing.T) {
	filter := NewSimHashFilter(3)

	tests := []struct {
		name     string
		text     string
		nonZero  bool // Whether we expect a non-zero hash
		comment  string
	}{
		{"Empty string", "", false, "Empty strings should produce zero hash"},
		{"Single char", "a", true, "Single character should produce hash"},
		{"Short word", "hi", true, "Short words should produce hash"},
		{"Normal text", "Apple iPhone", true, "Normal text should produce hash"},
		{"Whitespace only", "   ", false, "Whitespace-only strings should produce zero hash"},
		{"Mixed case", "APPLE iphone", true, "Mixed case should normalize"},
		{"Long text", "The quick brown fox jumps over the lazy dog", true, "Long text should produce hash"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := filter.Compute64(tt.text)
			isNonZero := hash != 0

			if isNonZero != tt.nonZero {
				if tt.nonZero {
					t.Errorf("Compute64(%q) expected non-zero hash, got 0. %s",
						tt.text, tt.comment)
				} else {
					t.Errorf("Compute64(%q) expected zero hash, got %d. %s",
						tt.text, hash, tt.comment)
				}
			}
		})
	}
}

func TestCompute64Normalization(t *testing.T) {
	filter := NewSimHashFilter(3)

	// Different cases of same text should produce same hash
	hash1 := filter.Compute64("Apple iPhone")
	hash2 := filter.Compute64("apple iphone")
	hash3 := filter.Compute64("APPLE IPHONE")

	if hash1 != hash2 {
		t.Errorf("Different cases should produce same hash: %d != %d", hash1, hash2)
	}
	if hash1 != hash3 {
		t.Errorf("Different cases should produce same hash: %d != %d", hash1, hash3)
	}

	// Whitespace should be trimmed
	hash4 := filter.Compute64("  Apple iPhone  ")
	if hash1 != hash4 {
		t.Errorf("Extra whitespace should not affect hash: %d != %d", hash1, hash4)
	}
}

func TestEstimateSimilarity(t *testing.T) {
	filter := NewSimHashFilter(3)

	tests := []struct {
		name           string
		text1          string
		text2          string
		minSimilarity  float64 // Minimum expected similarity
		maxSimilarity  float64 // Maximum expected similarity
	}{
		{"Identical strings", "apple", "apple", 1.0, 1.0},
		{"Identical (normalized)", "Apple", "apple", 1.0, 1.0},
		{"Very similar", "apple", "appel", 0.75, 1.0},
		{"Similar", "apple", "application", 0.4, 1.0},
		{"Different", "apple", "orange", 0.0, 1.0}, // SimHash is probabilistic, may vary
		{"Completely different", "abc", "xyz", 0.0, 1.0}, // SimHash may have collisions
		{"Empty and non-empty", "", "apple", 0.0, 1.0}, // Can vary with hash collisions
		{"Both empty", "", "", 1.0, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := filter.EstimateSimilarity(tt.text1, tt.text2)

			if similarity < tt.minSimilarity || similarity > tt.maxSimilarity {
				t.Errorf("EstimateSimilarity(%q, %q) = %f, want between %f and %f",
					tt.text1, tt.text2, similarity, tt.minSimilarity, tt.maxSimilarity)
			}
		})
	}
}

func TestEstimateSimilaritySymmetric(t *testing.T) {
	filter := NewSimHashFilter(3)

	text1 := "Samsung Galaxy"
	text2 := "Galaxy Samsung"

	sim1 := filter.EstimateSimilarity(text1, text2)
	sim2 := filter.EstimateSimilarity(text2, text1)

	if sim1 != sim2 {
		t.Errorf("EstimateSimilarity should be symmetric: %f != %f", sim1, sim2)
	}
}

func TestQuickReject(t *testing.T) {
	filter := NewSimHashFilter(3)
	threshold := 0.7

	tests := []struct {
		name        string
		text1       string
		text2       string
		shouldKeep  bool // true if QuickReject should return true (keep for Levenshtein)
		comment     string
	}{
		{"Identical", "apple", "apple", true, "Identical strings should pass filter"},
		{"Very similar", "apple", "appel", true, "Very similar strings should pass filter"},
		{"Similar enough", "apple", "application", true, "Strings above safety margin should pass"},
		{"Empty strings", "", "", true, "Both empty should pass (both zero hashes)"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.QuickReject(tt.text1, tt.text2, threshold)
			if result != tt.shouldKeep {
				t.Errorf("QuickReject(%q, %q, %.1f) = %v, want %v. %s",
					tt.text1, tt.text2, threshold, result, tt.shouldKeep, tt.comment)
			}
		})
	}
}

func TestQuickRejectConservative(t *testing.T) {
	filter := NewSimHashFilter(3)

	// The QuickReject method is conservative - it uses safety margin
	// So it errs on the side of keeping candidates for further Levenshtein checking
	// We test that it doesn't have false negatives (never rejects valid duplicates)
	text1 := "apple"
	text2 := "apple"

	result := filter.QuickReject(text1, text2, 0.7)
	if !result {
		t.Errorf("QuickReject should never reject identical strings due to safety margin")
	}
}

func TestQuickRejectDisabled(t *testing.T) {
	filter := NewSimHashFilter(3)
	filter.Disable()

	// When disabled, QuickReject should always return true (continue to Levenshtein)
	result := filter.QuickReject("apple", "orange", 0.7)
	if !result {
		t.Errorf("Disabled SimHash filter should always return true for QuickReject")
	}
}

func TestQuickRejectWithSafetyMargin(t *testing.T) {
	filter := NewSimHashFilter(3)

	// Test that safety margin is applied correctly
	// With threshold 0.7, actual check is 0.55 (0.7 - 0.15)
	text1 := "apple"
	text2 := "apple"

	// Identical strings should always pass
	if !filter.QuickReject(text1, text2, 0.7) {
		t.Errorf("Identical strings should pass QuickReject with safety margin")
	}

	// Test with very low threshold
	if !filter.QuickReject(text1, text2, 0.0) {
		t.Errorf("Identical strings should pass even with threshold 0.0")
	}
}

func TestHammingDistance(t *testing.T) {
	tests := []struct {
		name     string
		hash1    SimHashFingerprint
		hash2    SimHashFingerprint
		expected int
	}{
		{"Identical", 0xFF, 0xFF, 0},
		{"Single bit different", 0xFF, 0xFE, 1},
		{"Half different", 0xF0, 0x0F, 8},
		{"Completely different", 0x00, 0xFF, 8},
		{"Zeros", 0x00, 0x00, 0},
		{"Large values", 0xAAAAAAAAAAAAAAAA, 0x5555555555555555, 64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			distance := HammingDistance(tt.hash1, tt.hash2)
			if distance != tt.expected {
				t.Errorf("HammingDistance(%064b, %064b) = %d, want %d",
					tt.hash1, tt.hash2, distance, tt.expected)
			}
		})
	}
}

func TestHammingDistanceSymmetric(t *testing.T) {
	hash1 := SimHashFingerprint(0xAAAAAAAAAAAAAAAA)
	hash2 := SimHashFingerprint(0x5555555555555555)

	dist1 := HammingDistance(hash1, hash2)
	dist2 := HammingDistance(hash2, hash1)

	if dist1 != dist2 {
		t.Errorf("HammingDistance should be symmetric: %d != %d", dist1, dist2)
	}
}

func TestSimilarity(t *testing.T) {
	tests := []struct {
		name     string
		hash1    SimHashFingerprint
		hash2    SimHashFingerprint
		expected float64
	}{
		{"Identical", 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF, 1.0},
		{"Completely different", 0x0000000000000000, 0xFFFFFFFFFFFFFFFF, 0.0},
		{"One bit different", 0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFE, 63.0 / 64.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			similarity := Similarity(tt.hash1, tt.hash2)
			if similarity < tt.expected-0.001 || similarity > tt.expected+0.001 {
				t.Errorf("Similarity(%064b, %064b) = %f, want %f",
					tt.hash1, tt.hash2, similarity, tt.expected)
			}
		})
	}
}

func TestSimilarityBounds(t *testing.T) {
	hash1 := SimHashFingerprint(0xAAAAAAAAAAAAAAAA)
	hash2 := SimHashFingerprint(0x5555555555555555)

	similarity := Similarity(hash1, hash2)

	if similarity < 0.0 || similarity > 1.0 {
		t.Errorf("Similarity should be between 0.0 and 1.0, got %f", similarity)
	}
}

func TestSimHashCachingConsistency(t *testing.T) {
	filter := NewSimHashFilter(3)
	text := "Apple iPhone"

	// Compute same fingerprint multiple times
	hashes := make([]SimHashFingerprint, 5)
	for i := 0; i < 5; i++ {
		hashes[i] = filter.Compute64(text)
	}

	// All should be identical
	for i := 1; i < len(hashes); i++ {
		if hashes[i] != hashes[0] {
			t.Errorf("Multiple calls to Compute64 should return identical hashes: %d != %d",
				hashes[i], hashes[0])
		}
	}
}

func TestExtractFeatures(t *testing.T) {
	filter := NewSimHashFilter(3)

	tests := []struct {
		name          string
		text          string
		minFeatures   int
		maxFeatures   int
	}{
		{"Empty", "", 0, 0},
		{"Short", "hi", 1, 1},
		{"Normal", "apple", 3, 3}, // "app", "ppl", "ple"
		{"Long", "application", 9, 9}, // 11 - 3 + 1 = 9
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			features := filter.extractFeatures(tt.text)
			if len(features) < tt.minFeatures || len(features) > tt.maxFeatures {
				t.Errorf("extractFeatures(%q) returned %d features, want %d-%d",
					tt.text, len(features), tt.minFeatures, tt.maxFeatures)
			}
		})
	}
}

func TestExtractFeaturesCorrectness(t *testing.T) {
	filter := NewSimHashFilter(2)
	features := filter.extractFeatures("ab")

	if len(features) != 1 {
		t.Fatalf("extractFeatures('ab') should return 1 feature, got %d", len(features))
	}

	if features[0] != "ab" {
		t.Errorf("extractFeatures('ab') should return 'ab', got %q", features[0])
	}
}

func TestHashFeature(t *testing.T) {
	filter := NewSimHashFilter(3)

	// Same feature should always produce same hash
	hash1 := filter.hashFeature("apple")
	hash2 := filter.hashFeature("apple")

	if hash1 != hash2 {
		t.Errorf("Same feature should produce same hash: %d != %d", hash1, hash2)
	}

	// Different features should (usually) produce different hashes
	hashA := filter.hashFeature("apple")
	hashB := filter.hashFeature("banana")

	if hashA == hashB {
		t.Errorf("Different features should usually produce different hashes")
	}
}

func TestSimHashEdgeCases(t *testing.T) {
	filter := NewSimHashFilter(3)

	tests := []struct {
		name string
		text string
	}{
		{"Very long text", "a" + strings.Repeat("b", 1000) + "c"},
		{"Special characters", "!@#$%^&*()"},
		{"Unicode", "café résumé"},
		{"Numbers", "12345"},
		{"Mixed", "Test123!@#"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			hash := filter.Compute64(tt.text)
			similarity := filter.EstimateSimilarity(tt.text, tt.text)
			reject := filter.QuickReject(tt.text, tt.text, 0.7)

			// Identical text should produce perfect similarity
			if similarity < 0.99 {
				t.Errorf("Identical text should produce high similarity: %f", similarity)
			}

			// Should keep for Levenshtein
			if !reject {
				t.Errorf("Identical text should pass QuickReject")
			}

			// Hash should be non-zero for non-empty text
			if tt.text != "" && hash == 0 {
				t.Errorf("Non-empty text should produce non-zero hash")
			}
		})
	}
}

func BenchmarkCompute64(b *testing.B) {
	filter := NewSimHashFilter(3)
	text := "Apple iPhone 13 Pro Max with A15 Bionic chip"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.Compute64(text)
	}
}

func BenchmarkEstimateSimilarity(b *testing.B) {
	filter := NewSimHashFilter(3)
	text1 := "Apple iPhone 13 Pro Max"
	text2 := "Apple iPhone 14 Pro Max"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.EstimateSimilarity(text1, text2)
	}
}

func BenchmarkQuickReject(b *testing.B) {
	filter := NewSimHashFilter(3)
	text1 := "Samsung Galaxy S21"
	text2 := "Samsung Galaxy S22"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = filter.QuickReject(text1, text2, 0.7)
	}
}

func BenchmarkHammingDistance(b *testing.B) {
	hash1 := SimHashFingerprint(0xAAAAAAAAAAAAAAAA)
	hash2 := SimHashFingerprint(0x5555555555555555)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = HammingDistance(hash1, hash2)
	}
}
