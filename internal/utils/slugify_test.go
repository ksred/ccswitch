package utils

import (
	"testing"
)

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase conversion", "UPPERCASE", "uppercase"},
		{"space to dash", "hello world", "hello-world"},
		{"special chars removal", "test@#$%^&*()", "test"},
		{"multiple spaces", "too   many    spaces", "too-many-spaces"},
		{"trim dashes", "--trimmed--", "trimmed"},
		{"unicode handling", "café résumé", "caf-r-sum"},
		{"numbers preserved", "test123", "test123"},
		{"dash preserved", "already-dashed", "already-dashed"},
		{"underscore to dash", "test_underscore", "test-underscore"},
		{"mixed case and spaces", "My Feature Branch", "my-feature-branch"},
		{"empty string", "", ""},
		{"only special chars", "@#$%", ""},
		{"consecutive special chars", "test!!!branch", "test-branch"},
		{"leading/trailing spaces", "  spaced  ", "spaced"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Slugify(tt.input)
			if result != tt.expected {
				t.Errorf("Slugify(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkSlugify(b *testing.B) {
	input := "This is a Test String with Special-Characters_123"
	for i := 0; i < b.N; i++ {
		Slugify(input)
	}
}
