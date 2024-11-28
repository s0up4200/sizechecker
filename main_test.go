package main

import "testing"

func TestCleanSizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "autobrr format with space",
			input:    "12.5 GB",
			expected: "12.5GB",
		},
		{
			name:     "multiple spaces",
			input:    "12.5  GB",
			expected: "12.5GB",
		},
		{
			name:     "no spaces",
			input:    "12.5GB",
			expected: "12.5GB",
		},
		{
			name:     "leading and trailing spaces",
			input:    " 12.5 GB ",
			expected: "12.5GB",
		},
		{
			name:     "megabytes unit",
			input:    "500 MB",
			expected: "500MB",
		},
		{
			name:     "terabytes unit",
			input:    "2 TB",
			expected: "2TB",
		},
		{
			name:     "kibibytes unit",
			input:    "800 KiB",
			expected: "800KiB",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanSizeString(tt.input)
			if result != tt.expected {
				t.Errorf("cleanSizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
