# DuplicateCheck 🔍

A high-performance product similarity detection tool for ecommerce platforms. This project implements and compares multiple string similarity algorithms to identify duplicate or near-duplicate products in your catalog.

## 🎯 Purpose

In ecommerce, duplicate product listings can:
- Confuse customers and hurt user experience
- Reduce conversion rates
- Cause inventory management issues
- Impact SEO performance

This tool helps you automatically detect potential duplicates by comparing **product names AND descriptions** (up to 3000 characters) using advanced string similarity algorithms with customizable weighting.

## 🏗️ Architecture

The project is built with a pluggable architecture using the `DuplicateCheckEngine` interface. This allows you to:
- Easily add new similarity algorithms
- Compare different algorithms side-by-side
- Choose the best algorithm for your specific use case
- Compare both product names and descriptions with custom weights

### Current Algorithms

1. **Levenshtein Distance** (Edit Distance)
   - Measures minimum number of single-character edits (insertions, deletions, substitutions)
   - Time Complexity: O(m × n)
   - Space Complexity: O(min(m, n)) - optimized with two-row approach
   - **Supports descriptions up to 3000+ characters efficiently**
   - Default weighting: 70% name, 30% description
   - Best for: Detecting typos, OCR errors, slight variations
   - Performance: ~3,700 comparisons/sec

2. **Hybrid (MinHash + LSH → Levenshtein)** ⚡ **RECOMMENDED FOR SCALE**
   - Multi-stage architecture for massive performance gains
   - Stage 1: MinHash (100 hash functions) + LSH (20 bands) for fast filtering
   - Stage 2: Levenshtein verification on candidate pairs only
   - **500x speedup potential** on large datasets (500+ products)
   - Time per query: **~15µs** (vs 28ms naive approach)
   - Candidate reduction: Checks only **0.2%** of total comparisons
   - Accuracy: **100% recall** (no false negatives)
   - Index build time: ~70ms for 500 products, ~145ms for 1000 products
   - Best for: Large catalogs (500+ products), 1-vs-many queries

#### Performance Comparison

| Dataset Size | Naive (ms) | Hybrid (µs) | Speedup | Candidates Checked |
|-------------|-----------|-------------|---------|-------------------|
| 500 products | 28.7 | 15.3 | **1,874x** | 1 (0.2%) |
| 1000 products | ~60 | ~25 | **2,400x** | ~0.1% |

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

## 🎯 Algorithm Selection Guide

Choose the right algorithm for your use case:

### Use **Levenshtein Engine** when:
- ✅ Small to medium datasets (<500 products)
- ✅ Maximum accuracy is critical
- ✅ You need detailed edit distance information
- ✅ One-time batch comparisons
- ✅ Real-time comparisons of 2 products

### Use **Hybrid Engine** when:
- ⚡ Large datasets (500+ products)
- ⚡ Repeated 1-vs-many queries
- ⚡ Need to check one product against entire catalog
- ⚡ Performance is critical (API/real-time scenarios)
- ⚡ Can accept one-time indexing cost (~70-150ms)
- ⚡ 500-2400x speedup needed

**Recommendation**: For catalogs >500 products, use Hybrid. The indexing time pays off after just a few queries.

## 🚀 Usage

### Compare Two Products (Names Only)

```bash
# Compare similarity between two product names
go run . compare "Apple iPhone 14 Pro" "Apple iPhone 13 Pro"
```

### Compare Products with Descriptions

```bash
# Compare with descriptions for more accurate detection
go run . compare \
  "iPhone 14 Pro" \
  "iPhone 13 Pro" \
  "Latest flagship with A16 chip and Dynamic Island" \
  "Previous generation with A15 chip"
```

Output:
```
🔍 Comparing Products
=====================
Product A:
  Name: "iPhone 14 Pro"
  Description: "Latest flagship with A16 chip and Dynamic Island"
Product B:
  Name: "iPhone 13 Pro"
  Description: "Previous generation with A15 chip"

Algorithm: Levenshtein Distance
--------------------------------------------------
Name Similarity:        92.31% [███████████████████████████░░░]
Description Similarity: 22.92% [██████░░░░░░░░░░░░░░░░░░░░░░░░]
Combined Similarity:    71.49% [█████████████████████░░░░░░░░░]
  Interpretation: 🔍 Possibly related - manual check needed (70-85%)
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

# Run specific test suites
go test -v -run TestUserArticle           # User article duplication tests
go test -bench=BenchmarkUserArticle       # Article scanning benchmarks
```

### Test Suites

The comprehensive test suite includes:

1. **Basic Algorithm Tests** (`levenshtein_test.go`)
   - Edge cases (empty strings, Unicode, case sensitivity)
   - Name and description comparison
   - Custom weight configurations

2. **User Article Duplication Tests** (`user_articles_test.go`)
   - **Real-world scenario:** Check 1 new article against 500 existing articles
   - **Batch processing:** Check 10 articles against 500 existing articles  
   - **Custom weighting:** Test different title vs. content weight strategies
   - **Performance:** ~540ms to scan 500 articles with descriptions

3. **Benchmarks**
   - String comparison (short, medium, long)
   - Description comparison (750-2000+ chars)
   - Catalog scanning (10, 50, 100 products)
   - **User article scanning (100, 500, 1000 articles)**

### Real-World Test Results

```
TestUserArticleDuplicationScenario:
  ✅ 1 article vs 500 existing: 540ms
  ✅ Found duplicate at 85.79% similarity
  
TestBulkUserArticleScanning:
  ✅ 10 articles vs 500 existing: 470ms
  ✅ Early exit optimization (stops at first duplicate)
  
BenchmarkUserArticleScanning:
  • 100 articles:  ~25ms per scan
  • 500 articles:  ~125ms per scan
  • 1000 articles: ~247ms per scan
```
- Real-world ecommerce examples

## 📊 Algorithm Visualization

### Hybrid Architecture - Multi-Stage Pipeline

The hybrid engine uses a 3-stage approach for massive speedups:

```
┌─────────────────────────────────────────────────────────────┐
│ Stage 1: Fast Filtering (MinHash + LSH)                    │
│ ─────────────────────────────────────────────────────────── │
│ Input: 1 product vs 500 catalog products                    │
│                                                              │
│ 1. Tokenize text into 3-word shingles                      │
│    "Apple iPhone 14" → ["Apple iPhone 14"]                 │
│                                                              │
│ 2. Generate MinHash signature (100 hash functions)         │
│    Text → [h1, h2, h3, ... h100]                           │
│                                                              │
│ 3. LSH Banding (20 bands × 5 rows each)                   │
│    Similar products fall into same buckets                  │
│                                                              │
│ Candidate Reduction: 500 → 1-10 candidates (0.2-2%)       │
│ Time: ~300µs                                                │
└─────────────────────────────────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│ Stage 2: Precise Verification (Levenshtein)                │
│ ─────────────────────────────────────────────────────────── │
│ Input: Only LSH candidates (1-10 products instead of 500)  │
│                                                              │
│ Run full Levenshtein Distance on:                          │
│  • Product names (weighted 70%)                            │
│  • Descriptions up to 3000 chars (weighted 30%)            │
│                                                              │
│ Time: ~15µs (vs 28ms naive approach)                       │
│ Speedup: 500-2400x faster!                                  │
└─────────────────────────────────────────────────────────────┘
                           ↓
                   Final Results
              (100% Recall, No False Negatives)
```

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

### 1. Data Cleaning with Descriptions (Small Catalog)
```go
products := loadProductsFromDatabase() // Products with names and descriptions
engine := NewLevenshteinEngine()
duplicates := engine.FindDuplicates(products, 0.90)
// Review and merge duplicates - descriptions improve accuracy!
```

### 2. Large Catalog Deduplication (500+ Products)
```go
// Use hybrid engine for massive performance gains
products := loadLargeProductCatalog() // 500-10,000 products
engine := NewHybridEngine()

// One-time indexing (only needed once or when catalog changes)
engine.BuildIndex(products) // ~70ms for 500 products, ~145ms for 1000

// Now query is lightning fast (15µs instead of 28ms per product)
newProduct := Product{
    Name: "New Product to Check",
    Description: "Full product description...",
}

duplicates := engine.FindDuplicatesForOne(newProduct, 0.85)
// Returns potential duplicates in microseconds!
// 500x faster than naive approach
```

### 3. Real-time API Endpoint
```go
// Perfect for real-time duplicate checking as users add products
var catalogEngine *HybridEngine

func init() {
    catalogEngine = NewHybridEngine()
    // Build index once at startup
    products := loadAllProducts()
    catalogEngine.BuildIndex(products)
}

func CheckDuplicateHandler(w http.ResponseWriter, r *http.Request) {
    newProduct := parseProductFromRequest(r)
    
    // Ultra-fast query: ~15µs per check
    duplicates := catalogEngine.FindDuplicatesForOne(newProduct, 0.85)
    
    json.NewEncoder(w).Encode(duplicates)
}
```

### 4. Custom Weighting for Specific Use Cases
```go
// For products where descriptions are more important (e.g., books, media)
weights := ComparisonWeights{
    NameWeight:        0.3,  // 30% on title
    DescriptionWeight: 0.7,  // 70% on description
}
engine := NewLevenshteinEngineWithWeights(weights)
result := engine.CompareWithWeights(productA, productB, weights)
```

### 3. Import Validation
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
type MyCustomEngine struct{
    weights ComparisonWeights
}

func (e *MyCustomEngine) GetName() string {
    return "My Custom Algorithm"
}

func (e *MyCustomEngine) Compare(a, b Product) ComparisonResult {
    return e.CompareWithWeights(a, b, e.weights)
}

func (e *MyCustomEngine) CompareWithWeights(a, b Product, weights ComparisonWeights) ComparisonResult {
    // Your implementation here
    // Compare both name and description
    // Return ComparisonResult with all similarity metrics
}

func (e *MyCustomEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
    // Your implementation here
}

func (e *MyCustomEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
    // Your implementation here
}
```

## 📈 Performance

Benchmark results on Intel Xeon 8370C @ 2.80GHz:

### String Comparison Performance
```
BenchmarkLevenshteinDistance/Short_strings_(6-7_chars)           280 ns/op     128 B/op
BenchmarkLevenshteinDistance/Medium_strings_(~20_chars)         1268 ns/op     368 B/op
BenchmarkLevenshteinDistance/Long_strings_(~50_chars)           8221 ns/op    1856 B/op
```

### Description Comparison Performance (with names)
```
BenchmarkLevenshteinLongDescriptions/~750_chars              2.1 ms/op    30 KB/op
BenchmarkLevenshteinLongDescriptions/~2000_chars            15.3 ms/op    80 KB/op
```

### Catalog Scanning Performance
```
BenchmarkLevenshteinFindDuplicates/10_products              118 μs/op    46 KB/op
BenchmarkLevenshteinFindDuplicates/50_products              3.1 ms/op   1.2 MB/op
BenchmarkLevenshteinFindDuplicates/100_products            13.3 ms/op   5.4 MB/op
```

**Key Insights:**
- ✅ Handles descriptions up to 3000+ chars efficiently (< 50ms per comparison)
- ✅ Memory-efficient: O(min(m,n)) space complexity
- ✅ Scales well for catalogs with 100s-1000s of products

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