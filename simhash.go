package duplicatecheck

import (
	"fmt"
	"hash/fnv"
	"math/bits"
	"strings"
)

// SimHashFilter implements probabilistic similarity estimation using SimHash algorithm
// SimHash computes a 64-bit hash where similar strings have similar hashes
// This allows O(1) similarity estimation compared to O(m×n) for Levenshtein
//
// Algorithm:
// 1. Extract features (n-grams) from text
// 2. Hash each feature using FNV-64
// 3. Build 64-bit vector by summing hash bits
// 4. Compare using Hamming distance on final 64-bit hash
//
// Benefits:
// - O(1) similarity estimation (just count bit differences)
// - Very fast pre-filtering before expensive Levenshtein
// - Fingerprint can be cached along with product
//
// Trade-offs:
// - Probabilistic (may have false positives/negatives)
// - Less accurate than Levenshtein
// - Best used as pre-filter, not final verification
type SimHashFilter struct {
	featureSize int   // Size of n-grams (typically 3-5)
	enabled     bool  // Whether filter is enabled
	bitSize     int   // Usually 64 bits
}

// SimHashFingerprint represents a 64-bit SimHash for a string
type SimHashFingerprint uint64

// NewSimHashFilter creates a new SimHash filter with specified feature size
// featureSize: size of n-grams to extract (2-5 recommended, default 3)
func NewSimHashFilter(featureSize int) *SimHashFilter {
	if featureSize < 2 {
		featureSize = 3
	}
	if featureSize > 8 {
		featureSize = 8
	}

	return &SimHashFilter{
		featureSize: featureSize,
		enabled:     true,
		bitSize:     64,
	}
}

// Enable turns on SimHash filtering
func (s *SimHashFilter) Enable() {
	s.enabled = true
}

// Disable turns off SimHash filtering
func (s *SimHashFilter) Disable() {
	s.enabled = false
}

// IsEnabled returns whether SimHash filtering is active
func (s *SimHashFilter) IsEnabled() bool {
	return s.enabled
}

// Compute64 computes a 64-bit SimHash fingerprint for a string
// Returns a 64-bit hash where similar strings have similar hashes
func (s *SimHashFilter) Compute64(text string) SimHashFingerprint {
	// Normalize text
	text = strings.ToLower(strings.TrimSpace(text))
	if len(text) == 0 {
		return 0
	}

	// Extract features (n-grams)
	features := s.extractFeatures(text)
	if len(features) == 0 {
		return 0
	}

	// Build bit vector
	vector := make([]int, s.bitSize)
	for _, feature := range features {
		hash := s.hashFeature(feature)
		// For each bit position, increment if bit is set
		for i := 0; i < s.bitSize; i++ {
			if (hash & (uint64(1) << uint(i))) != 0 {
				vector[i]++
			} else {
				vector[i]--
			}
		}
	}

	// Convert vector to 64-bit hash
	var result uint64
	for i := 0; i < s.bitSize; i++ {
		if vector[i] > 0 {
			result |= (uint64(1) << uint(i))
		}
	}

	return SimHashFingerprint(result)
}

// EstimateSimilarity estimates similarity between two strings using Hamming distance
// Returns value between 0.0 (completely different) and 1.0 (identical)
// Calculation: 1.0 - (hammingDistance / 64)
func (s *SimHashFilter) EstimateSimilarity(text1, text2 string) float64 {
	hash1 := s.Compute64(text1)
	hash2 := s.Compute64(text2)

	// Calculate Hamming distance (count differing bits)
	xor := uint64(hash1) ^ uint64(hash2)
	hammingDistance := bits.OnesCount64(xor)

	// Convert to similarity (0.0 = completely different, 1.0 = identical)
	return 1.0 - float64(hammingDistance)/float64(s.bitSize)
}

// QuickReject determines if two strings should be rejected as dissimilar
// Returns true if strings should continue to Levenshtein (likely similar)
// Returns false if strings are definitely dissimilar (can skip Levenshtein)
func (s *SimHashFilter) QuickReject(text1, text2 string, threshold float64) bool {
	if !s.enabled {
		return true // If disabled, always continue to Levenshtein
	}

	// Compute estimated similarity
	similarity := s.EstimateSimilarity(text1, text2)

	// Conservative approach: only reject if very confident
	// Use threshold - 0.15 safety margin to avoid false negatives
	return similarity >= (threshold - 0.15)
}

// extractFeatures extracts n-grams from text
// Returns list of n-gram features
func (s *SimHashFilter) extractFeatures(text string) []string {
	runes := []rune(text)
	if len(runes) < s.featureSize {
		// For very short text, return the text itself as feature
		if len(text) > 0 {
			return []string{text}
		}
		return []string{}
	}

	features := make([]string, 0, len(runes)-s.featureSize+1)
	for i := 0; i <= len(runes)-s.featureSize; i++ {
		feature := string(runes[i : i+s.featureSize])
		features = append(features, feature)
	}

	return features
}

// hashFeature computes FNV-64 hash of a feature
func (s *SimHashFilter) hashFeature(feature string) uint64 {
	h := fnv.New64()
	_, _ = fmt.Fprint(h, feature) // nolint:errcheck // Hash.Write never returns error
	return h.Sum64()
}

// HammingDistance calculates the Hamming distance between two SimHash fingerprints
// Returns number of differing bits (0-64)
func HammingDistance(hash1, hash2 SimHashFingerprint) int {
	xor := uint64(hash1) ^ uint64(hash2)
	return bits.OnesCount64(xor)
}

// Similarity calculates similarity between two SimHash fingerprints
// Returns value between 0.0 and 1.0
func Similarity(hash1, hash2 SimHashFingerprint) float64 {
	distance := HammingDistance(hash1, hash2)
	return 1.0 - float64(distance)/64.0
}
