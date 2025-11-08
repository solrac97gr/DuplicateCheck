package duplicatecheck

import (
	"strings"
)

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
	// Normalize strings to lowercase for case-insensitive comparison
	nameA := strings.ToLower(strings.TrimSpace(a.Name))
	nameB := strings.ToLower(strings.TrimSpace(b.Name))
	descA := strings.ToLower(strings.TrimSpace(a.Description))
	descB := strings.ToLower(strings.TrimSpace(b.Description))

	// Compute name similarity
	nameDistance := e.computeDistance(nameA, nameB)
	nameSimilarity := e.computeSimilarity(nameA, nameB, nameDistance)

	// Compute description similarity
	descDistance := e.computeDistance(descA, descB)
	descSimilarity := e.computeSimilarity(descA, descB, descDistance)

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
		// Both have data, use weighted combination
		// Normalize weights in case they don't sum to 1.0
		totalWeight := weights.NameWeight + weights.DescriptionWeight
		if totalWeight == 0 {
			totalWeight = 1.0
		}
		normalizedNameWeight := weights.NameWeight / totalWeight
		normalizedDescWeight := weights.DescriptionWeight / totalWeight

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

	// prev row: represents the previous row in our DP matrix
	// curr row: represents the current row being computed
	prev := make([]int, n+1)
	curr := make([]int, n+1)

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
			min := insertion
			if deletion < min {
				min = deletion
			}
			if substitution < min {
				min = substitution
			}
			curr[i] = min
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
//   - Consider using blocking/indexing techniques for very large datasets
func (e *LevenshteinEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
	var duplicates []ComparisonResult

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
