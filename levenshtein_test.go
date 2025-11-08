package duplicatecheck

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
			productA:      Product{ID: "1", Name: "iPhone 14 Pro", Description: ""},
			productB:      Product{ID: "2", Name: "iPhone 14 Pro", Description: ""},
			expectedDist:  0,
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "Very similar products (one character difference)",
			productA:      Product{ID: "1", Name: "Apple iPhone 14", Description: ""},
			productB:      Product{ID: "2", Name: "Apple iPhone 13", Description: ""},
			expectedDist:  1,
			minSimilarity: 0.93,
			maxSimilarity: 0.95,
		},
		{
			name:          "Similar phone models",
			productA:      Product{ID: "1", Name: "Samsung Galaxy S21", Description: ""},
			productB:      Product{ID: "2", Name: "Samsung Galaxy S22", Description: ""},
			expectedDist:  1,
			minSimilarity: 0.93,
			maxSimilarity: 0.96,
		},
		{
			name:          "Classic example: kitten vs sitting",
			productA:      Product{ID: "1", Name: "kitten", Description: ""},
			productB:      Product{ID: "2", Name: "sitting", Description: ""},
			expectedDist:  3,
			minSimilarity: 0.57,
			maxSimilarity: 0.58,
		},
		{
			name:          "Different products",
			productA:      Product{ID: "1", Name: "Apple iPhone", Description: ""},
			productB:      Product{ID: "2", Name: "Samsung Galaxy", Description: ""},
			expectedDist:  12,
			minSimilarity: 0.0,
			maxSimilarity: 0.15,
		},
		{
			name:          "Empty product name",
			productA:      Product{ID: "1", Name: "", Description: ""},
			productB:      Product{ID: "2", Name: "iPhone", Description: ""},
			expectedDist:  6,
			minSimilarity: 0.0,
			maxSimilarity: 0.0,
		},
		{
			name:          "Both empty",
			productA:      Product{ID: "1", Name: "", Description: ""},
			productB:      Product{ID: "2", Name: "", Description: ""},
			expectedDist:  0,
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name:          "Case insensitive comparison",
			productA:      Product{ID: "1", Name: "IPHONE", Description: ""},
			productB:      Product{ID: "2", Name: "iphone", Description: ""},
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

func TestLevenshteinWithDescriptions(t *testing.T) {
	engine := NewLevenshteinEngine()

	tests := []struct {
		name          string
		productA      Product
		productB      Product
		minSimilarity float64
		maxSimilarity float64
	}{
		{
			name: "Identical names and descriptions",
			productA: Product{
				ID:          "1",
				Name:        "iPhone 14 Pro",
				Description: "The latest iPhone with advanced camera system",
			},
			productB: Product{
				ID:          "2",
				Name:        "iPhone 14 Pro",
				Description: "The latest iPhone with advanced camera system",
			},
			minSimilarity: 1.0,
			maxSimilarity: 1.0,
		},
		{
			name: "Same name, different descriptions",
			productA: Product{
				ID:          "1",
				Name:        "iPhone 14 Pro",
				Description: "Brand new sealed in box with 1 year warranty",
			},
			productB: Product{
				ID:          "2",
				Name:        "iPhone 14 Pro",
				Description: "Used excellent condition comes with charger",
			},
			minSimilarity: 0.75, // Name match carries it (70% weight)
			maxSimilarity: 0.85,
		},
		{
			name: "Different names, similar descriptions",
			productA: Product{
				ID:          "1",
				Name:        "iPhone 14 Pro",
				Description: "Latest flagship phone with triple camera system and A16 chip",
			},
			productB: Product{
				ID:          "2",
				Name:        "iPhone 13 Pro",
				Description: "Latest flagship phone with triple camera system and A16 chip",
			},
			minSimilarity: 0.90,
			maxSimilarity: 1.0,
		},
		{
			name: "Long descriptions (simulating real ecommerce)",
			productA: Product{
				ID:   "1",
				Name: "Samsung Galaxy S23 Ultra",
				Description: "Experience the pinnacle of smartphone innovation with the Samsung Galaxy S23 Ultra. " +
					"This flagship device features a stunning 6.8-inch Dynamic AMOLED 2X display, " +
					"powerful Snapdragon 8 Gen 2 processor, and an incredible camera system with " +
					"200MP main sensor. With 12GB RAM and 256GB storage, this phone handles everything " +
					"you throw at it. The 5000mAh battery ensures all-day performance.",
			},
			productB: Product{
				ID:   "2",
				Name: "Samsung Galaxy S23 Ultra",
				Description: "Experience the pinnacle of smartphone innovation with the Samsung Galaxy S23 Ultra. " +
					"This flagship device features a stunning 6.8-inch Dynamic AMOLED 2X display, " +
					"powerful Snapdragon 8 Gen 2 processor, and an incredible camera system with " +
					"200MP main sensor. With 8GB RAM and 512GB storage, this phone handles everything " +
					"you throw at it. The 5000mAh battery ensures all-day performance.",
			},
			minSimilarity: 0.95, // Very similar, minor spec differences
			maxSimilarity: 1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Compare(tt.productA, tt.productB)

			// Check combined similarity is in expected range
			if result.CombinedSimilarity < tt.minSimilarity || result.CombinedSimilarity > tt.maxSimilarity {
				t.Errorf("CombinedSimilarity = %.4f, want between %.4f and %.4f\n"+
					"  NameSimilarity: %.4f, DescriptionSimilarity: %.4f",
					result.CombinedSimilarity, tt.minSimilarity, tt.maxSimilarity,
					result.NameSimilarity, result.DescriptionSimilarity)
			}
		})
	}
}

func TestLevenshteinCustomWeights(t *testing.T) {
	productA := Product{
		ID:          "1",
		Name:        "iPhone 14",
		Description: "Brand new sealed",
	}
	productB := Product{
		ID:          "2",
		Name:        "iPhone 13",
		Description: "Brand new sealed",
	}

	// Test with name-only weighting (100% name, 0% description)
	weightsNameOnly := ComparisonWeights{NameWeight: 1.0, DescriptionWeight: 0.0}
	engine := NewLevenshteinEngineWithWeights(weightsNameOnly)
	result := engine.CompareWithWeights(productA, productB, weightsNameOnly)

	// Should be the same as name similarity
	if result.CombinedSimilarity != result.NameSimilarity {
		t.Errorf("With name-only weights, combined should equal name similarity. Got %.4f vs %.4f",
			result.CombinedSimilarity, result.NameSimilarity)
	}

	// Test with description-only weighting (0% name, 100% description)
	weightsDescOnly := ComparisonWeights{NameWeight: 0.0, DescriptionWeight: 1.0}
	result2 := engine.CompareWithWeights(productA, productB, weightsDescOnly)

	// Should be the same as description similarity
	if result2.CombinedSimilarity != result2.DescriptionSimilarity {
		t.Errorf("With desc-only weights, combined should equal description similarity. Got %.4f vs %.4f",
			result2.CombinedSimilarity, result2.DescriptionSimilarity)
	}

	// Test with equal weighting (50/50)
	weightsEqual := ComparisonWeights{NameWeight: 0.5, DescriptionWeight: 0.5}
	result3 := engine.CompareWithWeights(productA, productB, weightsEqual)

	expected := (result3.NameSimilarity + result3.DescriptionSimilarity) / 2.0
	if result3.CombinedSimilarity < expected-0.01 || result3.CombinedSimilarity > expected+0.01 {
		t.Errorf("With 50/50 weights, combined should be average. Got %.4f, want %.4f",
			result3.CombinedSimilarity, expected)
	}
}

func TestLevenshteinFindDuplicates(t *testing.T) {
	engine := NewLevenshteinEngine()

	products := []Product{
		{ID: "1", Name: "Apple iPhone 14 Pro", Description: ""},
		{ID: "2", Name: "Apple iPhone 14 Pro", Description: ""}, // Exact duplicate
		{ID: "3", Name: "Apple iPhone 13 Pro", Description: ""}, // Very similar
		{ID: "4", Name: "Samsung Galaxy S23", Description: ""},
		{ID: "5", Name: "Samsung Galaxy S22", Description: ""}, // Similar to 4
		{ID: "6", Name: "Sony Headphones WH-1000XM5", Description: ""},
		{ID: "7", Name: "Sony Headphones WH-1000XM4", Description: ""}, // Similar to 6
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

// BenchmarkLevenshteinLongDescriptions tests performance with realistic long product descriptions
func BenchmarkLevenshteinLongDescriptions(b *testing.B) {
	engine := NewLevenshteinEngine()

	longDesc1 := "The Samsung Galaxy S23 Ultra is a flagship smartphone featuring a stunning 6.8-inch Dynamic AMOLED 2X display with 120Hz refresh rate. " +
		"Powered by the latest Snapdragon 8 Gen 2 processor, it delivers exceptional performance for gaming, multitasking, and content creation. " +
		"The camera system is truly impressive with a 200MP main sensor, 12MP ultra-wide, and dual telephoto lenses offering 3x and 10x optical zoom. " +
		"With integrated S Pen support, 5000mAh battery with 45W fast charging, and up to 1TB of storage, this phone is built for power users. " +
		"The premium build quality features Gorilla Glass Victus 2 and IP68 water resistance. Available in Phantom Black, Cream, Green, and Lavender colors. " +
		"Includes 12GB RAM, 5G connectivity, Wi-Fi 6E, Bluetooth 5.3, stereo speakers, and wireless charging. Perfect for photography enthusiasts and professionals."

	longDesc2 := "The Samsung Galaxy S23 Ultra is a flagship smartphone featuring a stunning 6.8-inch Dynamic AMOLED 2X display with 120Hz refresh rate. " +
		"Powered by the latest Snapdragon 8 Gen 2 processor, it delivers exceptional performance for gaming, multitasking, and content creation. " +
		"The camera system is truly impressive with a 200MP main sensor, 12MP ultra-wide, and dual telephoto lenses offering 3x and 10x optical zoom. " +
		"With integrated S Pen support, 5000mAh battery with 45W fast charging, and up to 512GB of storage, this phone is built for power users. " +
		"The premium build quality features Gorilla Glass Victus 2 and IP68 water resistance. Available in Phantom Black, Cream, Green, and Lavender colors. " +
		"Includes 8GB RAM, 5G connectivity, Wi-Fi 6E, Bluetooth 5.3, stereo speakers, and wireless charging. Perfect for photography enthusiasts and professionals."

	benchmarks := []struct {
		name     string
		productA Product
		productB Product
	}{
		{
			name: "Long descriptions (~750 chars)",
			productA: Product{
				ID:          "1",
				Name:        "Samsung Galaxy S23 Ultra 1TB",
				Description: longDesc1,
			},
			productB: Product{
				ID:          "2",
				Name:        "Samsung Galaxy S23 Ultra 512GB",
				Description: longDesc2,
			},
		},
		{
			name: "Very long descriptions (~2000 chars)",
			productA: Product{
				ID:          "1",
				Name:        "Premium Laptop",
				Description: longDesc1 + " " + longDesc1 + " " + longDesc1[:500],
			},
			productB: Product{
				ID:          "2",
				Name:        "Premium Laptop",
				Description: longDesc2 + " " + longDesc2 + " " + longDesc2[:500],
			},
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
