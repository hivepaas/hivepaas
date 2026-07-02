package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLookup(t *testing.T) {
	tempDir1 := t.TempDir()
	tempDir2 := t.TempDir()

	file1 := filepath.Join(tempDir1, "file1.txt")
	file2 := filepath.Join(tempDir2, "file2.txt")

	_ = os.WriteFile(file1, []byte("test"), 0644)
	_ = os.WriteFile(file2, []byte("test"), 0644)

	tests := []struct {
		name       string
		filename   string
		lookupDirs []string
		wantFile   string
		wantDir    string
	}{
		{
			name:       "file in first dir",
			filename:   "file1.txt",
			lookupDirs: []string{tempDir1, tempDir2},
			wantFile:   filepath.Join(tempDir1, "file1.txt"),
			wantDir:    tempDir1,
		},
		{
			name:       "file in second dir",
			filename:   "file2.txt",
			lookupDirs: []string{tempDir1, tempDir2},
			wantFile:   filepath.Join(tempDir2, "file2.txt"),
			wantDir:    tempDir2,
		},
		{
			name:       "file not found",
			filename:   "file3.txt",
			lookupDirs: []string{tempDir1, tempDir2},
			wantFile:   "",
			wantDir:    "",
		},
		{
			name:       "empty dirs",
			filename:   "file1.txt",
			lookupDirs: []string{},
			wantFile:   "",
			wantDir:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotFile, gotDir := Lookup(tt.filename, tt.lookupDirs); gotFile != tt.wantFile || gotDir != tt.wantDir {
				t.Errorf("Lookup() = %v, %v, want %v, %v", gotFile, gotDir, tt.wantFile, tt.wantDir)
			}
		})
	}
}
