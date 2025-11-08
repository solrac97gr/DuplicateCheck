package main

// Product represents an item in your ecommerce system
type Product struct {
	ID          string
	Name        string
	Description string // Product description up to 3000 characters
}

// ComparisonResult contains the similarity metrics between two products
type ComparisonResult struct {
	ProductA              Product
	ProductB              Product
	NameDistance          int     // Raw distance score for names
	NameSimilarity        float64 // Normalized similarity for names [0.0-1.0]
	DescriptionDistance   int     // Raw distance score for descriptions
	DescriptionSimilarity float64 // Normalized similarity for descriptions [0.0-1.0]
	CombinedSimilarity    float64 // Weighted combined similarity score [0.0-1.0]
	Distance              int     // Legacy: kept for backward compatibility
	Similarity            float64 // Legacy: kept for backward compatibility (same as CombinedSimilarity)
}

// ComparisonWeights defines how much weight to give to name vs description
type ComparisonWeights struct {
	NameWeight        float64 // Weight for name similarity (0.0-1.0)
	DescriptionWeight float64 // Weight for description similarity (0.0-1.0)
}

// DefaultWeights returns sensible default weights
// Name is weighted more heavily since it's typically more distinctive
func DefaultWeights() ComparisonWeights {
	return ComparisonWeights{
		NameWeight:        0.7, // 70% weight on name
		DescriptionWeight: 0.3, // 30% weight on description
	}
}

// DuplicateCheckEngine is the interface that all similarity algorithms must implement.
// This allows us to swap different algorithms (Levenshtein, Jaro-Winkler, Cosine, etc.)
// and compare their performance and accuracy for detecting duplicate products.
type DuplicateCheckEngine interface {
	// GetName returns the human-readable name of the algorithm
	GetName() string

	// Compare computes the similarity between two products based on their names and descriptions
	// Returns a ComparisonResult with distance and similarity metrics
	Compare(a, b Product) ComparisonResult

	// CompareWithWeights allows custom weighting of name vs description similarity
	CompareWithWeights(a, b Product, weights ComparisonWeights) ComparisonResult

	// FindDuplicates searches for potential duplicates in a product list
	// Returns pairs of products that exceed the similarity threshold [0.0-1.0]
	FindDuplicates(products []Product, threshold float64) []ComparisonResult
}
