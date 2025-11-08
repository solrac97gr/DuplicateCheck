package duplicatecheck

import (
	"fmt"
	"strings"
	"testing"
)

// TestUserArticleDuplicationScenario simulates a real-world scenario where we need to
// check if a user's new article is a duplicate of any of their existing articles.
// This is common for:
// - Detecting self-plagiarism
// - Preventing accidental re-submission of content
// - Finding similar drafts or variations
func TestUserArticleDuplicationScenario(t *testing.T) {
	engine := NewLevenshteinEngine()

	// Simulate a user with 500 existing articles
	userArticles := generateUserArticles(500)

	// New article the user is trying to submit
	newArticle := Product{
		ID:   "NEW_ARTICLE",
		Name: "Understanding Machine Learning Algorithms in 2025",
		Description: "Machine learning has revolutionized how we approach data analysis and prediction. " +
			"In this comprehensive guide, we explore the fundamental algorithms that power modern AI systems. " +
			"From supervised learning techniques like linear regression and decision trees to unsupervised methods " +
			"such as clustering and dimensionality reduction, this article covers everything you need to know. " +
			"We'll also dive into neural networks, deep learning architectures, and the latest advances in " +
			"transformer models that have changed natural language processing forever. " +
			"Whether you're a beginner or experienced practitioner, this guide will help you understand " +
			"the mathematical foundations and practical applications of these powerful tools.",
	}

	t.Run("Check new article against user's 500 existing articles", func(t *testing.T) {
		threshold := 0.85 // 85% similarity = likely duplicate
		duplicatesFound := 0
		var matchedArticles []ComparisonResult

		// Compare new article against all existing articles
		for _, existingArticle := range userArticles {
			result := engine.Compare(newArticle, existingArticle)

			if result.CombinedSimilarity >= threshold {
				duplicatesFound++
				matchedArticles = append(matchedArticles, result)
			}
		}

		// Log results
		t.Logf("Checked 1 new article against %d existing articles", len(userArticles))
		t.Logf("Found %d potential duplicates (≥%.0f%% similar)", duplicatesFound, threshold*100)

		if duplicatesFound > 0 {
			t.Logf("Top matches:")
			for i, match := range matchedArticles {
				if i >= 5 { // Show top 5
					break
				}
				t.Logf("  - %s: %.2f%% similar (Name: %.2f%%, Desc: %.2f%%)",
					match.ProductB.ID,
					match.CombinedSimilarity*100,
					match.NameSimilarity*100,
					match.DescriptionSimilarity*100)
			}
		}

		// Should find at least one near-duplicate (we planted one)
		if duplicatesFound == 0 {
			t.Error("Expected to find at least one duplicate in the test data")
		}
	})

	t.Run("Performance: Time to scan 500 articles", func(t *testing.T) {
		// Measure performance
		comparisons := 0
		for range userArticles {
			engine.Compare(newArticle, userArticles[comparisons])
			comparisons++
		}
		t.Logf("Completed %d comparisons successfully", comparisons)
	})
}

// TestBulkUserArticleScanning tests scanning multiple new articles against a user's catalog
func TestBulkUserArticleScanning(t *testing.T) {
	engine := NewLevenshteinEngine()

	// User has 500 existing articles
	existingArticles := generateUserArticles(500)

	// User is submitting 10 new articles
	newArticles := []Product{
		{
			ID:   "BATCH_001",
			Name: "Top 10 JavaScript Frameworks in 2025",
			Description: "JavaScript frameworks have evolved significantly. React continues to dominate, " +
				"but newer frameworks like Solid.js and Qwik are gaining traction. " +
				"This article compares their features, performance, and use cases.",
		},
		{
			ID:   "BATCH_002",
			Name: "Python vs Go: Which Language for Backend Development",
			Description: "Choosing the right backend language is crucial for project success. " +
				"Python offers simplicity and rich libraries, while Go provides speed and concurrency. " +
				"We analyze both languages across multiple dimensions.",
		},
		{
			ID:   "BATCH_003",
			Name: "Understanding Machine Learning Algorithms in 2025", // DUPLICATE!
			Description: "Machine learning has revolutionized how we approach data analysis and prediction. " +
				"In this comprehensive guide, we explore the fundamental algorithms that power modern AI systems. " +
				"From supervised learning techniques like linear regression and decision trees to unsupervised methods " +
				"such as clustering and dimensionality reduction, this article covers everything you need to know.",
		},
	}

	t.Run("Scan 10 new articles against 500 existing articles", func(t *testing.T) {
		threshold := 0.85
		totalComparisons := 0
		articlesWithDuplicates := 0

		for _, newArticle := range newArticles {
			hasDuplicate := false

			for _, existing := range existingArticles {
				result := engine.Compare(newArticle, existing)
				totalComparisons++

				if result.CombinedSimilarity >= threshold {
					hasDuplicate = true
					t.Logf("Duplicate found: %s matches %s (%.2f%% similar)",
						newArticle.ID, existing.ID, result.CombinedSimilarity*100)
					break // Found duplicate, move to next article
				}
			}

			if hasDuplicate {
				articlesWithDuplicates++
			}
		}

		t.Logf("Scanned %d new articles against %d existing articles", len(newArticles), len(existingArticles))
		t.Logf("Total comparisons: %d", totalComparisons)
		t.Logf("Articles with duplicates: %d", articlesWithDuplicates)

		// Should find at least the one we planted
		if articlesWithDuplicates == 0 {
			t.Error("Expected to find at least one duplicate")
		}
	})
}

// BenchmarkUserArticleScanning benchmarks the realistic scenario of checking
// a new article against a user's existing catalog
func BenchmarkUserArticleScanning(b *testing.B) {
	engine := NewLevenshteinEngine()

	benchmarks := []struct {
		name         string
		articleCount int
	}{
		{"100 articles", 100},
		{"500 articles", 500},
		{"1000 articles", 1000},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			userArticles := generateUserArticles(bm.articleCount)
			newArticle := Product{
				ID:   "NEW_ARTICLE",
				Name: "Understanding Machine Learning Algorithms",
				Description: "Machine learning has revolutionized how we approach data analysis. " +
					"This comprehensive guide explores the fundamental algorithms that power modern AI systems.",
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				for _, existing := range userArticles {
					engine.Compare(newArticle, existing)
				}
			}
		})
	}
}

// generateUserArticles creates a realistic set of user articles for testing
func generateUserArticles(count int) []Product {
	articles := make([]Product, count)

	// Article topics and templates
	topics := []string{
		"Understanding",
		"Complete Guide to",
		"Introduction to",
		"Advanced Techniques in",
		"Best Practices for",
		"How to Master",
		"Deep Dive into",
		"Exploring",
	}

	subjects := []string{
		"Machine Learning",
		"Web Development",
		"Cloud Computing",
		"Data Science",
		"Artificial Intelligence",
		"Blockchain Technology",
		"Cybersecurity",
		"DevOps",
		"Mobile Development",
		"Database Design",
		"API Development",
		"Microservices Architecture",
		"Container Orchestration",
		"Serverless Computing",
		"GraphQL",
	}

	years := []string{"2023", "2024", "2025"}

	// Generate diverse articles
	for i := 0; i < count; i++ {
		topicIdx := i % len(topics)
		subjectIdx := (i / len(topics)) % len(subjects)
		yearIdx := i % len(years)

		title := fmt.Sprintf("%s %s in %s",
			topics[topicIdx],
			subjects[subjectIdx],
			years[yearIdx])

		// Generate description with variation
		description := generateArticleDescription(subjects[subjectIdx], i)

		// Add a near-duplicate for testing (article #250)
		if i == 250 {
			title = "Understanding Machine Learning Algorithms in 2025"
			description = "Machine learning has revolutionized how we approach data analysis and prediction. " +
				"In this comprehensive guide, we explore the fundamental algorithms that power modern AI systems. " +
				"From supervised learning techniques like linear regression and decision trees to unsupervised methods " +
				"such as clustering and dimensionality reduction, this article covers everything you need to know."
		}

		articles[i] = Product{
			ID:          fmt.Sprintf("ARTICLE_%04d", i+1),
			Name:        title,
			Description: description,
		}
	}

	return articles
}

// generateArticleDescription creates varied article descriptions
func generateArticleDescription(subject string, seed int) string {
	templates := []string{
		"%s has become increasingly important in modern software development. " +
			"This article explores the key concepts, best practices, and real-world applications. " +
			"We cover everything from basic principles to advanced techniques that professionals use daily. " +
			"Whether you're just starting out or looking to deepen your expertise, this guide provides " +
			"valuable insights and practical examples. Learn how to apply these concepts in your projects " +
			"and stay ahead of the curve in this rapidly evolving field.",

		"In the ever-changing landscape of technology, %s stands out as a critical skill. " +
			"This comprehensive guide breaks down complex topics into digestible sections. " +
			"We examine industry trends, common challenges, and proven solutions that work. " +
			"Through detailed examples and step-by-step tutorials, you'll gain hands-on experience. " +
			"Discover tools, frameworks, and methodologies that leading companies use to build scalable solutions.",

		"Master %s with this in-depth tutorial covering fundamentals to advanced concepts. " +
			"We've compiled insights from industry experts and real-world case studies. " +
			"Learn optimization techniques, performance best practices, and security considerations. " +
			"This guide includes code samples, architectural patterns, and troubleshooting tips. " +
			"Perfect for developers looking to enhance their skills and build production-ready applications.",

		"Explore the world of %s through practical examples and clear explanations. " +
			"This article demystifies complex concepts and provides actionable knowledge. " +
			"From setup and configuration to deployment and monitoring, we cover the complete lifecycle. " +
			"Understand trade-offs, make informed decisions, and avoid common pitfalls. " +
			"Includes comparison with alternatives and recommendations for different use cases.",
	}

	templateIdx := seed % len(templates)
	description := fmt.Sprintf(templates[templateIdx], subject)

	// Add some variation based on seed
	if seed%3 == 0 {
		description += " Updated with the latest features and industry standards. " +
			"Includes bonus section on emerging trends and future predictions."
	} else if seed%5 == 0 {
		description += " Features interviews with senior engineers and technical leaders. " +
			"Real production examples from Fortune 500 companies."
	}

	return description
}

// TestUserArticleWithCustomWeights tests article comparison with different weighting strategies
func TestUserArticleWithCustomWeights(t *testing.T) {
	article1 := Product{
		ID:          "ARTICLE_A",
		Name:        "Introduction to Docker Containers and Kubernetes",
		Description: strings.Repeat("Docker revolutionizes application deployment by providing lightweight containerization. ", 20),
	}

	article2 := Product{
		ID:          "ARTICLE_B",
		Name:        "Introduction to Docker Containers and Microservices",                                                     // Very similar title
		Description: strings.Repeat("Kubernetes orchestrates container deployments at scale in production environments. ", 20), // Different content
	}

	t.Run("Title-focused weighting (90/10)", func(t *testing.T) {
		// For news sites where headlines are critical
		weights := ComparisonWeights{
			NameWeight:        0.9,
			DescriptionWeight: 0.1,
		}
		engine := NewLevenshteinEngineWithWeights(weights)
		result := engine.CompareWithWeights(article1, article2, weights)

		t.Logf("Title-focused: %.2f%% similar", result.CombinedSimilarity*100)
		t.Logf("  Name: %.2f%%, Description: %.2f%%",
			result.NameSimilarity*100, result.DescriptionSimilarity*100)
		// Should be high because titles are very similar
		if result.CombinedSimilarity < 0.75 {
			t.Errorf("Expected higher similarity with title focus, got %.2f%%", result.CombinedSimilarity*100)
		}
	})

	t.Run("Content-focused weighting (30/70)", func(t *testing.T) {
		// For academic papers where content matters more
		weights := ComparisonWeights{
			NameWeight:        0.3,
			DescriptionWeight: 0.7,
		}
		engine := NewLevenshteinEngineWithWeights(weights)
		result := engine.CompareWithWeights(article1, article2, weights)

		t.Logf("Content-focused: %.2f%% similar", result.CombinedSimilarity*100)
		// Should be lower because content is different
		if result.CombinedSimilarity > 0.50 {
			t.Errorf("Expected lower similarity with content focus, got %.2f%%", result.CombinedSimilarity*100)
		}
	})

	t.Run("Balanced weighting (50/50)", func(t *testing.T) {
		weights := ComparisonWeights{
			NameWeight:        0.5,
			DescriptionWeight: 0.5,
		}
		engine := NewLevenshteinEngineWithWeights(weights)
		result := engine.CompareWithWeights(article1, article2, weights)

		t.Logf("Balanced: %.2f%% similar", result.CombinedSimilarity*100)
		t.Logf("  Name: %.2f%%, Description: %.2f%%",
			result.NameSimilarity*100, result.DescriptionSimilarity*100)
	})
}
