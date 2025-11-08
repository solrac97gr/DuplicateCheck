package duplicatecheck

import (
	"testing"
)

func TestGetNgrams(t *testing.T) {
	product := Product{
		ID:   "test-1",
		Name: "Apple iPhone",
	}

	tests := []struct {
		name     string
		n        int
		minCount int // Minimum expected n-grams
	}{
		{"Bigrams", 2, 11}, // "apple iphone" has 11 bigrams
		{"Trigrams", 3, 10},
		{"4-grams", 4, 9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ngrams := product.GetNgrams(tt.n)
			if len(ngrams) < tt.minCount {
				t.Errorf("GetNgrams(%d) returned %d n-grams, want at least %d",
					tt.n, len(ngrams), tt.minCount)
			}

			// Verify all n-grams have correct length
			for _, ng := range ngrams {
				if len([]rune(ng[0])) != tt.n {
					t.Errorf("N-gram %q has length %d, want %d",
						ng[0], len([]rune(ng[0])), tt.n)
				}
			}
		})
	}
}

func TestNgramCaching(t *testing.T) {
	product := Product{
		ID:   "test-cache",
		Name: "Samsung Galaxy",
	}

	// First call - generates and caches
	ngrams1 := product.GetNgrams(3)
	initialCount := len(ngrams1)

	// Second call - should return cached version
	ngrams2 := product.GetNgrams(3)
	if len(ngrams2) != initialCount {
		t.Errorf("Cached n-grams count mismatch: first=%d, second=%d",
			initialCount, len(ngrams2))
	}

	// Verify they're identical
	if len(ngrams1) != len(ngrams2) {
		t.Errorf("Cache returned different count: first=%d, second=%d",
			len(ngrams1), len(ngrams2))
	}

	// Different n value should generate different cache
	ngrams3 := product.GetNgrams(4)
	if len(ngrams3) == initialCount {
		t.Errorf("4-grams should have different count than 3-grams")
	}
}

func TestNgramEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		product  Product
		n        int
		expected int
	}{
		{"Empty name", Product{ID: "1", Name: ""}, 2, 0},
		{"Single character", Product{ID: "2", Name: "a"}, 2, 0},
		{"n equals name length", Product{ID: "3", Name: "test"}, 4, 1},
		{"n larger than name", Product{ID: "4", Name: "hi"}, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ngrams := tt.product.GetNgrams(tt.n)
			if len(ngrams) != tt.expected {
				t.Errorf("GetNgrams(%d) returned %d, want %d",
					tt.n, len(ngrams), tt.expected)
			}
		})
	}
}

func TestGenerateNgrams(t *testing.T) {
	text := "apple"
	ngrams := generateNgrams(text, 2)

	expected := []string{"ap", "pp", "pl", "le"}
	if len(ngrams) != len(expected) {
		t.Errorf("generateNgrams returned %d n-grams, want %d",
			len(ngrams), len(expected))
	}

	for i, ng := range ngrams {
		if i < len(expected) && ng[0] != expected[i] {
			t.Errorf("N-gram %d: got %q, want %q", i, ng[0], expected[i])
		}
	}
}

func TestNgramConcurrency(t *testing.T) {
	product := Product{
		ID:   "test-concurrent",
		Name: "Concurrent Test Product",
	}

	// Simulate concurrent access
	results := make(chan int, 3)
	for i := 0; i < 3; i++ {
		go func() {
			ngrams := product.GetNgrams(3)
			results <- len(ngrams)
		}()
	}

	// Verify all goroutines got same result
	first := <-results
	for i := 0; i < 2; i++ {
		if result := <-results; result != first {
			t.Errorf("Concurrent access returned different counts: %d vs %d",
				first, result)
		}
	}
}

func BenchmarkGetNgrams(b *testing.B) {
	product := Product{
		ID:   "bench-1",
		Name: "Lorem ipsum dolor sit amet consectetur adipiscing elit",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = product.GetNgrams(3)
	}
}

func BenchmarkNgramGeneration(b *testing.B) {
	longText := "The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = generateNgrams(longText, 3)
	}
}
