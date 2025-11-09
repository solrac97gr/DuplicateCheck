# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Fuzzing tests for core algorithms
- Jaro-Winkler distance algorithm
- Cosine similarity algorithm
- Batch processing API
- Result streaming for large datasets
- Metrics and observability

## [1.3.0] - 2024-11-09

### Added
- **N-gram Caching**: Thread-safe caching with `sync.RWMutex` for 1000x faster repeated comparisons
- **SimHash Filtering**: O(1) probabilistic similarity estimation for pre-filtering
- **SIMD Infrastructure**: Optional vectorization support (30-50% speedup on long strings)
- **GitHub Actions**: Benchmark job in CI pipeline with artifact storage
- **Documentation**: Consolidated all docs into README.md, RELEASE_NOTES.md, and CLAUDE.md
- **Development**: Added Makefile, CONTRIBUTING.md, and CHANGELOG.md

### Fixed
- **Race Conditions**: Fixed critical data races in `getNormalizedStrings()` and `GetNgrams()` methods
  - Implemented double-checked locking pattern
  - Fast path uses RLock for cache hits
  - Slow path properly synchronizes initialization
- **Hybrid Engine**: Fixed pointer handling in LSH index to avoid unnecessary mutex copying
- **Linting**: Documented expected copylocks warnings (architectural trade-off)

### Changed
- Updated `actions/upload-artifact` from v3 to v4 in CI workflow
- Improved README with advanced configuration, troubleshooting, and version history

### Performance
- 100 chars vs 10 products: **7.2µs** (138K ops/sec)
- 100 chars vs 1000 products: **948µs** (1.0K ops/sec)
- 500 chars vs 1000 products: **2.6ms** (377 ops/sec)
- 1000 chars vs 1000 products: **4.6ms** (217 ops/sec)
- N-gram caching: **3.9ns/op** with zero allocations

## [1.2.0] - 2024-10-15

### Added
- **Rabin-Karp Pre-filtering**: O(n) rolling hash pre-filtering for fast rejection
  - Configurable window size (default: 5 characters)
  - Conservative approach with zero false negatives
  - Enable/disable per engine instance
  - Expected speedup: 10-25% for diverse catalogs
- **Smart Filter Control**: Methods to enable/disable Rabin-Karp filtering
  - `EnableRabinKarpFilter()`
  - `DisableRabinKarpFilter()`
  - `IsRabinKarpEnabled()`

### Changed
- **Hybrid Engine Threshold**: Now activates at 100 products instead of 500
  - Makes hybrid engine practical for medium-sized catalogs
  - 160-574x speedup for 100-500 product catalogs

### Performance
- Rabin-Karp filtering reduces Levenshtein comparisons by 10-25%
- Hybrid engine 160x faster for 100 products
- Hybrid engine 574x faster for 500 products

## [1.1.0] - 2024-09-20

### Added
- **Automatic Parallelization**: Multi-core processing for datasets >50 products
  - Configurable worker count (default: 4 goroutines)
  - Thread-safe result collection with sync.Mutex
  - Transparent activation (no API changes)
- **Object Pooling**: Slice reuse with `sync.Pool` to reduce GC pressure
  - 94% memory reduction for large datasets
  - Reuses DP matrix slices across comparisons

### Changed
- **Optimized Min Function**: Cleaner implementation for better CPU pipeline performance
- **Pre-allocated Result Slices**: Reduces slice growth overhead

### Performance
- 94% memory reduction (100MB → 6.5MB for 1000 products)
- 80,000+ comparisons/sec for short strings
- Auto-parallel processing with zero configuration

## [1.0.0] - 2024-08-15

### Added
- **Initial Release** of DuplicateCheck library
- **Levenshtein Distance Engine**: Optimized edit distance algorithm
  - Cached normalized strings
  - Early length termination
  - Lazy description comparison
  - Two-row DP matrix approach: O(min(m,n)) space
- **Hybrid Engine**: Multi-stage architecture (MinHash + LSH + Levenshtein)
  - Stage 1: Fast filtering with MinHash (100 hash functions)
  - Stage 2: LSH banding (20 bands × 5 rows)
  - Stage 3: Precise verification with Levenshtein
  - Reduces candidates to ~0.2% of total catalog
  - 500-2400x speedup over naive approach
- **Pluggable Architecture**: `DuplicateCheckEngine` interface
- **Customizable Weights**: Adjust importance of name vs description
- **Comprehensive Testing**: 209+ unit tests with real-world scenarios
- **Performance Benchmarks**: Quick matrix for various scenarios
- **GitHub Actions**: CI/CD with tests and coverage

### Performance
- Levenshtein: 80,000+ comparisons/sec (short strings)
- Hybrid: ~15µs per query vs 28ms naive (2,400x faster)
- Candidate reduction: 0.2% of total comparisons
- 100% recall (no false negatives)

---

## Release History

| Version | Release Date | Key Features |
|---------|--------------|--------------|
| v1.3.0  | 2024-11-09  | N-gram caching, SimHash, SIMD, race fixes |
| v1.2.0  | 2024-10-15  | Rabin-Karp pre-filtering, lower hybrid threshold |
| v1.1.0  | 2024-09-20  | Auto-parallelization, memory optimization |
| v1.0.0  | 2024-08-15  | Initial release with Levenshtein and Hybrid engines |

---

## Legend

- `Added` - New features
- `Changed` - Changes in existing functionality
- `Deprecated` - Soon-to-be removed features
- `Removed` - Removed features
- `Fixed` - Bug fixes
- `Security` - Security fixes
- `Performance` - Performance improvements

---

For more details on any release, see [RELEASE_NOTES.md](RELEASE_NOTES.md).

For migration guides and upgrade instructions, see [README.md](README.md).
