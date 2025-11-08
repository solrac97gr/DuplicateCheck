package duplicatecheck

import (
	"hash/fnv"
	"math"
	"sort"
	"strings"
)

// HybridEngine implements a multi-stage hybrid architecture for efficient duplicate detection
// Stage 1: Fast filtering using MinHash + LSH to reduce millions to hundreds
// Stage 2: Medium refinement using n-grams and blocking
// Stage 3: Precise verification using Levenshtein on final candidates
type HybridEngine struct {
	levenshteinEngine *LevenshteinEngine
	lshIndex          *LSHIndex
	numHashFunctions  int
	numBands          int
	shingleSize       int
}

// LSHIndex implements Locality Sensitive Hashing for fast similarity search
type LSHIndex struct {
	bands       []map[uint64][]string // Each band maps hash -> product IDs
	numBands    int
	rowsPerBand int
	products    map[string]Product // Product ID -> Product
}

// NewHybridEngine creates a hybrid duplicate detection engine
func NewHybridEngine() *HybridEngine {
	return &HybridEngine{
		levenshteinEngine: NewLevenshteinEngine(),
		numHashFunctions:  100, // Number of MinHash functions
		numBands:          20,  // Number of LSH bands
		shingleSize:       3,   // 3-gram shingles
	}
}

// GetName returns the name of this algorithm
func (e *HybridEngine) GetName() string {
	return "Hybrid (MinHash+LSH → Levenshtein)"
}

// BuildIndex creates the LSH index for a collection of products
// This is done once during initialization or when products change
func (e *HybridEngine) BuildIndex(products []Product) {
	rowsPerBand := e.numHashFunctions / e.numBands

	e.lshIndex = &LSHIndex{
		bands:       make([]map[uint64][]string, e.numBands),
		numBands:    e.numBands,
		rowsPerBand: rowsPerBand,
		products:    make(map[string]Product),
	}

	// Initialize band maps
	for i := 0; i < e.numBands; i++ {
		e.lshIndex.bands[i] = make(map[uint64][]string)
	}

	// Index each product
	for _, product := range products {
		e.indexProduct(product)
	}
}

// indexProduct adds a product to the LSH index
func (e *HybridEngine) indexProduct(product Product) {
	// Store product
	e.lshIndex.products[product.ID] = product

	// Generate combined text for hashing
	text := strings.ToLower(product.Name + " " + product.Description)

	// Generate shingles (n-grams)
	shingles := generateShingles(text, e.shingleSize)

	// Compute MinHash signature
	signature := computeMinHashSignature(shingles, e.numHashFunctions)

	// Add to LSH bands
	for bandIdx := 0; bandIdx < e.numBands; bandIdx++ {
		// Hash this band's rows together
		bandHash := hashBand(signature, bandIdx*e.lshIndex.rowsPerBand,
			(bandIdx+1)*e.lshIndex.rowsPerBand)

		// Add product ID to this band bucket
		e.lshIndex.bands[bandIdx][bandHash] = append(
			e.lshIndex.bands[bandIdx][bandHash],
			product.ID,
		)
	}
}

// Compare implements single product comparison (for interface compatibility)
func (e *HybridEngine) Compare(a, b Product) ComparisonResult {
	return e.levenshteinEngine.Compare(a, b)
}

// CompareWithWeights implements weighted comparison (for interface compatibility)
func (e *HybridEngine) CompareWithWeights(a, b Product, weights ComparisonWeights) ComparisonResult {
	return e.levenshteinEngine.CompareWithWeights(a, b, weights)
}

// FindDuplicates uses the hybrid multi-stage approach
// Stage 1: LSH filtering (reduces to ~1-5% of corpus)
// Stage 2: Levenshtein verification on candidates
func (e *HybridEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
	if e.lshIndex == nil {
		// Fallback to regular Levenshtein if index not built
		return e.levenshteinEngine.FindDuplicates(products, threshold)
	}

	var duplicates []ComparisonResult
	checked := make(map[string]bool) // Track checked pairs to avoid duplicates

	// For each product, find candidates using LSH
	for _, product := range products {
		candidates := e.findCandidates(product)

		// Stage 3: Precise verification with Levenshtein
		for _, candidateID := range candidates {
			// Skip self-comparison
			if candidateID == product.ID {
				continue
			}

			// Skip if already checked this pair
			pairKey := makePairKey(product.ID, candidateID)
			if checked[pairKey] {
				continue
			}
			checked[pairKey] = true

			// Get candidate product
			candidate, exists := e.lshIndex.products[candidateID]
			if !exists {
				continue
			}

			// Precise comparison with Levenshtein
			result := e.levenshteinEngine.Compare(product, candidate)

			if result.CombinedSimilarity >= threshold {
				duplicates = append(duplicates, result)
			}
		}
	}

	return duplicates
}

// FindDuplicatesForOne finds duplicates for a single product against the indexed corpus
// This is the key method for the "1 article vs 500 articles" scenario
func (e *HybridEngine) FindDuplicatesForOne(product Product, threshold float64) []ComparisonResult {
	if e.lshIndex == nil {
		return nil
	}

	// Stage 1: Fast LSH filtering
	candidates := e.findCandidates(product)

	var duplicates []ComparisonResult

	// Stage 2: Precise verification with Levenshtein (only on candidates)
	for _, candidateID := range candidates {
		candidate, exists := e.lshIndex.products[candidateID]
		if !exists {
			continue
		}

		result := e.levenshteinEngine.Compare(product, candidate)

		if result.CombinedSimilarity >= threshold {
			duplicates = append(duplicates, result)
		}
	}

	return duplicates
}

// findCandidates uses LSH to find similar products quickly
// Returns product IDs that are likely similar
func (e *HybridEngine) findCandidates(product Product) []string {
	// Generate combined text
	text := strings.ToLower(product.Name + " " + product.Description)

	// Generate shingles
	shingles := generateShingles(text, e.shingleSize)

	// Compute MinHash signature
	signature := computeMinHashSignature(shingles, e.numHashFunctions)

	// Find candidates by checking all bands
	candidateSet := make(map[string]bool)

	for bandIdx := 0; bandIdx < e.numBands; bandIdx++ {
		// Hash this band
		bandHash := hashBand(signature, bandIdx*e.lshIndex.rowsPerBand,
			(bandIdx+1)*e.lshIndex.rowsPerBand)

		// Get all products in this bucket
		if bucket, exists := e.lshIndex.bands[bandIdx][bandHash]; exists {
			for _, productID := range bucket {
				candidateSet[productID] = true
			}
		}
	}

	// Convert set to slice
	candidates := make([]string, 0, len(candidateSet))
	for id := range candidateSet {
		candidates = append(candidates, id)
	}

	return candidates
}

// generateShingles creates n-gram shingles from text
func generateShingles(text string, n int) []string {
	// Clean text: lowercase and split into tokens
	tokens := strings.Fields(text)

	if len(tokens) < n {
		return []string{text}
	}

	shingles := make([]string, 0, len(tokens)-n+1)

	// Create word n-grams
	for i := 0; i <= len(tokens)-n; i++ {
		shingle := strings.Join(tokens[i:i+n], " ")
		shingles = append(shingles, shingle)
	}

	return shingles
}

// computeMinHashSignature computes MinHash signature for a set of shingles
func computeMinHashSignature(shingles []string, numHashes int) []uint32 {
	signature := make([]uint32, numHashes)

	// Initialize with max values
	for i := range signature {
		signature[i] = math.MaxUint32
	}

	// For each shingle
	for _, shingle := range shingles {
		// Hash with different seeds
		for i := 0; i < numHashes; i++ {
			hash := hashWithSeed(shingle, uint32(i))
			if hash < signature[i] {
				signature[i] = hash
			}
		}
	}

	return signature
}

// hashWithSeed hashes a string with a given seed
func hashWithSeed(s string, seed uint32) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	h.Write([]byte{byte(seed), byte(seed >> 8), byte(seed >> 16), byte(seed >> 24)})
	return h.Sum32()
}

// hashBand hashes a portion of the signature to create a band hash
func hashBand(signature []uint32, start, end int) uint64 {
	h := fnv.New64a()
	for i := start; i < end && i < len(signature); i++ {
		bytes := []byte{
			byte(signature[i]),
			byte(signature[i] >> 8),
			byte(signature[i] >> 16),
			byte(signature[i] >> 24),
		}
		h.Write(bytes)
	}
	return h.Sum64()
}

// makePairKey creates a consistent key for a product pair (order-independent)
func makePairKey(id1, id2 string) string {
	if id1 < id2 {
		return id1 + "|" + id2
	}
	return id2 + "|" + id1
}

// GetIndexStats returns statistics about the LSH index
func (e *HybridEngine) GetIndexStats() map[string]interface{} {
	if e.lshIndex == nil {
		return map[string]interface{}{"indexed": false}
	}

	stats := map[string]interface{}{
		"indexed":        true,
		"total_products": len(e.lshIndex.products),
		"num_bands":      e.numBands,
		"rows_per_band":  e.lshIndex.rowsPerBand,
	}

	// Calculate average bucket size
	totalBuckets := 0
	totalProducts := 0
	maxBucketSize := 0

	for _, band := range e.lshIndex.bands {
		totalBuckets += len(band)
		for _, bucket := range band {
			size := len(bucket)
			totalProducts += size
			if size > maxBucketSize {
				maxBucketSize = size
			}
		}
	}

	if totalBuckets > 0 {
		stats["avg_bucket_size"] = float64(totalProducts) / float64(totalBuckets)
	}
	stats["max_bucket_size"] = maxBucketSize
	stats["total_buckets"] = totalBuckets

	return stats
}

// EstimateCandidateReduction estimates how many candidates LSH will find
func (e *HybridEngine) EstimateCandidateReduction(product Product) int {
	if e.lshIndex == nil {
		return 0
	}
	candidates := e.findCandidates(product)
	return len(candidates)
}

// BlockingStrategy implements simple blocking for additional optimization
type BlockingStrategy struct {
	// Block by first few characters of name
	blockSize int
}

// NewBlockingStrategy creates a blocking strategy
func NewBlockingStrategy(blockSize int) *BlockingStrategy {
	return &BlockingStrategy{blockSize: blockSize}
}

// GetBlockKey returns the block key for a product
func (s *BlockingStrategy) GetBlockKey(product Product) string {
	name := strings.ToLower(strings.TrimSpace(product.Name))
	if len(name) <= s.blockSize {
		return name
	}
	return name[:s.blockSize]
}

// GroupByBlocks groups products by their block keys
func (s *BlockingStrategy) GroupByBlocks(products []Product) map[string][]Product {
	blocks := make(map[string][]Product)

	for _, product := range products {
		key := s.GetBlockKey(product)
		blocks[key] = append(blocks[key], product)
	}

	return blocks
}

// SortByRelevance sorts comparison results by similarity (descending)
func SortByRelevance(results []ComparisonResult) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].CombinedSimilarity > results[j].CombinedSimilarity
	})
}
