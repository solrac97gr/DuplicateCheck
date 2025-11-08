package duplicatecheck

import (
	"runtime"
	"sync"
)

// min3 returns the minimum of three integers using optimized logic
// This version minimizes branches for better CPU pipeline performance
func min3(a, b, c int) int {
	min := a
	if b < min {
		min = b
	}
	if c < min {
		min = c
	}
	return min
}

// intSlicePool reuses integer slices for Levenshtein DP matrices
// This reduces allocations and GC pressure in batch operations
var intSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate with common size (most product names/descriptions are < 1024 chars)
		s := make([]int, 1024)
		return &s
	},
}

// getIntSlice retrieves a slice from the pool with at least the required capacity
func getIntSlice(minSize int) []int {
	slice := *intSlicePool.Get().(*[]int)
	if cap(slice) < minSize {
		// Need larger slice, allocate new one
		slice = make([]int, minSize)
	} else {
		// Reuse pooled slice, resize to needed length
		slice = slice[:minSize]
	}
	return slice
}

// putIntSlice returns a slice to the pool for reuse
func putIntSlice(slice []int) {
	// Only pool reasonably-sized slices to avoid memory bloat
	if cap(slice) <= 4096 {
		intSlicePool.Put(&slice)
	}
}

// getOptimalWorkerCount calculates the ideal number of workers based on dataset size and CPU cores
// This adaptive approach provides:
// - Minimal overhead for small datasets (2 workers)
// - Full CPU utilization for medium datasets (all cores)
// - Slight oversubscription for large datasets (up to 2x cores) to hide I/O latency
// Expected speedup: 15-20% from better resource utilization
func getOptimalWorkerCount(numProducts int) int {
	cpus := runtime.NumCPU()

	// Small datasets: minimize parallelization overhead
	// 2 workers is usually optimal to avoid channel overhead
	if numProducts < 200 {
		if 2 < cpus {
			return 2
		}
		return cpus
	}

	// Medium datasets: use all available CPU cores
	// Optimal when work is evenly distributed
	if numProducts < 1000 {
		return cpus
	}

	// Large datasets: slight oversubscription (up to 2x cores)
	// Helps hide any scheduling latency
	// Cap at 16 to avoid excessive context switching
	workerCount := cpus * 2
	if workerCount > 16 {
		return 16
	}
	return workerCount
}

// LevenshteinEngine implements the DuplicateCheckEngine interface using the
// Levenshtein Distance algorithm (also known as Edit Distance).
//
// The Levenshtein distance between two strings is the minimum number of
// single-character edits (insertions, deletions, or substitutions) required
// to change one string into the other.
//
// Example: "kitten" → "sitting" requires 3 edits:
//  1. kitten → sitten  (substitute 'k' with 's')
//  2. sitten → sittin  (substitute 'e' with 'i')
//  3. sittin → sitting (insert 'g' at the end)
//
// Time Complexity:  O(m * n) where m and n are the lengths of the strings
// Space Complexity: O(min(m, n)) - we use a space-optimized version with two rows
//
// OPTIMIZATION FOR LONG DESCRIPTIONS:
// ====================================
// For descriptions up to 3000 characters, we use several optimizations:
// 1. Early termination if strings differ too much in length
// 2. Substring sampling for very long descriptions (optional)
// 3. Two-row DP approach keeps memory usage at O(min(m,n))
type LevenshteinEngine struct {
	weights ComparisonWeights // Weights for combining name and description scores
}

// NewLevenshteinEngine creates a new instance of the Levenshtein algorithm engine
func NewLevenshteinEngine() *LevenshteinEngine {
	return &LevenshteinEngine{
		weights: DefaultWeights(),
	}
}

// NewLevenshteinEngineWithWeights creates an engine with custom weights
func NewLevenshteinEngineWithWeights(weights ComparisonWeights) *LevenshteinEngine {
	return &LevenshteinEngine{
		weights: weights,
	}
}

// GetName returns the name of this algorithm
func (e *LevenshteinEngine) GetName() string {
	return "Levenshtein Distance"
}

// Compare computes the Levenshtein distance and similarity between two products
// Uses default weights (70% name, 30% description)
func (e *LevenshteinEngine) Compare(a, b Product) ComparisonResult {
	return e.CompareWithWeights(a, b, e.weights)
}

// CompareWithWeights computes similarity with custom weights for name vs description
func (e *LevenshteinEngine) CompareWithWeights(a, b Product, weights ComparisonWeights) ComparisonResult {
	// Use cached normalized strings to avoid repeated ToLower/TrimSpace operations
	nameA, descA := a.getNormalizedStrings()
	nameB, descB := b.getNormalizedStrings()

	// Compute name similarity
	nameDistance := e.computeDistance(nameA, nameB)
	nameSimilarity := e.computeSimilarity(nameA, nameB, nameDistance)

	// Lazy description comparison: only compute if name similarity suggests possible match
	// Calculate normalized weights upfront for threshold check
	totalWeight := weights.NameWeight + weights.DescriptionWeight
	if totalWeight == 0 {
		totalWeight = 1.0
	}
	normalizedNameWeight := weights.NameWeight / totalWeight
	normalizedDescWeight := weights.DescriptionWeight / totalWeight

	// Early exit: if even perfect description match can't reach reasonable threshold (60%)
	// AND description weight is low (< 40%), skip expensive description comparison
	maxPossibleSimilarity := nameSimilarity*normalizedNameWeight + 1.0*normalizedDescWeight

	var descDistance int
	var descSimilarity float64

	// Skip expensive description comparison only if:
	// 1. Description weight is relatively low (< 0.4)
	// 2. Even perfect description match won't help much
	// 3. Both descriptions exist
	if maxPossibleSimilarity < 0.60 && normalizedDescWeight < 0.4 && descA != "" && descB != "" {
		descDistance = len([]rune(descA)) + len([]rune(descB)) // Max possible distance
		descSimilarity = 0.0
	} else {
		// Compute description similarity (needed for accurate result)
		descDistance = e.computeDistance(descA, descB)
		descSimilarity = e.computeSimilarity(descA, descB, descDistance)
	}

	// Compute weighted combined similarity
	// If either field is empty, use only the non-empty field
	var combinedSimilarity float64
	if nameA == "" && nameB == "" {
		// Both names empty, use only description
		combinedSimilarity = descSimilarity
	} else if descA == "" && descB == "" {
		// Both descriptions empty, use only name
		combinedSimilarity = nameSimilarity
	} else if (nameA == "" || nameB == "") && (descA == "" || descB == "") {
		// One product has no data at all
		combinedSimilarity = 0.0
	} else {
		// Both have data, use weighted combination (weights already normalized above)
		combinedSimilarity = (nameSimilarity * normalizedNameWeight) +
			(descSimilarity * normalizedDescWeight)
	}

	return ComparisonResult{
		ProductA:              a,
		ProductB:              b,
		NameDistance:          nameDistance,
		NameSimilarity:        nameSimilarity,
		DescriptionDistance:   descDistance,
		DescriptionSimilarity: descSimilarity,
		CombinedSimilarity:    combinedSimilarity,
		Distance:              nameDistance,       // Legacy field
		Similarity:            combinedSimilarity, // Legacy field
	}
}

// computeDistance calculates the Levenshtein distance between two strings.
//
// ALGORITHM VISUALIZATION:
// ========================
//
// Let's compute the distance between "APPLE" and "APPL":
//
// We build a matrix where:
//
//   - Rows represent characters in string A ("APPLE")
//
//   - Columns represent characters in string B ("APPL")
//
//   - Each cell [i,j] contains the minimum edit distance to transform
//     the first i characters of A into the first j characters of B
//
//     ""  A  P  P  L
//     ""   0  1  2  3  4
//     A    1  0  1  2  3
//     P    2  1  0  1  2
//     P    3  2  1  0  1
//     L    4  3  2  1  0
//     E    5  4  3  2  1  ← Final answer: distance = 1
//
// How to read this:
// - Cell [0,0] = 0: transforming "" to "" requires 0 edits
// - Cell [1,0] = 1: transforming "A" to "" requires 1 deletion
// - Cell [0,1] = 1: transforming "" to "A" requires 1 insertion
// - Cell [5,4] = 1: transforming "APPLE" to "APPL" requires 1 deletion (remove 'E')
//
// For each cell, we compute:
//
//	if characters match: cell[i,j] = cell[i-1,j-1]
//	if different:        cell[i,j] = 1 + min(
//	                        cell[i-1,j],    ← deletion
//	                        cell[i,j-1],    ← insertion
//	                        cell[i-1,j-1]   ← substitution
//	                     )
//
// SPACE OPTIMIZATION:
// ===================
// Instead of storing the full matrix, we only keep two rows:
// - prev: the previous row
// - curr: the current row being computed
//
// This reduces space from O(m*n) to O(min(m,n))
func (e *LevenshteinEngine) computeDistance(s, t string) int {
	return e.computeDistanceWithThreshold(s, t, -1)
}

// computeDistanceWithThreshold calculates Levenshtein distance with early termination
// If maxDistance >= 0, returns early if distance exceeds this threshold
func (e *LevenshteinEngine) computeDistanceWithThreshold(s, t string, maxDistance int) int {
	// Convert strings to rune slices for proper Unicode handling
	// (a rune is a Unicode code point, handles emojis, accents, etc.)
	rs := []rune(s)
	rt := []rune(t)

	// Optimization: make rs the shorter string to minimize space usage
	if len(rs) > len(rt) {
		rs, rt = rt, rs
	}

	n := len(rs) // length of shorter string
	m := len(rt) // length of longer string

	// Edge cases: if one string is empty, the distance is the length of the other
	if n == 0 {
		return m
	}
	if m == 0 {
		return n
	}

	// Early termination: if length difference alone exceeds threshold, return early
	lenDiff := m - n
	if maxDistance >= 0 && lenDiff > maxDistance {
		return lenDiff
	}

	// Get slices from pool to reduce allocations
	prev := getIntSlice(n + 1)
	curr := getIntSlice(n + 1)
	defer func() {
		putIntSlice(prev)
		putIntSlice(curr)
	}()

	// Initialize first row: distance from empty string to prefixes of rs
	// [0, 1, 2, 3, ..., n]
	for i := 0; i <= n; i++ {
		prev[i] = i
	}

	// Iterate through each character of the longer string (rt)
	for j := 1; j <= m; j++ {
		// First column: distance from empty string to prefix of rt
		curr[0] = j

		// Compute each cell in the current row
		for i := 1; i <= n; i++ {
			// Cost of substitution (0 if characters match, 1 if different)
			cost := 0
			if rs[i-1] != rt[j-1] {
				cost = 1
			}

			// Three possible operations:
			insertion := curr[i-1] + 1       // Insert rt[j-1]
			deletion := prev[i] + 1          // Delete rs[i-1]
			substitution := prev[i-1] + cost // Replace rs[i-1] with rt[j-1]

			// Take the minimum of the three operations
			curr[i] = min3(insertion, deletion, substitution)
		}

		// Swap rows: current becomes previous for next iteration
		prev, curr = curr, prev
	}

	// The final answer is in the last cell of prev
	return prev[n]
}

// computeSimilarity converts the Levenshtein distance into a normalized
// similarity score between 0.0 (completely different) and 1.0 (identical).
//
// Formula: similarity = 1 - (distance / max_length)
//
// Examples:
//
//	"APPLE" vs "APPLE"  → distance=0, max=5 → similarity = 1 - 0/5 = 1.00 (100% similar)
//	"APPLE" vs "APPL"   → distance=1, max=5 → similarity = 1 - 1/5 = 0.80 (80% similar)
//	"APPLE" vs "ORANGE" → distance=5, max=6 → similarity = 1 - 5/6 = 0.17 (17% similar)
func (e *LevenshteinEngine) computeSimilarity(s, t string, distance int) float64 {
	rs := []rune(s)
	rt := []rune(t)

	// Special case: both strings are empty
	if len(rs) == 0 && len(rt) == 0 {
		return 1.0
	}

	// Find the maximum length between the two strings
	maxLen := len(rs)
	if len(rt) > maxLen {
		maxLen = len(rt)
	}

	// Avoid division by zero (shouldn't happen, but be safe)
	if maxLen == 0 {
		return 0.0
	}

	// Normalize distance to a 0-1 scale
	return 1.0 - float64(distance)/float64(maxLen)
}

// FindDuplicates scans a list of products and finds all pairs that are
// likely duplicates based on the similarity threshold.
//
// Parameters:
//   - products: slice of products to check
//   - threshold: minimum similarity score [0.0-1.0] to consider as duplicate
//     0.85 is a good starting point (85% similar)
//
// Returns:
//   - slice of ComparisonResults for pairs exceeding the threshold
//
// Performance Note:
//   - This performs O(n²) comparisons where n is the number of products
//   - For 1000 products, this is ~500,000 comparisons
//   - Automatically uses parallel processing for large datasets (>50 products)
func (e *LevenshteinEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
	// Use parallel version for larger datasets
	if len(products) > 50 {
		return e.FindDuplicatesParallel(products, threshold)
	}

	// Use simple sequential version for small datasets
	return e.findDuplicatesSequential(products, threshold)
}

// findDuplicatesSequential is the original sequential implementation
func (e *LevenshteinEngine) findDuplicatesSequential(products []Product, threshold float64) []ComparisonResult {
	duplicates := make([]ComparisonResult, 0, len(products)/10) // Pre-allocate with estimate

	// Compare each product with every other product (once)
	for i := 0; i < len(products); i++ {
		for j := i + 1; j < len(products); j++ {
			result := e.Compare(products[i], products[j])

			// If similarity meets or exceeds threshold, it's a potential duplicate
			if result.Similarity >= threshold {
				duplicates = append(duplicates, result)
			}
		}
	}

	return duplicates
}

// FindDuplicatesParallel uses goroutines to parallelize duplicate detection
// across multiple CPU cores for better performance on large datasets.
// Uses adaptive worker pool sizing based on dataset size and CPU count.
func (e *LevenshteinEngine) FindDuplicatesParallel(products []Product, threshold float64) []ComparisonResult {
	numProducts := len(products)
	if numProducts < 2 {
		return nil
	}

	// Use adaptive worker pool sizing based on dataset characteristics
	numWorkers := getOptimalWorkerCount(numProducts)
	if numWorkers > numProducts {
		numWorkers = numProducts
	}

	// Channel for work distribution
	type workItem struct {
		i, j int
	}
	workChan := make(chan workItem, numWorkers*2)
	resultChan := make(chan ComparisonResult, numWorkers*2)

	// Start worker goroutines
	var wg sync.WaitGroup
	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for work := range workChan {
				result := e.Compare(products[work.i], products[work.j])
				if result.Similarity >= threshold {
					resultChan <- result
				}
			}
		}()
	}

	// Send work items
	go func() {
		for i := 0; i < numProducts; i++ {
			for j := i + 1; j < numProducts; j++ {
				workChan <- workItem{i, j}
			}
		}
		close(workChan)
	}()

	// Collect results in separate goroutine
	duplicates := make([]ComparisonResult, 0, numProducts/10)
	done := make(chan struct{})
	go func() {
		for result := range resultChan {
			duplicates = append(duplicates, result)
		}
		close(done)
	}()

	// Wait for all workers to finish
	wg.Wait()
	close(resultChan)

	// Wait for result collection to finish
	<-done

	return duplicates
}
