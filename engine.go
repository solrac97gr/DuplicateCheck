package duplicatecheck

import (
	"strings"
	"sync"
	"sync/atomic"
)

// Product represents an item in your ecommerce system
type Product struct {
	ID          string
	Name        string
	Description string // Product description up to 3000 characters
	// Cached normalized versions (lazy initialization)
	normalizedName string
	normalizedDesc string
	normalized     uint32 // atomic flag: 0 = not normalized, 1 = normalized
	// N-gram caching for repeated comparisons
	ngramsCache map[int][][2]string // ngramsCache[n] = n-grams for this n value
	ngramsMutex sync.RWMutex         // Protects ngramsCache and normalized strings
}

// getNormalizedStrings returns cached normalized (lowercase, trimmed) versions of Name and Description
// This avoids repeated string operations in batch comparisons
// Uses double-checked locking pattern with atomic flag for thread-safe lazy initialization
func (p *Product) getNormalizedStrings() (name, desc string) {
	// Fast path: if already normalized, return immediately (atomic read, no lock needed)
	if atomic.LoadUint32(&p.normalized) == 1 {
		return p.normalizedName, p.normalizedDesc
	}

	// Slow path: need to normalize - acquire lock
	p.ngramsMutex.Lock()
	defer p.ngramsMutex.Unlock()

	// Double-check: another goroutine might have done the work while waiting for lock
	if p.normalized == 0 {
		p.normalizedName = strings.ToLower(strings.TrimSpace(p.Name))
		p.normalizedDesc = strings.ToLower(strings.TrimSpace(p.Description))
		// Atomic store to ensure visibility across goroutines
		atomic.StoreUint32(&p.normalized, 1)
	}
	return p.normalizedName, p.normalizedDesc
}

// GetNgrams returns cached n-grams for the product name
// Generates and caches n-grams on first call, returns cached version on subsequent calls
// n parameter specifies the n-gram size (e.g., 2 for bigrams, 3 for trigrams)
// Thread-safe with double-checked locking pattern
func (p *Product) GetNgrams(n int) [][2]string {
	if n < 1 {
		return [][2]string{}
	}

	// Check if already cached (fast path - read-heavy, most calls hit this)
	p.ngramsMutex.RLock()
	if p.ngramsCache != nil {
		if cached, exists := p.ngramsCache[n]; exists {
			p.ngramsMutex.RUnlock()
			return cached
		}
	}
	p.ngramsMutex.RUnlock()

	// Slow path: need to generate and cache
	// Generate n-grams (outside lock to minimize contention)
	name, _ := p.getNormalizedStrings()
	ngrams := generateNgrams(name, n)

	// Store result with proper locking
	p.ngramsMutex.Lock()
	defer p.ngramsMutex.Unlock()

	// Ensure cache is initialized
	if p.ngramsCache == nil {
		p.ngramsCache = make(map[int][][2]string)
	}

	// Double-check: another goroutine might have already cached this n-gram size
	if cached, exists := p.ngramsCache[n]; exists {
		return cached
	}

	// Cache the result
	p.ngramsCache[n] = ngrams
	return ngrams
}

// generateNgrams generates n-grams of size n from a string
// Returns pairs of (ngram_string, position) for efficient comparison
func generateNgrams(s string, n int) [][2]string {
	if n < 1 || len(s) < n {
		return [][2]string{}
	}

	ngrams := make([][2]string, 0, len(s)-n+1)
	runes := []rune(s)

	for i := 0; i <= len(runes)-n; i++ {
		ngram := string(runes[i : i+n])
		ngrams = append(ngrams, [2]string{ngram, string(rune(i))}) // Store ngram and position
	}

	return ngrams
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
