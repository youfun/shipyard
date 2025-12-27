package deploy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCalculateMD5(t *testing.T) {
	// Create a temporary file for testing
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")

	// Write known content to calculate expected MD5
	content := []byte("hello world")
	if err := os.WriteFile(testFile, content, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Expected MD5 of "hello world"
	expectedMD5 := "5eb63bbbe01eeed093cb22bb8f5acdc3"

	got, err := calculateMD5(testFile)
	if err != nil {
		t.Errorf("calculateMD5() error = %v", err)
		return
	}
	if got != expectedMD5 {
		t.Errorf("calculateMD5() = %v, want %v", got, expectedMD5)
	}
}

func TestCalculateMD5_FileNotFound(t *testing.T) {
	_, err := calculateMD5("/nonexistent/file.txt")
	if err == nil {
		t.Error("calculateMD5() expected error for non-existent file, got nil")
	}
}

func TestCalculateMD5_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "empty.txt")

	if err := os.WriteFile(testFile, []byte{}, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Expected MD5 of empty file
	expectedMD5 := "d41d8cd98f00b204e9800998ecf8427e"

	got, err := calculateMD5(testFile)
	if err != nil {
		t.Errorf("calculateMD5() error = %v", err)
		return
	}
	if got != expectedMD5 {
		t.Errorf("calculateMD5() = %v, want %v", got, expectedMD5)
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	// Create source file
	content := []byte("test content for copy")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy file
	if err := copyFile(srcFile, dstFile); err != nil {
		t.Errorf("copyFile() error = %v", err)
		return
	}

	// Verify destination file exists and has same content
	gotContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
		return
	}
	if string(gotContent) != string(content) {
		t.Errorf("copyFile() content = %v, want %v", string(gotContent), string(content))
	}

	// Verify source file still exists
	if _, err := os.Stat(srcFile); os.IsNotExist(err) {
		t.Error("copyFile() should not delete source file")
	}
}

func TestCopyFile_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	err := copyFile("/nonexistent/source.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("copyFile() expected error for non-existent source, got nil")
	}
}

func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	srcFile := filepath.Join(tmpDir, "source.txt")
	dstFile := filepath.Join(tmpDir, "dest.txt")

	// Create source file
	content := []byte("test content for move")
	if err := os.WriteFile(srcFile, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Move file
	if err := moveFile(srcFile, dstFile); err != nil {
		t.Errorf("moveFile() error = %v", err)
		return
	}

	// Verify destination file exists and has same content
	gotContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Errorf("Failed to read destination file: %v", err)
		return
	}
	if string(gotContent) != string(content) {
		t.Errorf("moveFile() content = %v, want %v", string(gotContent), string(content))
	}

	// Verify source file no longer exists
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("moveFile() should delete source file after copy")
	}
}

func TestMoveFile_SourceNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	err := moveFile("/nonexistent/source.txt", filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("moveFile() expected error for non-existent source, got nil")
	}
}
