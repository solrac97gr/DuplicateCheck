# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

DuplicateCheck is a high-performance Go library for detecting duplicate or near-duplicate products in ecommerce catalogs. It uses two main algorithms: an optimized Levenshtein distance engine for smaller datasets and a hybrid MinHash+LSH engine for large-scale deduplication.

## Project Structure

```
DuplicateCheck/
├── engine.go              # Core interfaces and types (Product, ComparisonResult, DuplicateCheckEngine)
├── levenshtein.go         # Levenshtein distance engine with optimizations (caching, pooling, parallelization)
├── hybrid.go              # Hybrid engine: MinHash + LSH for fast filtering, Levenshtein for verification
├── levenshtein_test.go    # Tests for Levenshtein engine
├── hybrid_test.go         # Tests for Hybrid engine
├── example_test.go        # Usage examples
├── user_articles_test.go  # Real-world scenario tests
├── quick_bench_test.go    # Comprehensive performance benchmarks with percentiles
└── README.md              # Full documentation with architecture, performance, and best practices
```

## Key Architecture Concepts

### Pluggable Engine Architecture
All engines implement the `DuplicateCheckEngine` interface in [engine.go](engine.go):
- `Compare(a, b Product) ComparisonResult` - Compare two products
- `CompareWithWeights(a, b Product, weights ComparisonWeights) ComparisonResult` - With custom weights
- `FindDuplicates(products []Product, threshold float64) []ComparisonResult` - Batch duplicate detection

### LevenshteinEngine ([levenshtein.go](levenshtein.go))
Optimized edit distance algorithm with:
- **Slice pooling** (`sync.Pool`) - Reuses DP matrix slices to reduce GC pressure
- **Normalized string caching** - Avoids repeated `ToLower()` and `TrimSpace()` in batch operations
- **Early length termination** - Skips impossible matches based on string length
- **Lazy description comparison** - Only computes descriptions when name similarity is promising
- **Automatic parallelization** - Uses goroutines for datasets >50 products (4 worker default)
- **Product normalization** - Lazy-initialized normalized versions cached on the Product struct

Performance: 80,000+ comparisons/sec for short strings, up to 411x faster than naive implementation with all optimizations.

### HybridEngine ([hybrid.go](hybrid.go))
Multi-stage architecture for massive datasets:
1. **Stage 1 (Fast Filtering)**: Generate 3-gram shingles → MinHash signatures (100 hash functions) → LSH banding (20 bands × 5 rows)
2. **Stage 2 (Candidate Selection)**: Reduces candidates to ~0.2% of total catalog
3. **Stage 3 (Verification)**: Applies optimized Levenshtein on final candidates only

Uses LSHIndex to organize products into buckets. Activates automatically at 100+ products.

### ComparisonWeights
Default weights (70% name, 30% description) can be customized:
- Name-heavy products (phones, laptops): 0.80/0.20
- Description-heavy products (books, articles): 0.40/0.60

## Common Commands

### Running Tests
```bash
# All tests
go test ./...

# With coverage
go test -cover ./...

# Verbose output
go test -v ./...

# Specific test
go test -v -run TestLevenshteinDistance

# User article tests (real-world scenario)
go test -v -run TestUserArticle
```

### Benchmarking
```bash
# Quick performance matrix (recommended!)
# Tests 100/500/1000 char descriptions across 10/100/1000 product catalogs
go test -bench=BenchmarkQuickMatrix -timeout=10m

# All benchmarks
go test -bench=. -benchmem

# Compare Hybrid vs Levenshtein
go test -bench=BenchmarkHybridVsNaive -benchtime=5s
```

### Building
```bash
# Build binary
go build -o duplicatecheck

# Build with version info
go build -ldflags "-X main.Version=1.0.0" -o duplicatecheck
```

## Performance Characteristics

### Algorithm Selection
- **<100 products**: Use LevenshteinEngine (now 400x faster with optimizations)
- **100-500 products**: Either works, Hybrid gives 160-574x speedup
- **500+ products**: Use HybridEngine (up to 2400x speedup)

### Memory Optimizations
- Object pooling via `sync.Pool` reduces allocations
- Normalized string caching prevents repeated conversions
- Two-row DP matrix approach: O(min(m,n)) space instead of O(m×n)
- Result: 94% memory reduction (100MB → 6.5MB for 1000 products)

### Key Performance Numbers
- Levenshtein: 80,000+ comparisons/sec for short strings
- Hybrid indexing: ~15ms for 100 products, ~75ms for 500 products
- Query time: ~15µs (Hybrid) vs 28ms (naive approach) for large catalogs

## Important Development Notes

### String Normalization Strategy
The Product struct includes:
- `normalizedName` and `normalizedDesc` - Cached normalized versions
- `normalized` - Flag to track if caching has been done
- `getNormalizedStrings()` - Lazy initialization method

Always call `getNormalizedStrings()` instead of normalizing on-the-fly in loops.

### Parallelization Details
- Threshold: >50 products automatically triggers parallel processing
- Workers: 4 goroutines by default
- Uses sync.Mutex for thread-safe result collection
- Safe for concurrent calls to FindDuplicates

### Hybrid Engine Index Building
- Must call `BuildIndex()` once before querying
- Index is immutable after building
- Create new HybridEngine instance if products change
- Don't rebuild index for every query (major performance anti-pattern)

### Testing Pattern
Test files follow naming conventions:
- `*_test.go` - Standard Go tests
- Example tests in [example_test.go](example_test.go) serve as documentation
- Real-world scenarios in [user_articles_test.go](user_articles_test.go)
- Performance matrices in [quick_bench_test.go](quick_bench_test.go)

## Future Performance Roadmap

See [FUTURE_IMPROVEMENTS.md](FUTURE_IMPROVEMENTS.md) for planned optimizations:
- Phase 3: SIMD/vectorization (30-50% speedup)
- Phase 3: Rabin-Karp rolling hash pre-filtering (40-60% speedup)
- Phase 4: Approximate similarity hashing
- Phase 4: Bloom filters for negative caching

Current status: 411x faster with phases 1-2 complete.

## Common Issues & Debugging

### Issue: Slow Hybrid queries
**Solution**: Ensure you're calling `BuildIndex()` once before querying multiple products, not rebuilding for each query.

### Issue: High memory usage with LevenshteinEngine
**Solution**: Consider switching to HybridEngine for catalogs >500 products. Use pooling is already optimized.

### Issue: False negatives (missing duplicates)
**Solution**: Lower your similarity threshold. Default weights are 70% name / 30% description - adjust if needed. Both engines have 100% recall.

### Issue: Performance degradation on large strings
**Solution**: For descriptions >500 chars, Hybrid engine significantly outperforms Levenshtein. The early termination optimization depends on string lengths.
