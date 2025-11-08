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
	if len(os.Args) != 4 {
		fmt.Println("Error: compare requires exactly 2 product names")
		fmt.Println("Usage: duplicatecheck compare <product1> <product2>")
		os.Exit(1)
	}

	productA := Product{ID: "A", Name: os.Args[2]}
	productB := Product{ID: "B", Name: os.Args[3]}

	// Test with Levenshtein algorithm
	fmt.Println("🔍 Comparing Products")
	fmt.Println("=====================")
	fmt.Printf("Product A: %q\n", productA.Name)
	fmt.Printf("Product B: %q\n", productB.Name)
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

func handleFindDuplicates() {
	// Sample ecommerce product catalog
	products := []Product{
		{ID: "P001", Name: "Apple iPhone 14 Pro Max 256GB Silver"},
		{ID: "P002", Name: "Apple iPhone 14 Pro Max 256GB silver"}, // Duplicate (case difference)
		{ID: "P003", Name: "Apple iPhone 14 Pro 128GB Gold"},
		{ID: "P004", Name: "Apple iPhone 13 Pro 128GB Gold"},
		{ID: "P005", Name: "Samsung Galaxy S23 Ultra 512GB Black"},
		{ID: "P006", Name: "Samsung Galaxy S23 Ultra 512GB Phantom Black"},
		{ID: "P007", Name: "Samsung Galaxy S22 Ultra 512GB Black"},
		{ID: "P008", Name: "Sony WH-1000XM5 Wireless Headphones Black"},
		{ID: "P009", Name: "Sony WH-1000XM4 Wireless Headphones Black"},
		{ID: "P010", Name: "Apple MacBook Pro 16 inch M2 Pro"},
		{ID: "P011", Name: "Dell XPS 15 Laptop"},
		{ID: "P012", Name: "Apple AirPods Pro 2nd Generation"},
		{ID: "P013", Name: "Apple Airpods Pro 2nd Gen"}, // Duplicate (spacing/abbreviation)
	}

	fmt.Println("🔍 Finding Duplicate Products")
	fmt.Println("==============================")
	fmt.Printf("Analyzing %d products...\n\n", len(products))

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
			fmt.Printf("%d. Match Score: %.2f%% (Distance: %d)\n",
				i+1, dup.Similarity*100, dup.Distance)
			fmt.Printf("   Product %s: %q\n", dup.ProductA.ID, dup.ProductA.Name)
			fmt.Printf("   Product %s: %q\n", dup.ProductB.ID, dup.ProductB.Name)
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
			Product{ID: "1", Name: "iPhone"},
			Product{ID: "2", Name: "iPhone"},
			"Exact Match",
		},
		{
			Product{ID: "1", Name: "iPhone 14"},
			Product{ID: "2", Name: "iPhone 13"},
			"One Character Difference",
		},
		{
			Product{ID: "1", Name: "Samsung Galaxy S23"},
			Product{ID: "2", Name: "Samsung Galaxy S23 Ultra"},
			"Substring/Addition",
		},
		{
			Product{ID: "1", Name: "Nike Air Max"},
			Product{ID: "2", Name: "Adidas Ultraboost"},
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
			fmt.Printf("    Distance:   %d edits\n", result.Distance)
			fmt.Printf("    Similarity: %.2f%% ", result.Similarity*100)
			printSimilarityBar(result.Similarity)
			fmt.Println()

			interpretSimilarity(result.Similarity)
			fmt.Println()
		}
		fmt.Println()
	}
}

func printComparisonResult(engine DuplicateCheckEngine, result ComparisonResult) {
	fmt.Printf("Algorithm: %s\n", engine.GetName())
	fmt.Println(strings.Repeat("-", 50))
	fmt.Printf("Edit Distance:  %d\n", result.Distance)
	fmt.Printf("Similarity:     %.2f%% ", result.Similarity*100)
	printSimilarityBar(result.Similarity)
	fmt.Println()
	interpretSimilarity(result.Similarity)
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
