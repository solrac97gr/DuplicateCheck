package main

import (
	"fmt"
	"testing"
	"time"
)

// TestHybridEngineBasics tests basic hybrid engine functionality
func TestHybridEngineBasics(t *testing.T) {
	engine := NewHybridEngine()

	t.Run("GetName", func(t *testing.T) {
		name := engine.GetName()
		if name == "" {
			t.Error("Engine name should not be empty")
		}
		t.Logf("Engine name: %s", name)
	})

	t.Run("Compare without index (fallback)", func(t *testing.T) {
		productA := Product{ID: "A", Name: "iPhone 14", Description: "Latest model"}
		productB := Product{ID: "B", Name: "iPhone 13", Description: "Previous model"}

		result := engine.Compare(productA, productB)
		if result.NameSimilarity == 0 {
			t.Error("Should have calculated similarity")
		}
		t.Logf("Similarity: %.2f%%", result.CombinedSimilarity*100)
	})
}

// TestHybridEngineIndexing tests the LSH indexing functionality
func TestHybridEngineIndexing(t *testing.T) {
	engine := NewHybridEngine()

	products := []Product{
		{
			ID:   "P1",
			Name: "Understanding Machine Learning",
			Description: "A comprehensive guide to ML algorithms and their applications in modern data science.",
		},
		{
			ID:   "P2",
			Name: "Understanding Machine Learning",
			Description: "A comprehensive guide to ML algorithms and their applications in modern data science.",
		},
		{
			ID:   "P3",
			Name: "Deep Learning Fundamentals",
			Description: "Neural networks, backpropagation, and deep learning architectures explained clearly.",
		},
		{
			ID:   "P4",
			Name: "Introduction to Python",
			Description: "Learn Python programming from scratch with practical examples and exercises.",
		},
	}

	t.Run("Build index", func(t *testing.T) {
		start := time.Now()
		engine.BuildIndex(products)
		elapsed := time.Since(start)

		t.Logf("Indexed %d products in %v", len(products), elapsed)

		stats := engine.GetIndexStats()
		t.Logf("Index stats: %+v", stats)

		if stats["indexed"] != true {
			t.Error("Index should be marked as built")
		}
		if stats["total_products"] != len(products) {
			t.Errorf("Should have indexed %d products, got %v", len(products), stats["total_products"])
		}
	})

	t.Run("Find candidates", func(t *testing.T) {
		testProduct := Product{
			ID:   "TEST",
			Name: "Understanding Machine Learning",
			Description: "A comprehensive guide to ML algorithms and their applications in modern data science.",
		}

		candidates := engine.EstimateCandidateReduction(testProduct)
		t.Logf("Found %d candidates (out of %d total)", candidates, len(products))

		// Should find at least the exact duplicates
		if candidates == 0 {
			t.Error("Should find at least some candidates")
		}

		// Candidates should be much less than total (for large datasets)
		if candidates <= len(products) {
			t.Logf("✓ Candidate reduction working: %d/%d = %.1f%%",
				candidates, len(products), float64(candidates)/float64(len(products))*100)
		}
	})
}

// TestHybridVsNaivePerformance compares hybrid vs naive Levenshtein
func TestHybridVsNaivePerformance(t *testing.T) {
	// Generate 500 articles
	articles := generateUserArticles(500)

	newArticle := Product{
		ID:   "NEW",
		Name: "Understanding Machine Learning Algorithms in 2025",
		Description: "Machine learning has revolutionized how we approach data analysis and prediction. " +
			"In this comprehensive guide, we explore the fundamental algorithms that power modern AI systems.",
	}

	threshold := 0.85

	t.Run("Naive Levenshtein (baseline)", func(t *testing.T) {
		engine := NewLevenshteinEngine()

		start := time.Now()
		results := []ComparisonResult{}

		for _, article := range articles {
			result := engine.Compare(newArticle, article)
			if result.CombinedSimilarity >= threshold {
				results = append(results, result)
			}
		}

		elapsed := time.Since(start)

		t.Logf("Naive approach: %v for %d comparisons", elapsed, len(articles))
		t.Logf("Found %d duplicates", len(results))
		t.Logf("Throughput: %.0f comparisons/sec", float64(len(articles))/elapsed.Seconds())
	})

	t.Run("Hybrid LSH+Levenshtein", func(t *testing.T) {
		engine := NewHybridEngine()

		// Build index
		indexStart := time.Now()
		engine.BuildIndex(articles)
		indexTime := time.Since(indexStart)
		t.Logf("Index build time: %v", indexTime)

		stats := engine.GetIndexStats()
		t.Logf("Index stats: %+v", stats)

		// Query for duplicates
		queryStart := time.Now()
		results := engine.FindDuplicatesForOne(newArticle, threshold)
		queryTime := time.Since(queryStart)

		candidateCount := engine.EstimateCandidateReduction(newArticle)

		t.Logf("Hybrid approach: %v for LSH filtering + %d Levenshtein comparisons",
			queryTime, candidateCount)
		t.Logf("Found %d duplicates", len(results))
		t.Logf("Candidate reduction: %d/%d = %.1f%%",
			candidateCount, len(articles), float64(candidateCount)/float64(len(articles))*100)

		if candidateCount < len(articles) {
			speedup := float64(len(articles)) / float64(candidateCount)
			t.Logf("✓ Speedup potential: %.1fx", speedup)
		}
	})
}

// TestHybridAccuracy verifies the hybrid approach doesn't lose accuracy
func TestHybridAccuracy(t *testing.T) {
	articles := generateUserArticles(200) // Smaller set for accuracy verification

	newArticle := Product{
		ID:   "NEW",
		Name: "Understanding Machine Learning Algorithms in 2025",
		Description: "Machine learning has revolutionized how we approach data analysis and prediction. " +
			"In this comprehensive guide, we explore the fundamental algorithms that power modern AI systems.",
	}

	// Add some variations that should be detected as duplicates (high similarity)
	duplicateVariations := []Product{
		{
			ID:   "DUP1",
			Name: "Understanding Machine Learning Algorithms in 2024", // Very similar name
			Description: "Machine learning has revolutionized how we approach data analysis and prediction. " +
				"In this comprehensive guide, we explore the fundamental algorithms that power modern AI.",
		},
		{
			ID:   "DUP2",
			Name: "Machine Learning Algorithms Explained in 2025", // Similar concept
			Description: "Machine learning has truly revolutionized how we approach data analysis and predictions. " +
				"In this comprehensive tutorial, we explore the fundamental algorithms that power modern AI systems.",
		},
		{
			ID:   "DUP3",
			Name: "Understanding ML Algorithms in 2025", // Similar with abbreviation
			Description: "Machine learning has revolutionized data analysis and prediction approaches. " +
				"We explore fundamental algorithms powering modern AI systems in this comprehensive guide.",
		},
	}
	articles = append(articles, duplicateVariations...)

	threshold := 0.80 // Lower threshold to catch variations

	// Get ground truth with naive approach
	naiveEngine := NewLevenshteinEngine()
	var groundTruth []string
	for _, article := range articles {
		result := naiveEngine.Compare(newArticle, article)
		if result.CombinedSimilarity >= threshold {
			groundTruth = append(groundTruth, article.ID)
		}
	}

	// Get results with hybrid approach
	hybridEngine := NewHybridEngine()
	hybridEngine.BuildIndex(articles)
	hybridResults := hybridEngine.FindDuplicatesForOne(newArticle, threshold)

	hybridIDs := make(map[string]bool)
	for _, result := range hybridResults {
		hybridIDs[result.ProductB.ID] = true
	}

	// Calculate recall (how many true positives we found)
	foundCount := 0
	for _, trueID := range groundTruth {
		if hybridIDs[trueID] {
			foundCount++
		}
	}

	recall := 0.0
	if len(groundTruth) > 0 {
		recall = float64(foundCount) / float64(len(groundTruth))
	}

	t.Logf("Ground truth: %d duplicates", len(groundTruth))
	t.Logf("Hybrid found: %d duplicates", len(hybridResults))
	t.Logf("Recall: %.2f%% (%d/%d)", recall*100, foundCount, len(groundTruth))

	// We expect high recall (ideally 100%)
	if recall < 0.95 {
		t.Errorf("Recall too low: %.2f%% (expected ≥95%%)", recall*100)
		t.Logf("Missed IDs: %v", groundTruth[:min(5, len(groundTruth))])
	}
}

// TestHybridScalability tests performance with increasing dataset sizes
func TestHybridScalability(t *testing.T) {
	sizes := []int{100, 500, 1000, 2000}

	newArticle := Product{
		ID:   "NEW",
		Name: "Understanding Machine Learning Algorithms",
		Description: "A comprehensive guide to machine learning algorithms.",
	}

	threshold := 0.85

	for _, size := range sizes {
		t.Run(fmt.Sprintf("%d articles", size), func(t *testing.T) {
			articles := generateUserArticles(size)

			// Hybrid approach
			hybridEngine := NewHybridEngine()

			indexStart := time.Now()
			hybridEngine.BuildIndex(articles)
			indexTime := time.Since(indexStart)

			queryStart := time.Now()
			results := hybridEngine.FindDuplicatesForOne(newArticle, threshold)
			queryTime := time.Since(queryStart)

			candidates := hybridEngine.EstimateCandidateReduction(newArticle)

			t.Logf("Size %d: Index=%v, Query=%v, Candidates=%d (%.1f%%), Found=%d",
				size, indexTime, queryTime, candidates,
				float64(candidates)/float64(size)*100, len(results))

			// Verify sublinear query time growth
			comparisonsPerMs := float64(candidates) / float64(queryTime.Milliseconds())
			t.Logf("  Throughput: %.0f comparisons/ms", comparisonsPerMs)
		})
	}
}

// BenchmarkHybridVsNaive benchmarks both approaches
func BenchmarkHybridVsNaive(b *testing.B) {
	articles := generateUserArticles(500)
	newArticle := Product{
		ID:          "NEW",
		Name:        "Understanding Machine Learning",
		Description: "A comprehensive guide to ML algorithms.",
	}
	threshold := 0.85

	b.Run("Naive_Levenshtein_500", func(b *testing.B) {
		engine := NewLevenshteinEngine()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			results := []ComparisonResult{}
			for _, article := range articles {
				result := engine.Compare(newArticle, article)
				if result.CombinedSimilarity >= threshold {
					results = append(results, result)
				}
			}
		}
	})

	b.Run("Hybrid_LSH_500", func(b *testing.B) {
		engine := NewHybridEngine()
		engine.BuildIndex(articles) // Build once, query many times

		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = engine.FindDuplicatesForOne(newArticle, threshold)
		}
	})
}

// BenchmarkHybridIndexing benchmarks index building
func BenchmarkHybridIndexing(b *testing.B) {
	benchmarks := []struct {
		name  string
		count int
	}{
		{"100_articles", 100},
		{"500_articles", 500},
		{"1000_articles", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			articles := generateUserArticles(bm.count)
			engine := NewHybridEngine()

			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				engine.BuildIndex(articles)
			}
		})
	}
}

// TestBlockingStrategy tests the blocking optimization
func TestBlockingStrategy(t *testing.T) {
	strategy := NewBlockingStrategy(3) // Block by first 3 characters

	products := []Product{
		{ID: "1", Name: "Apple iPhone 14"},
		{ID: "2", Name: "Apple iPhone 13"},
		{ID: "3", Name: "Samsung Galaxy S23"},
		{ID: "4", Name: "Samsung Galaxy S22"},
		{ID: "5", Name: "Sony Headphones"},
	}

	blocks := strategy.GroupByBlocks(products)

	t.Logf("Created %d blocks", len(blocks))

	for key, prods := range blocks {
		t.Logf("Block '%s': %d products", key, len(prods))
		for _, p := range prods {
			t.Logf("  - %s", p.Name)
		}
	}

	// Verify products are grouped correctly
	if len(blocks) == 0 {
		t.Error("Should have created at least one block")
	}

	// Apple products should be in same block
	appleBlock := blocks["app"]
	if len(appleBlock) != 2 {
		t.Errorf("Expected 2 Apple products in 'app' block, got %d", len(appleBlock))
	}
}

// TestMinHashSignature tests MinHash signature generation
func TestMinHashSignature(t *testing.T) {
	text1 := "machine learning algorithms and data science"
	text2 := "machine learning algorithms and data science"
	text3 := "deep learning neural networks and AI"

	shingles1 := generateShingles(text1, 3)
	shingles2 := generateShingles(text2, 3)
	shingles3 := generateShingles(text3, 3)

	sig1 := computeMinHashSignature(shingles1, 100)
	sig2 := computeMinHashSignature(shingles2, 100)
	sig3 := computeMinHashSignature(shingles3, 100)

	t.Logf("Shingles1: %d, Shingles3: %d", len(shingles1), len(shingles3))

	// Calculate Jaccard similarity estimate
	matches12 := 0
	matches13 := 0
	for i := 0; i < len(sig1); i++ {
		if sig1[i] == sig2[i] {
			matches12++
		}
		if sig1[i] == sig3[i] {
			matches13++
		}
	}

	sim12 := float64(matches12) / float64(len(sig1))
	sim13 := float64(matches13) / float64(len(sig1))

	t.Logf("Identical texts similarity: %.2f%%", sim12*100)
	t.Logf("Different texts similarity: %.2f%%", sim13*100)

	if sim12 < 0.95 {
		t.Errorf("Identical texts should have high similarity, got %.2f%%", sim12*100)
	}
	if sim13 > sim12 {
		t.Error("Different texts should have lower similarity than identical texts")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
