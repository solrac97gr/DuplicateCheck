package duplicatecheck_test

import (
	"fmt"
	"testing"

	"github.com/solrac97gr/duplicatecheck"
)

// Example_basic demonstrates the basic usage of the library with Levenshtein engine
func Example_basic() {
	// Create a Levenshtein engine (good for small/medium catalogs)
	engine := duplicatecheck.NewLevenshteinEngine()

	// Define products
	products := []duplicatecheck.Product{
		{ID: "1", Name: "iPhone 13 Pro", Description: "128GB Blue"},
		{ID: "2", Name: "iPhone 13 Pro Max", Description: "128GB Blue"},
		{ID: "3", Name: "Samsung Galaxy S21", Description: "256GB Black"},
	}

	// Find duplicates with 80% threshold
	duplicates := engine.FindDuplicates(products, 0.80)

	fmt.Printf("Found %d potential duplicates\n", len(duplicates))
	// Output: Found 1 potential duplicates
}

// Example_hybrid demonstrates high-performance duplicate detection for large catalogs
func Example_hybrid() {
	// Create a Hybrid engine (MinHash + LSH for large catalogs)
	engine := duplicatecheck.NewHybridEngine()

	// Build index from your product catalog with realistic descriptions
	catalog := []duplicatecheck.Product{
		{
			ID:   "1",
			Name: "iPhone 13 Pro Max",
			Description: "Apple iPhone 13 Pro Max with advanced camera system, " +
				"Super Retina XDR display with ProMotion technology, " +
				"A15 Bionic chip, and 5G connectivity. Available in multiple colors.",
		},
		{
			ID:          "2",
			Name:        "Samsung Galaxy S22",
			Description: "Samsung Galaxy S22 with premium design and powerful performance.",
		},
	}
	engine.BuildIndex(catalog)

	// Check if new product is duplicate (very similar to product 1)
	newProduct := duplicatecheck.Product{
		ID:   "new",
		Name: "iPhone 13 Pro Max",
		Description: "Apple iPhone 13 Pro Max featuring advanced camera system, " +
			"Super Retina XDR display with ProMotion, " +
			"A15 Bionic processor, and 5G support. Multiple colors available.",
	}

	_ = engine.FindDuplicatesForOne(newProduct, 0.75)

	// LSH is probabilistic, so results may vary
	// For deterministic comparison, use LevenshteinEngine
	fmt.Printf("Hybrid engine can find duplicates with O(1) candidate lookup\n")
}

// Example_customWeights shows how to adjust the comparison weights
func Example_customWeights() {
	// Create custom weights (emphasize description over name)
	weights := duplicatecheck.ComparisonWeights{
		NameWeight:        0.30, // 30% weight on name
		DescriptionWeight: 0.70, // 70% weight on description
	}

	engine := duplicatecheck.NewLevenshteinEngineWithWeights(weights)

	p1 := duplicatecheck.Product{
		ID:          "1",
		Name:        "Product A",
		Description: "Very detailed description about the product features",
	}

	p2 := duplicatecheck.Product{
		ID:          "2",
		Name:        "Product B",
		Description: "Very detailed description about the product features",
	}

	result := engine.Compare(p1, p2)
	fmt.Printf("Similarity: %.2f%%\n", result.Similarity*100)
}

// TestExampleIntegration verifies that examples work correctly
func TestExampleIntegration(t *testing.T) {
	t.Run("Levenshtein Engine", func(t *testing.T) {
		engine := duplicatecheck.NewLevenshteinEngine()

		products := []duplicatecheck.Product{
			{ID: "1", Name: "Test Product", Description: "Description"},
			{ID: "2", Name: "Test Product", Description: "Description Similar"},
		}

		duplicates := engine.FindDuplicates(products, 0.80)

		if len(duplicates) == 0 {
			t.Error("Expected to find duplicates")
		}
	})

	t.Run("Hybrid Engine", func(t *testing.T) {
		engine := duplicatecheck.NewHybridEngine()

		catalog := []duplicatecheck.Product{
			{ID: "1", Name: "iPhone 13 Pro Max", Description: "128GB Blue"},
			{ID: "2", Name: "Samsung S21", Description: "256GB"},
		}
		engine.BuildIndex(catalog)

		newProduct := duplicatecheck.Product{
			ID:          "new",
			Name:        "iPhone 13 Pro Max",
			Description: "128GB Blue Like New",
		}

		duplicates := engine.FindDuplicatesForOne(newProduct, 0.85)

		if len(duplicates) == 0 {
			t.Error("Expected to find similar products")
		}
	})

	t.Run("Custom Weights", func(t *testing.T) {
		weights := duplicatecheck.ComparisonWeights{
			NameWeight:        0.50,
			DescriptionWeight: 0.50,
		}

		engine := duplicatecheck.NewLevenshteinEngineWithWeights(weights)

		p1 := duplicatecheck.Product{ID: "1", Name: "A", Description: "Same"}
		p2 := duplicatecheck.Product{ID: "2", Name: "B", Description: "Same"}

		result := engine.Compare(p1, p2)

		if result.Similarity <= 0 {
			t.Error("Expected similarity greater than 0")
		}
	})

	t.Run("Interface Usage", func(t *testing.T) {
		// Both engines implement DuplicateCheckEngine interface
		var engine duplicatecheck.DuplicateCheckEngine

		// Use Levenshtein
		engine = duplicatecheck.NewLevenshteinEngine()
		if engine.GetName() != "Levenshtein Distance" {
			t.Errorf("Expected 'Levenshtein Distance', got %s", engine.GetName())
		}

		// Switch to Hybrid
		engine = duplicatecheck.NewHybridEngine()
		if engine.GetName() != "Hybrid (MinHash+LSH → Levenshtein)" {
			t.Errorf("Expected 'Hybrid (MinHash+LSH → Levenshtein)', got %s", engine.GetName())
		}
	})
}
