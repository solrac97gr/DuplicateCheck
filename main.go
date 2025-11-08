package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	command := os.Args[1]

	switch command {
	case "compare":
		handleCompare()
	case "find":
		handleFindDuplicates()
	case "demo":
		handleDemo()
	default:
		fmt.Printf("Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("DuplicateCheck - Product Similarity Detection Tool")
	fmt.Println("==================================================")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  duplicatecheck compare <product1> <product2>")
	fmt.Println("    Compare two product names and show similarity metrics")
	fmt.Println()
	fmt.Println("  duplicatecheck find")
	fmt.Println("    Find potential duplicates in a sample product catalog")
	fmt.Println()
	fmt.Println("  duplicatecheck demo")
	fmt.Println("    Run a demonstration showing how different algorithms work")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  duplicatecheck compare \"Apple iPhone 14\" \"Apple iPhone 13\"")
	fmt.Println("  duplicatecheck find")
	fmt.Println("  duplicatecheck demo")
}

func handleCompare() {
	if len(os.Args) < 4 || len(os.Args) > 6 {
		fmt.Println("Error: compare requires 2 product names and optionally 2 descriptions")
		fmt.Println("Usage: duplicatecheck compare <name1> <name2> [description1] [description2]")
		fmt.Println()
		fmt.Println("Examples:")
		fmt.Println("  duplicatecheck compare \"Apple iPhone 14\" \"Apple iPhone 13\"")
		fmt.Println("  duplicatecheck compare \"iPhone 14\" \"iPhone 13\" \"Latest model\" \"Previous model\"")
		os.Exit(1)
	}

	productA := Product{ID: "A", Name: os.Args[2]}
	productB := Product{ID: "B", Name: os.Args[3]}
	
	// Optional descriptions
	if len(os.Args) >= 5 {
		productA.Description = os.Args[4]
	}
	if len(os.Args) >= 6 {
		productB.Description = os.Args[5]
	}

	// Test with Levenshtein algorithm
	fmt.Println("🔍 Comparing Products")
	fmt.Println("=====================")
	fmt.Printf("Product A:\n")
	fmt.Printf("  Name: %q\n", productA.Name)
	if productA.Description != "" {
		fmt.Printf("  Description: %q\n", truncateString(productA.Description, 100))
	}
	fmt.Printf("Product B:\n")
	fmt.Printf("  Name: %q\n", productB.Name)
	if productB.Description != "" {
		fmt.Printf("  Description: %q\n", truncateString(productB.Description, 100))
	}
	fmt.Println()

	engines := []DuplicateCheckEngine{
		NewLevenshteinEngine(),
		// TODO: Add more algorithms here as we implement them
		// NewJaroWinklerEngine(),
		// NewCosineEngine(),
	}

	for _, engine := range engines {
		result := engine.Compare(productA, productB)
		printComparisonResult(engine, result)
	}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func handleFindDuplicates() {
	// Sample ecommerce product catalog with descriptions
	products := []Product{
		{
			ID:   "P001",
			Name: "Apple iPhone 14 Pro Max 256GB Silver",
			Description: "The iPhone 14 Pro Max features a stunning Super Retina XDR display with ProMotion, " +
				"A16 Bionic chip, and an advanced triple-camera system. Available in 256GB storage capacity.",
		},
		{
			ID:   "P002",
			Name: "Apple iPhone 14 Pro Max 256GB silver",
			Description: "The iPhone 14 Pro Max features a stunning Super Retina XDR display with ProMotion, " +
				"A16 Bionic chip, and an advanced triple-camera system. Available in 256GB storage capacity.",
		},
		{
			ID:   "P003",
			Name: "Apple iPhone 14 Pro 128GB Gold",
			Description: "Experience the power of iPhone 14 Pro with 128GB storage, featuring the Dynamic Island, " +
				"always-on display, and professional camera capabilities.",
		},
		{
			ID:   "P004",
			Name: "Apple iPhone 13 Pro 128GB Gold",
			Description: "Previous generation iPhone Pro model with 128GB storage, featuring excellent performance " +
				"and camera system. Great value for money.",
		},
		{
			ID:   "P005",
			Name: "Samsung Galaxy S23 Ultra 512GB Black",
			Description: "Samsung's flagship phone with integrated S Pen, 200MP camera, and massive 512GB storage. " +
				"Perfect for power users and content creators.",
		},
		{
			ID:   "P006",
			Name: "Samsung Galaxy S23 Ultra 512GB Phantom Black",
			Description: "Samsung's flagship phone with integrated S Pen, 200MP camera, and massive 512GB storage. " +
				"Perfect for power users and content creators. Now in Phantom Black color.",
		},
		{
			ID:   "P007",
			Name: "Samsung Galaxy S22 Ultra 512GB Black",
			Description: "Last year's Samsung flagship with S Pen, excellent camera, and 512GB storage. " +
				"Still a powerful device at a great price.",
		},
		{
			ID:   "P008",
			Name: "Sony WH-1000XM5 Wireless Headphones Black",
			Description: "Industry-leading noise cancellation with Auto NC Optimizer, 30-hour battery life, " +
				"and premium comfort. The latest model from Sony.",
		},
		{
			ID:   "P009",
			Name: "Sony WH-1000XM4 Wireless Headphones Black",
			Description: "Industry-leading noise cancellation with LDAC audio codec, 30-hour battery life, " +
				"and premium comfort. Previous generation but still excellent.",
		},
		{
			ID:   "P010",
			Name: "Apple MacBook Pro 16 inch M2 Pro",
			Description: "Professional laptop with M2 Pro chip, 16-inch Liquid Retina XDR display, " +
				"and all-day battery life. Perfect for creative professionals.",
		},
		{
			ID:   "P011",
			Name: "Dell XPS 15 Laptop",
			Description: "Premium Windows laptop with Intel Core i7, 15.6-inch OLED display, and sleek design. " +
				"Great alternative to MacBook Pro.",
		},
		{
			ID:   "P012",
			Name: "Apple AirPods Pro 2nd Generation",
			Description: "Next-generation AirPods with improved active noise cancellation, adaptive audio, " +
				"and USB-C charging case. Includes multiple ear tip sizes.",
		},
		{
			ID:   "P013",
			Name: "Apple Airpods Pro 2nd Gen",
			Description: "Next-generation AirPods with improved active noise cancellation, adaptive audio, " +
				"and USB-C charging case. Includes multiple ear tip sizes.",
		},
	}

	fmt.Println("🔍 Finding Duplicate Products")
	fmt.Println("==============================")
	fmt.Printf("Analyzing %d products (with names and descriptions)...\n\n", len(products))

	threshold := 0.85 // 85% similarity threshold

	engines := []DuplicateCheckEngine{
		NewLevenshteinEngine(),
		// TODO: Add more algorithms here
	}

	for _, engine := range engines {
		fmt.Printf("Algorithm: %s\n", engine.GetName())
		fmt.Println(strings.Repeat("-", 50))

		duplicates := engine.FindDuplicates(products, threshold)

		if len(duplicates) == 0 {
			fmt.Printf("✅ No duplicates found with threshold %.2f\n\n", threshold)
			continue
		}

		fmt.Printf("⚠️  Found %d potential duplicate pair(s):\n\n", len(duplicates))

		for i, dup := range duplicates {
			fmt.Printf("%d. Combined Similarity: %.2f%%\n", i+1, dup.CombinedSimilarity*100)
			fmt.Printf("   (Name: %.2f%%, Description: %.2f%%)\n",
				dup.NameSimilarity*100, dup.DescriptionSimilarity*100)
			fmt.Printf("   Product %s: %q\n", dup.ProductA.ID, dup.ProductA.Name)
			if dup.ProductA.Description != "" {
				fmt.Printf("      Desc: %s\n", truncateString(dup.ProductA.Description, 80))
			}
			fmt.Printf("   Product %s: %q\n", dup.ProductB.ID, dup.ProductB.Name)
			if dup.ProductB.Description != "" {
				fmt.Printf("      Desc: %s\n", truncateString(dup.ProductB.Description, 80))
			}
			fmt.Println()
		}
	}
}

func handleDemo() {
	fmt.Println("🎓 DuplicateCheck Algorithm Demonstration")
	fmt.Println("=========================================")
	fmt.Println()

	// Demo products showing different similarity levels
	examples := []struct {
		productA    Product
		productB    Product
		description string
	}{
		{
			Product{ID: "1", Name: "iPhone", Description: ""},
			Product{ID: "2", Name: "iPhone", Description: ""},
			"Exact Match",
		},
		{
			Product{ID: "1", Name: "iPhone 14", Description: ""},
			Product{ID: "2", Name: "iPhone 13", Description: ""},
			"One Character Difference",
		},
		{
			Product{ID: "1", Name: "Samsung Galaxy S23", Description: ""},
			Product{ID: "2", Name: "Samsung Galaxy S23 Ultra", Description: ""},
			"Substring/Addition",
		},
		{
			Product{ID: "1", Name: "Nike Air Max", Description: ""},
			Product{ID: "2", Name: "Adidas Ultraboost", Description: ""},
			"Completely Different",
		},
	}

	engines := []DuplicateCheckEngine{
		NewLevenshteinEngine(),
		// TODO: Add more algorithms here
	}

	for _, engine := range engines {
		fmt.Printf("📊 Algorithm: %s\n", engine.GetName())
		fmt.Println(strings.Repeat("=", 60))
		fmt.Println()

		for _, example := range examples {
			fmt.Printf("Test Case: %s\n", example.description)
			fmt.Printf("  Product A: %q\n", example.productA.Name)
			fmt.Printf("  Product B: %q\n", example.productB.Name)

			result := engine.Compare(example.productA, example.productB)

			fmt.Printf("  Results:\n")
			fmt.Printf("    Name Similarity: %.2f%% ", result.NameSimilarity*100)
			printSimilarityBar(result.NameSimilarity)
			fmt.Println()

			interpretSimilarity(result.CombinedSimilarity)
			fmt.Println()
		}
		fmt.Println()
	}
}

func printComparisonResult(engine DuplicateCheckEngine, result ComparisonResult) {
	fmt.Printf("Algorithm: %s\n", engine.GetName())
	fmt.Println(strings.Repeat("-", 50))
	
	// Show detailed breakdown if descriptions are present
	if result.ProductA.Description != "" || result.ProductB.Description != "" {
		fmt.Printf("Name Similarity:        %.2f%% ", result.NameSimilarity*100)
		printSimilarityBar(result.NameSimilarity)
		fmt.Println()
		
		fmt.Printf("Description Similarity: %.2f%% ", result.DescriptionSimilarity*100)
		printSimilarityBar(result.DescriptionSimilarity)
		fmt.Println()
		
		fmt.Printf("Combined Similarity:    %.2f%% ", result.CombinedSimilarity*100)
		printSimilarityBar(result.CombinedSimilarity)
		fmt.Println()
	} else {
		// Legacy output for backward compatibility
		fmt.Printf("Name Similarity: %.2f%% ", result.NameSimilarity*100)
		printSimilarityBar(result.NameSimilarity)
		fmt.Println()
	}
	
	interpretSimilarity(result.CombinedSimilarity)
	fmt.Println()
}

func printSimilarityBar(similarity float64) {
	barLength := 30
	filled := int(similarity * float64(barLength))

	fmt.Print("[")
	for i := 0; i < barLength; i++ {
		if i < filled {
			fmt.Print("█")
		} else {
			fmt.Print("░")
		}
	}
	fmt.Print("]")
}

func interpretSimilarity(similarity float64) {
	switch {
	case similarity >= 0.95:
		fmt.Println("  Interpretation: ✅ Almost certainly duplicates (≥95%)")
	case similarity >= 0.85:
		fmt.Println("  Interpretation: ⚠️  Likely duplicates - review recommended (85-95%)")
	case similarity >= 0.70:
		fmt.Println("  Interpretation: 🔍 Possibly related - manual check needed (70-85%)")
	case similarity >= 0.50:
		fmt.Println("  Interpretation: 📊 Some similarity - probably different items (50-70%)")
	default:
		fmt.Println("  Interpretation: ❌ Very different products (<50%)")
	}
}
