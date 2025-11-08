// TEMPLATE FOR NEW ALGORITHMS
// ============================
// Copy this file and replace the placeholder implementation with your algorithm
// For example: jaro_winkler.go, cosine_similarity.go, jaccard.go, etc.

package main

// TemplateEngine is a template for implementing new duplicate detection algorithms
// Replace "Template" with your algorithm name (e.g., "JaroWinkler", "Cosine", etc.)
type TemplateEngine struct {
	// Add any configuration fields your algorithm needs
	// For example:
	// threshold float64
	// prefixScale float64 (for Jaro-Winkler)
}

// NewTemplateEngine creates a new instance of your algorithm
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		// Initialize your configuration here
	}
}

// GetName returns the human-readable name of your algorithm
func (e *TemplateEngine) GetName() string {
	return "Your Algorithm Name"
}

// Compare computes the similarity between two products
//
// IMPLEMENTATION GUIDE:
// =====================
// 1. Extract product names and normalize them (lowercase, trim, etc.)
// 2. Compute your algorithm's distance/similarity metric
// 3. Convert to normalized similarity score [0.0-1.0] where 1.0 = identical
// 4. Return ComparisonResult with all metrics
//
// EXAMPLE STRUCTURE:
//
//	nameA := strings.ToLower(a.Name)
//	nameB := strings.ToLower(b.Name)
//
//	distance := e.computeYourDistance(nameA, nameB)
//	similarity := e.convertToSimilarity(distance, nameA, nameB)
//
//	return ComparisonResult{
//	    ProductA:   a,
//	    ProductB:   b,
//	    Distance:   distance,
//	    Similarity: similarity,
//	}
func (e *TemplateEngine) Compare(a, b Product) ComparisonResult {
	// TODO: Implement your comparison logic here

	return ComparisonResult{
		ProductA:   a,
		ProductB:   b,
		Distance:   0,   // Replace with actual distance
		Similarity: 0.0, // Replace with actual similarity
	}
}

// FindDuplicates scans a list of products and finds potential duplicates
//
// IMPLEMENTATION GUIDE:
// =====================
// The standard O(n²) approach works for most cases:
//
//	var duplicates []ComparisonResult
//	for i := 0; i < len(products); i++ {
//	    for j := i + 1; j < len(products); j++ {
//	        result := e.Compare(products[i], products[j])
//	        if result.Similarity >= threshold {
//	            duplicates = append(duplicates, result)
//	        }
//	    }
//	}
//	return duplicates
//
// For large datasets, consider optimizations:
// - Blocking/bucketing (group by category, brand, price range first)
// - Parallel processing (use goroutines with worker pools)
// - Approximate nearest neighbor (LSH, KD-trees, etc.)
func (e *TemplateEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
	var duplicates []ComparisonResult

	// TODO: Implement your duplicate finding logic here
	// Use the standard nested loop or optimize as needed

	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			result := e.Compare(products[i], products[j])
			if result.Similarity >= threshold {
				duplicates = append(duplicates, result)
			}
		}
	}

	return duplicates
}

// NEXT STEPS:
// ===========
// 1. Copy this file to a new file (e.g., jaro_winkler.go)
// 2. Replace "Template" with your algorithm name throughout
// 3. Implement the actual algorithm logic
// 4. Create a corresponding test file (*_test.go)
// 5. Add your engine to the engines slice in main.go:
//    engines := []DuplicateCheckEngine{
//        NewLevenshteinEngine(),
//        NewYourAlgorithmEngine(),  // <- Add here
//    }
// 6. Run tests: go test ./...
// 7. Run benchmarks: go test -bench=. -benchmem

// TESTING TEMPLATE:
// =================
// Create a file named template_test.go with:
/*
package main

import "testing"

func TestTemplateDistance(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name          string
		productA      Product
		productB      Product
		minSimilarity float64
	}{
		{
			name:          "Identical products",
			productA:      Product{ID: "1", Name: "iPhone"},
			productB:      Product{ID: "2", Name: "iPhone"},
			minSimilarity: 1.0,
		},
		// Add more test cases
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := engine.Compare(tt.productA, tt.productB)
			if result.Similarity < tt.minSimilarity {
				t.Errorf("Similarity = %.4f, want >= %.4f",
					result.Similarity, tt.minSimilarity)
			}
		})
	}
}
*/
