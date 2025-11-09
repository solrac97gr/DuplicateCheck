# Contributing to DuplicateCheck

Thank you for your interest in contributing to DuplicateCheck! This document provides guidelines and instructions for contributing.

## 📋 Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [How to Contribute](#how-to-contribute)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Pull Request Process](#pull-request-process)
- [Adding a New Algorithm](#adding-a-new-algorithm)

## 📜 Code of Conduct

This project follows a simple code of conduct:
- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community
- Show empathy towards other contributors

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Git
- Basic understanding of string similarity algorithms

### Quick Setup

1. **Fork the repository** on GitHub
2. **Clone your fork:**
   ```bash
   git clone https://github.com/YOUR_USERNAME/DuplicateCheck.git
   cd DuplicateCheck
   ```
3. **Add upstream remote:**
   ```bash
   git remote add upstream https://github.com/solrac97gr/DuplicateCheck.git
   ```

## 🛠️ Development Setup

### Install Dependencies

```bash
# Download dependencies
make deps

# Or using go directly
go mod download
```

### Verify Setup

```bash
# Run tests
make test

# Run quick benchmarks
make quick-bench
```

## 🤝 How to Contribute

### Reporting Bugs

Before creating a bug report, please check existing issues. When creating a bug report, include:

- **Clear title and description**
- **Go version** (`go version`)
- **Operating system** and architecture
- **Steps to reproduce** the issue
- **Expected vs actual behavior**
- **Sample code** if applicable

**Example:**
```markdown
**Bug:** Race condition in GetNgrams under high concurrency

**Environment:**
- Go 1.22
- macOS M1
- 1000+ concurrent goroutines

**Steps to Reproduce:**
1. Create 1000 products
2. Call GetNgrams(3) concurrently from 100 goroutines
3. Run with -race flag

**Expected:** No race conditions
**Actual:** Data race detected

**Sample Code:**
```go
// Your minimal reproduction code
```
```

### Suggesting Enhancements

Enhancement suggestions are welcome! Include:

- **Clear use case** - Why is this needed?
- **Proposed API** - How would it work?
- **Performance impact** - Expected overhead?
- **Backward compatibility** - Breaking changes?

### First Contribution?

Look for issues labeled:
- `good first issue` - Good for newcomers
- `help wanted` - We need help with these
- `documentation` - Improve documentation

## 💻 Coding Standards

### Go Code Style

Follow standard Go conventions:

```bash
# Format code
make fmt

# Run linter
make lint

# Run vet
make vet
```

### Code Organization

```
DuplicateCheck/
├── engine.go           # Core interfaces and types
├── levenshtein.go      # Levenshtein implementation
├── hybrid.go           # Hybrid (LSH+MinHash) implementation
├── *_test.go          # Tests (same package)
└── examples/          # Usage examples
```

### Naming Conventions

- **Exported types/functions:** `PascalCase`
- **Unexported types/functions:** `camelCase`
- **Interfaces:** Use `-er` suffix when appropriate (`DuplicateCheckEngine`)
- **Test functions:** `TestFeatureName` or `BenchmarkFeatureName`

### Comments

- Document all exported functions, types, and constants
- Use `godoc` format
- Include examples for complex APIs

**Example:**
```go
// Compare computes the similarity between two products.
// Returns a ComparisonResult with distance and similarity metrics.
//
// The comparison uses default weights (70% name, 30% description).
// For custom weights, use CompareWithWeights instead.
//
// Example:
//   result := engine.Compare(productA, productB)
//   fmt.Printf("Similarity: %.2f%%\n", result.CombinedSimilarity*100)
func (e *LevenshteinEngine) Compare(a, b Product) ComparisonResult {
    // ...
}
```

## 🧪 Testing Guidelines

### Test Coverage Requirements

- **Minimum coverage:** 75% (current: 76.3%)
- **New features:** 80%+ coverage required
- **Critical paths:** 90%+ coverage (comparison algorithms, caching)

### Running Tests

```bash
# All tests with race detector
make test

# Quick tests (no race detector)
make test-short

# With coverage report
make cover
```

### Writing Tests

**Structure:**
```go
func TestFeatureName(t *testing.T) {
    // Arrange
    engine := NewLevenshteinEngine()
    productA := Product{ID: "1", Name: "iPhone 14"}
    productB := Product{ID: "2", Name: "iPhone 13"}

    // Act
    result := engine.Compare(productA, productB)

    // Assert
    if result.CombinedSimilarity < 0.5 {
        t.Errorf("Expected similarity > 0.5, got %.2f", result.CombinedSimilarity)
    }
}
```

**Table-Driven Tests:**
```go
func TestLevenshteinDistance(t *testing.T) {
    tests := []struct {
        name     string
        a, b     string
        expected int
    }{
        {"exact match", "hello", "hello", 0},
        {"one insertion", "hello", "helo", 1},
        {"one deletion", "helo", "hello", 1},
        {"one substitution", "hello", "hallo", 1},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            engine := NewLevenshteinEngine()
            distance := engine.computeDistance(tt.a, tt.b)
            if distance != tt.expected {
                t.Errorf("expected %d, got %d", tt.expected, distance)
            }
        })
    }
}
```

### Benchmark Requirements

Every new algorithm or optimization must include benchmarks:

```go
func BenchmarkNewFeature(b *testing.B) {
    engine := NewLevenshteinEngine()
    productA := Product{ID: "1", Name: "Test Product"}
    productB := Product{ID: "2", Name: "Test Product 2"}

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        engine.Compare(productA, productB)
    }
}
```

## 📤 Pull Request Process

### Before Submitting

1. **Update from upstream:**
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks:**
   ```bash
   make check    # Quick check
   make ci       # Full CI checks
   ```

3. **Update documentation:**
   - Update README.md if adding features
   - Add entries to CHANGELOG.md
   - Update RELEASE_NOTES.md if needed

### PR Title Format

Use conventional commits format:

```
feat: Add Jaro-Winkler distance algorithm
fix: Race condition in GetNgrams under high concurrency
docs: Update API documentation for hybrid engine
perf: Optimize n-gram generation by 20%
test: Add fuzzing tests for Levenshtein distance
chore: Update dependencies
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix (non-breaking change which fixes an issue)
- [ ] New feature (non-breaking change which adds functionality)
- [ ] Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] Documentation update

## Changes Made
- Change 1
- Change 2

## Testing
- [ ] All tests pass (`make test`)
- [ ] Added new tests for new features
- [ ] Benchmarks included for performance changes
- [ ] Coverage not decreased

## Benchmarks (if applicable)
```
Before: 1.2ms per operation
After: 0.8ms per operation
Improvement: 33% faster
```

## Documentation
- [ ] Updated README.md
- [ ] Updated CHANGELOG.md
- [ ] Added code comments
```

### Review Process

1. **Automated checks** must pass (tests, linting)
2. **Maintainer review** (typically within 2-3 days)
3. **Address feedback** and push updates
4. **Approval and merge** by maintainer

## 🎯 Adding a New Algorithm

Want to add a new similarity algorithm? Great! Here's the process:

### 1. Implement the Interface

```go
// myalgorithm.go
type MyAlgorithmEngine struct {
    weights ComparisonWeights
}

func NewMyAlgorithmEngine() *MyAlgorithmEngine {
    return &MyAlgorithmEngine{
        weights: DefaultWeights(),
    }
}

// Implement all DuplicateCheckEngine methods
func (e *MyAlgorithmEngine) GetName() string {
    return "My Algorithm"
}

func (e *MyAlgorithmEngine) Compare(a, b Product) ComparisonResult {
    // Your implementation
}

func (e *MyAlgorithmEngine) CompareWithWeights(a, b Product, weights ComparisonWeights) ComparisonResult {
    // Your implementation
}

func (e *MyAlgorithmEngine) FindDuplicates(products []Product, threshold float64) []ComparisonResult {
    // Your implementation
}
```

### 2. Add Comprehensive Tests

```go
// myalgorithm_test.go
func TestMyAlgorithm(t *testing.T) {
    // Test exact matches
    // Test similar products
    // Test dissimilar products
    // Test edge cases (empty strings, Unicode, etc.)
}

func BenchmarkMyAlgorithm(b *testing.B) {
    // Benchmark different scenarios
}
```

### 3. Add to README

Update the "Algorithm Selection Guide" section with:
- Algorithm description
- Time/space complexity
- When to use it
- Performance comparison

### 4. Submit PR

Include benchmark comparison with existing algorithms:

```
Algorithm         | 100 products | 1000 products | Use Case
------------------|--------------|---------------|----------
Levenshtein       | 180µs        | 7.2ms         | General purpose
Hybrid            | 106µs        | 4.3ms         | Large catalogs
MyAlgorithm       | 120µs        | 5.8ms         | Special case X
```

## 📞 Getting Help

- **Questions?** Open a [discussion](https://github.com/solrac97gr/DuplicateCheck/discussions)
- **Bug reports?** Open an [issue](https://github.com/solrac97gr/DuplicateCheck/issues)
- **Want to chat?** Mention in an issue that you're working on it

## 🙏 Thank You!

Your contributions make DuplicateCheck better for everyone. Every contribution, no matter how small, is valued and appreciated!

---

**Maintained by:** Carlos García (@solrac97gr)
