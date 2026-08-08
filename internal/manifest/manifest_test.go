package manifest

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestParseManifestWithContent(t *testing.T) {
	tests := []struct {
		name                  string
		filename              string
		expectedWorkflowCount int
		wantErr               error
	}{
		{
			name:                  "manifest does not exist",
			filename:              "manifest_does_not_exist.txt",
			expectedWorkflowCount: 0,
			wantErr:               ErrOpenManifest,
		},
		{
			name:                  "manifest with one workflow",
			filename:              "manifest_one_file.txt",
			expectedWorkflowCount: 1,
			wantErr:               nil,
		},
		{
			name:                  "manifest with two workflows and comments",
			filename:              "manifest_two_files_and_comments.txt",
			expectedWorkflowCount: 2,
			wantErr:               nil,
		},
		{
			name:                  "manifest with only comments and blank lines",
			filename:              "manifest_comments_and_blank_lines_only.txt",
			expectedWorkflowCount: 0,
			wantErr:               ErrEmptyManifest,
		},
		{
			name:                  "manifest empty",
			filename:              "manifest_empty.txt",
			expectedWorkflowCount: 0,
			wantErr:               ErrEmptyManifest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fullPath := filepath.Join("testdata", tt.filename)

			wfs, err := ParseManifest(fullPath)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got error: %s %v", tt.wantErr, fullPath, err)
			}

			if len(wfs) != tt.expectedWorkflowCount {
				t.Fatalf("expected %v workflow(s), got %v", tt.expectedWorkflowCount, len(wfs))
			}
		})
	}
}
