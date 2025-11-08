package main

// Product represents an item in your ecommerce system
type Product struct {
	ID   string
	Name string
}

// ComparisonResult contains the similarity metrics between two products
type ComparisonResult struct {
	ProductA   Product
	ProductB   Product
	Distance   int     // The raw distance score (lower is more similar)
	Similarity float64 // Normalized similarity score [0.0-1.0] where 1.0 means identical
}

// DuplicateCheckEngine is the interface that all similarity algorithms must implement.
// This allows us to swap different algorithms (Levenshtein, Jaro-Winkler, Cosine, etc.)
// and compare their performance and accuracy for detecting duplicate products.
type DuplicateCheckEngine interface {
	// GetName returns the human-readable name of the algorithm
	GetName() string

	// Compare computes the similarity between two products based on their names
	// Returns a ComparisonResult with distance and similarity metrics
	Compare(a, b Product) ComparisonResult

	// FindDuplicates searches for potential duplicates in a product list
	// Returns pairs of products that exceed the similarity threshold [0.0-1.0]
	FindDuplicates(products []Product, threshold float64) []ComparisonResult
}
