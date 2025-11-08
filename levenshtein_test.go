package main

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	engine := NewLevenshteinEngine()

	tests := []struct {
		name          string
		productA      Product
		productB      Product
		expectedDist  int
		minSimilarity float64
		maxSimilarity float64
	}{
		{
			name:          "Identical products",
			productA:      Product{ID: "1", Name: "iPhone 14 Pro"},
			productB:      Product{ID: "2", Name: "iPhone 14 Pro"},
			expectedDist:  0,
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "Very similar products (one character difference)",
			productA:      Product{ID: "1", Name: "Apple iPhone 14"},
			productB:      Product{ID: "2", Name: "Apple iPhone 13"},
			expectedDist:  1,
			minSimilarity: 0.93,
			maxSimilarity: 0.95,
		},
		{
			name:           "Similar phone models",
			productA:       Product{ID: "1", Name: "Samsung Galaxy S21"},
			productB:       Product{ID: "2", Name: "Samsung Galaxy S22"},
			expectedDist:   1,
			minSimilarity:  0.93,
			maxSimilarity:  0.96,
		},
		{
			name:          "Classic example: kitten vs sitting",
			productA:      Product{ID: "1", Name: "kitten"},
			productB:      Product{ID: "2", Name: "sitting"},
			expectedDist:  3,
			minSimilarity: 0.57,
			maxSimilarity: 0.58,
		},
		{
			name:           "Different products",
			productA:       Product{ID: "1", Name: "Apple iPhone"},
			productB:       Product{ID: "2", Name: "Samsung Galaxy"},
			expectedDist:   12,
			minSimilarity:  0.0,
			maxSimilarity:  0.15,
		},
		{
			name:          "Empty product name",
			productA:      Product{ID: "1", Name: ""},
			productB:      Product{ID: "2", Name: "iPhone"},
			expectedDist:  6,
			minSimilarity: 0.0,
			maxSimilarity: 0.0,
		},
		{
			name:          "Both empty",
			productA:      Product{ID: "1", Name: ""},
			productB:      Product{ID: "2", Name: ""},
			expectedDist:  0,
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "Case insensitive comparison",
			productA:      Product{ID: "1", Name: "IPHONE"},
			productB:      Product{ID: "2", Name: "iphone"},
			expectedDist:  0,
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Compare(tt.productA, tt.productB)

			// Check distance
			if result.Distance != tt.expectedDist {
				t.Errorf("Distance = %d, want %d", result.Distance, tt.expectedDist)
			}

			// Check similarity is in expected range
			if result.Similarity < tt.minSimilarity || result.Similarity > tt.maxSimilarity {
				t.Errorf("Similarity = %.4f, want between %.4f and %.4f",
					result.Similarity, tt.minSimilarity, tt.maxSimilarity)
			}

			// Verify result contains the correct products
			if result.ProductA.ID != tt.productA.ID {
				t.Errorf("ProductA.ID = %s, want %s", result.ProductA.ID, tt.productA.ID)
			}
			if result.ProductB.ID != tt.productB.ID {
				t.Errorf("ProductB.ID = %s, want %s", result.ProductB.ID, tt.productB.ID)
			}
		})
	}
}

func TestLevenshteinFindDuplicates(t *testing.T) {
	engine := NewLevenshteinEngine()

	products := []Product{
		{ID: "1", Name: "Apple iPhone 14 Pro"},
		{ID: "2", Name: "Apple iPhone 14 Pro"}, // Exact duplicate
		{ID: "3", Name: "Apple iPhone 13 Pro"}, // Very similar
		{ID: "4", Name: "Samsung Galaxy S23"},
		{ID: "5", Name: "Samsung Galaxy S22"}, // Similar to 4
		{ID: "6", Name: "Sony Headphones WH-1000XM5"},
		{ID: "7", Name: "Sony Headphones WH-1000XM4"}, // Similar to 6
	}

	t.Run("High threshold (0.95) - only exact or near-exact matches", func(t *testing.T) {
		duplicates := engine.FindDuplicates(products, 0.95)

		if len(duplicates) < 1 {
			t.Errorf("Expected at least 1 duplicate pair at 0.95 threshold, got %d", len(duplicates))
		}

		// Check that we found the exact duplicate
		foundExactDuplicate := false
		for _, dup := range duplicates {
			if (dup.ProductA.ID == "1" && dup.ProductB.ID == "2") ||
				(dup.ProductA.ID == "2" && dup.ProductB.ID == "1") {
				foundExactDuplicate = true
				if dup.Similarity != 1.0 {
					t.Errorf("Exact duplicate similarity = %.4f, want 1.0", dup.Similarity)
				}
			}
		}

		if !foundExactDuplicate {
			t.Error("Expected to find exact duplicate between products 1 and 2")
		}
	})

	t.Run("Medium threshold (0.80) - catches more similar items", func(t *testing.T) {
		duplicates := engine.FindDuplicates(products, 0.80)

		if len(duplicates) < 3 {
			t.Errorf("Expected at least 3 duplicate pairs at 0.80 threshold, got %d", len(duplicates))
		}

		// All returned pairs should meet the threshold
		for _, dup := range duplicates {
			if dup.Similarity < 0.80 {
				t.Errorf("Found duplicate with similarity %.4f below threshold 0.80", dup.Similarity)
			}
		}
	})

	t.Run("Low threshold (0.50) - very permissive", func(t *testing.T) {
		duplicates := engine.FindDuplicates(products, 0.50)

		// Should find even more pairs
		if len(duplicates) < 3 {
			t.Errorf("Expected at least 3 duplicate pairs at 0.50 threshold, got %d", len(duplicates))
		}
	})

	t.Run("Empty product list", func(t *testing.T) {
		duplicates := engine.FindDuplicates([]Product{}, 0.80)

		if len(duplicates) != 0 {
			t.Errorf("Expected 0 duplicates for empty list, got %d", len(duplicates))
		}
	})

	t.Run("Single product", func(t *testing.T) {
		duplicates := engine.FindDuplicates([]Product{products[0]}, 0.80)

		if len(duplicates) != 0 {
			t.Errorf("Expected 0 duplicates for single product, got %d", len(duplicates))
		}
	})
}

func TestLevenshteinGetName(t *testing.T) {
	engine := NewLevenshteinEngine()
	name := engine.GetName()

	if name == "" {
		t.Error("Engine name should not be empty")
	}

	if name != "Levenshtein Distance" {
		t.Errorf("Engine name = %s, want 'Levenshtein Distance'", name)
	}
}

// Benchmark the distance computation for different string lengths
func BenchmarkLevenshteinDistance(b *testing.B) {
	engine := NewLevenshteinEngine()

	benchmarks := []struct {
		name     string
		productA Product
		productB Product
	}{
		{
			name:     "Short strings (6-7 chars)",
			productA: Product{ID: "1", Name: "kitten"},
			productB: Product{ID: "2", Name: "sitting"},
		},
		{
			name:     "Medium strings (~20 chars)",
			productA: Product{ID: "1", Name: "Apple iPhone 14 Pro"},
			productB: Product{ID: "2", Name: "Apple iPhone 13 Pro"},
		},
		{
			name:     "Long strings (~50 chars)",
			productA: Product{ID: "1", Name: "Sony WH-1000XM5 Wireless Noise-Cancelling Headphones"},
			productB: Product{ID: "2", Name: "Sony WH-1000XM4 Wireless Noise-Cancelling Headphones"},
		},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.Compare(bm.productA, bm.productB)
			}
		})
	}
}

// Benchmark finding duplicates in different sized datasets
func BenchmarkLevenshteinFindDuplicates(b *testing.B) {
	engine := NewLevenshteinEngine()

	// Generate test products
	generateProducts := func(n int) []Product {
		products := make([]Product, n)
		for i := 0; i < n; i++ {
			products[i] = Product{
				ID:   string(rune(i)),
				Name: "Sample Product Name Number " + string(rune(i+'0')),
			}
		}
		return products
	}

	benchmarks := []struct {
		name         string
		productCount int
	}{
		{"10 products", 10},
		{"50 products", 50},
		{"100 products", 100},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			products := generateProducts(bm.productCount)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				engine.FindDuplicates(products, 0.80)
			}
		})
	}
}
