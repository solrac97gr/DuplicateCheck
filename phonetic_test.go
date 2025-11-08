package duplicatecheck

import "testing"

func TestSoundexCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"Robert", "Robert", "R163"},
		{"Rubin", "Rubin", "R150"},
		{"Ashcraft", "Ashcraft", "A226"}, // A-sh(2)-cr(2)-a-ft(1) -> A226
		{"Pfister", "Pfister", "P236"},
		{"Jackson", "Jackson", "J250"},
		{"Tymczak", "Tymczak", "T522"},

		// Product/Brand names - these have same Soundex which is what we want
		{"iPhone", "iPhone", "I150"}, // I-ph(1)-o-n(5)-e
		{"IPhone", "IPhone", "I150"},
		{"iFone", "iFone", "I150"},
		{"Samsung", "Samsung", "S525"}, // S-a-m(5)-s(2)-u-n(5)-g
		{"Samsong", "Samsong", "S525"},
		{"Samsoong", "Samsoong", "S525"},
		{"Lenovo", "Lenovo", "L510"}, // L-e-n(5)-o-v(1)-o
		{"Lenova", "Lenova", "L510"},

		// Edge cases
		{"Empty", "", ""},
		{"Single letter", "A", "A000"},
		{"Single consonant", "B", "B000"},
		{"Numbers and spaces", "  ABC123  ", "A120"},
		{"All vowels", "AEIOU", "A000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SoundexCode(tt.input)
			if result != tt.expected {
				t.Errorf("SoundexCode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestPhoneticFilter(t *testing.T) {
	filter := NewPhoneticFilter()

	tests := []struct {
		name      string
		nameA     string
		nameB     string
		shouldMatch bool
	}{
		// Same pronunciation variants - should potentially match
		{"iPhone variants", "iPhone", "IPhone", true},
		{"iPhone variants 2", "iPhone", "iFone", true},
		{"Samsung variants", "Samsung", "Samsong", true},
		{"Lenovo variants", "Lenovo", "Lenova", true},

		// Different pronunciations - should NOT match
		{"Different brands 1", "Samsung", "iPhone", false},
		{"Different brands 2", "Lenovo", "Dell", false},
		{"Robert vs Alice", "Robert", "Alice", false},

		// Edge cases
		{"Empty string A", "", "Samsung", true},
		{"Empty string B", "Samsung", "", true},
		{"Both empty", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filter.MaybeMatch(tt.nameA, tt.nameB)
			if result != tt.shouldMatch {
				t.Errorf("MaybeMatch(%q, %q) = %v, want %v",
					tt.nameA, tt.nameB, result, tt.shouldMatch)
			}
		})
	}
}

func TestPhoneticFilterEnableDisable(t *testing.T) {
	filter := NewPhoneticFilter()

	// Initially enabled
	if !filter.IsEnabled() {
		t.Error("Filter should be enabled by default")
	}

	// When disabled, should always return true (skip filtering)
	filter.Disable()
	if filter.IsEnabled() {
		t.Error("Filter should be disabled")
	}
	if !filter.MaybeMatch("Samsung", "iPhone") {
		t.Error("When disabled, MaybeMatch should return true")
	}

	// When re-enabled, should filter again
	filter.Enable()
	if !filter.IsEnabled() {
		t.Error("Filter should be enabled")
	}
	if filter.MaybeMatch("Samsung", "iPhone") {
		t.Error("When enabled, different Soundex codes should return false")
	}
}

func BenchmarkSoundexCode(b *testing.B) {
	testStrings := []string{
		"iPhone",
		"Samsung",
		"Apple",
		"Microsoft",
		"Google",
		"Robert",
		"Ashcraft",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, s := range testStrings {
			_ = SoundexCode(s)
		}
	}
}

func BenchmarkPhoneticFilter(b *testing.B) {
	filter := NewPhoneticFilter()
	pairs := []struct {
		a, b string
	}{
		{"iPhone", "IPhone"},
		{"Samsung", "Samsong"},
		{"Lenovo", "Lenova"},
		{"Apple", "Apples"},
		{"Robert", "Alice"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range pairs {
			_ = filter.MaybeMatch(p.a, p.b)
		}
	}
}
