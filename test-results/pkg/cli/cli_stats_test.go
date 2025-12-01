package cli

import (
	"testing"
)

func TestParseArtifactStats(t *testing.T) {
	tests := []struct {
		name     string
		output   string
		expected ArtifactStats
	}{
		{
			name: "modern push format",
			output: `[2024-01-15 10:30:45.123] Successfully pushed artifact for current job.
[2024-01-15 10:30:45.124] * Local source: /tmp/test-results123.
[2024-01-15 10:30:45.125] * Remote destination: test-results/junit.json.
[2024-01-15 10:30:45.126] Pushed 3 files. Total of 1.5 MB`,
			expected: ArtifactStats{
				FileCount: 3,
				TotalSize: 1572864, // 1.5 * 1024 * 1024
			},
		},
		{
			name: "modern pull format",
			output: `[2024-01-15 10:30:45.123] Successfully pulled artifact.
[2024-01-15 10:30:45.126] Pulled 5 files. Total of 2.3 GB`,
			expected: ArtifactStats{
				FileCount: 5,
				TotalSize: 2469606195, // 2.3 * 1024 * 1024 * 1024
			},
		},
		{
			name:   "single file",
			output: `[2024-01-15 10:30:45.126] Pushed 1 file. Total of 512 KB`,
			expected: ArtifactStats{
				FileCount: 1,
				TotalSize: 524288, // 512 * 1024
			},
		},
		{
			name:   "bytes only",
			output: `[2024-01-15 10:30:45.126] Pushed 2 files. Total of 1024 B`,
			expected: ArtifactStats{
				FileCount: 2,
				TotalSize: 1024,
			},
		},
		{
			name:     "old format or no stats",
			output:   `Successfully pushed artifact`,
			expected: ArtifactStats{},
		},
		{
			name:     "empty output",
			output:   "",
			expected: ArtifactStats{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseArtifactStats(tt.output)
			if result.FileCount != tt.expected.FileCount {
				t.Errorf("FileCount: got %d, want %d", result.FileCount, tt.expected.FileCount)
			}
			// Allow small difference due to float precision
			sizeDiff := result.TotalSize - tt.expected.TotalSize
			if sizeDiff < 0 {
				sizeDiff = -sizeDiff
			}
			if sizeDiff > 10 {
				t.Errorf("TotalSize: got %d, want %d", result.TotalSize, tt.expected.TotalSize)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1572864, "1.5 MB"},
		{1073741824, "1.0 GB"},
		{10737418240, "10.0 GB"},
	}

	for _, tt := range tests {
		result := FormatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatBytes(%d): got %s, want %s", tt.bytes, result, tt.expected)
		}
	}
}
