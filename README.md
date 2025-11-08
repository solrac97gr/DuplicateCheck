# DuplicateCheck 🔍

A high-performance product similarity detection tool for ecommerce platforms. This project implements and compares multiple string similarity algorithms to identify duplicate or near-duplicate products in your catalog.

## 🎯 Purpose

In ecommerce, duplicate product listings can:
- Confuse customers and hurt user experience
- Reduce conversion rates
- Cause inventory management issues
- Impact SEO performance

This tool helps you automatically detect potential duplicates by comparing product names using advanced string similarity algorithms.

## 🏗️ Architecture

The project is built with a pluggable architecture using the `DuplicateCheckEngine` interface. This allows you to:
- Easily add new similarity algorithms
- Compare different algorithms side-by-side
- Choose the best algorithm for your specific use case

### Current Algorithms

1. **Levenshtein Distance** (Edit Distance)
   - Measures minimum number of single-character edits (insertions, deletions, substitutions)
   - Time Complexity: O(m × n)
   - Space Complexity: O(min(m, n)) - optimized with two-row approach
   - Best for: Detecting typos, OCR errors, slight variations

### Coming Soon

- Jaro-Winkler Distance (better for short strings, prefix matching)
- Cosine Similarity (good for longer text, word-based)
- Jaccard Similarity (set-based comparison)
- Soundex/Metaphone (phonetic matching)

## 📦 Installation

```bash
# Clone the repository
git clone https://github.com/solrac97gr/DuplicateCheck.git
cd DuplicateCheck

# Build the tool
go build -o duplicatecheck

# Or run directly
go run .
```

## 🚀 Usage

### Compare Two Products

```bash
# Compare similarity between two product names
go run . compare "Apple iPhone 14 Pro" "Apple iPhone 13 Pro"
```

Output:
```
🔍 Comparing Products
=====================
Product A: "Apple iPhone 14 Pro"
Product B: "Apple iPhone 13 Pro"

Algorithm: Levenshtein Distance
--------------------------------------------------
Edit Distance:  1
Similarity:     94.74% [████████████████████████████░░]
  Interpretation: ✅ Almost certainly duplicates (≥95%)
```

### Find Duplicates in Catalog

```bash
# Scan a sample product catalog for potential duplicates
go run . find
```

This will analyze a built-in sample catalog and report all pairs with ≥85% similarity.

### Run Demo

```bash
# See how the algorithms work with various examples
go run . demo
```

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run benchmarks
go test -bench=. -benchmem

# Run specific benchmark
go test -bench=BenchmarkLevenshteinDistance -benchmem
```

### Test Results

The test suite includes:
- Unit tests for edge cases (empty strings, Unicode, case sensitivity)
- Integration tests for duplicate detection
- Benchmarks for performance testing on different string lengths
- Real-world ecommerce examples

## 📊 Algorithm Visualization

### Levenshtein Distance - How It Works

Let's transform "APPLE" into "APPL":

```
       ""  A  P  P  L
   ""   0  1  2  3  4
   A    1  0  1  2  3
   P    2  1  0  1  2
   P    3  2  1  0  1
   L    4  3  2  1  0
   E    5  4  3  2  1  ← Distance = 1
```

Each cell shows the minimum edits needed to transform:
- `cell[i,j]` = min edits to transform first i chars of "APPLE" into first j chars of "APPL"
- Final answer (bottom-right): **1 edit** (delete 'E')

### Operations:
- **Insertion**: Add a character
- **Deletion**: Remove a character  
- **Substitution**: Replace one character with another

For each cell, we choose the minimum cost:
```
cell[i,j] = min(
    cell[i-1,j] + 1,      // deletion
    cell[i,j-1] + 1,      // insertion
    cell[i-1,j-1] + cost  // substitution (cost=0 if match, 1 if different)
)
```

## 🎯 Use Cases

### 1. Data Cleaning
```go
products := loadProductsFromDatabase()
engine := NewLevenshteinEngine()
duplicates := engine.FindDuplicates(products, 0.90)
// Review and merge duplicates
```

### 2. Import Validation
```go
// Check new imports against existing catalog
for _, newProduct := range imports {
    for _, existing := range catalog {
        result := engine.Compare(newProduct, existing)
        if result.Similarity > 0.85 {
            log.Warning("Possible duplicate detected")
        }
    }
}
```

### 3. Search Enhancement
```go
// Find similar products for "did you mean?" suggestions
searchTerm := "iPone 14"  // typo
matches := findSimilarProducts(searchTerm, catalog, 0.70)
```

## 🔧 Customization

### Adjusting Similarity Threshold

```go
// Conservative (fewer false positives, might miss some duplicates)
duplicates := engine.FindDuplicates(products, 0.95)

// Balanced (recommended starting point)
duplicates := engine.FindDuplicates(products, 0.85)

// Aggressive (catch more duplicates, more false positives)
duplicates := engine.FindDuplicates(products, 0.70)
```

### Adding Your Own Algorithm

1. Implement the `DuplicateCheckEngine` interface in a new file
2. Add tests in `*_test.go`
3. Register in `main.go` engines slice

Example:
```go
type MyCustomEngine struct{}

func (e *MyCustomEngine) GetName() string {
    return "My Custom Algorithm"
}

func (e *MyCustomEngine) Compare(a, b Product) ComparisonResult {
    // Your implementation here
}

func (e *MyCustomEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
    // Your implementation here
}
```

## 📈 Performance

Benchmark results on Apple M1:

```
BenchmarkLevenshteinDistance/Short_strings_(6-7_chars)-8         3000000    450 ns/op
BenchmarkLevenshteinDistance/Medium_strings_(~20_chars)-8        1000000   1200 ns/op
BenchmarkLevenshteinDistance/Long_strings_(~50_chars)-8           500000   3100 ns/op

BenchmarkLevenshteinFindDuplicates/10_products-8                 50000     32000 ns/op
BenchmarkLevenshteinFindDuplicates/50_products-8                  5000    350000 ns/op
BenchmarkLevenshteinFindDuplicates/100_products-8                 1000   1400000 ns/op
```

**Note**: `FindDuplicates` is O(n²) - for large catalogs (>10,000 products), consider:
- Blocking/bucketing strategies (group by category, brand, price range)
- Parallel processing
- Approximate nearest neighbor algorithms

## 🤝 Contributing

This is an experimental project for testing different similarity algorithms. Feel free to:
- Add new algorithms
- Optimize existing implementations
- Add test cases
- Improve documentation

## 📝 License

MIT License - feel free to use in your projects!

## 🔗 References

- [Levenshtein Distance - Wikipedia](https://en.wikipedia.org/wiki/Levenshtein_distance)
- [String Similarity Metrics](https://en.wikipedia.org/wiki/String_metric)
- [Go Documentation](https://golang.org/doc/) 