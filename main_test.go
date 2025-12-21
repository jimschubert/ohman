package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const defaultRegex = `(.+)\s\((\d+)\)\.(pdf|mobi|mp4|epub|wav|mp3)$`

// Helper function to create a temporary directory with test files
func setupTestDir(t *testing.T) string {
	t.Helper()
	return t.TempDir()
}

// Helper function to create a file with content
func createTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file %s: %v", path, err)
	}
}

// Helper function to create a file with a specific modification time
func createTestFileWithModTime(t *testing.T, path, content string, modTime time.Time) {
	t.Helper()
	createTestFile(t, path, content)
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("failed to set mod time for %s: %v", path, err)
	}
}

// Helper function to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func TestCLI_Run_NoPath(t *testing.T) {
	t.Parallel()
	cli := &CLI{
		Path:  []string{},
		Regex: defaultRegex,
	}

	err := cli.Run(nil)
	if err == nil {
		t.Fatal("expected error when no path is specified")
	}
	if !strings.Contains(err.Error(), "at least one path must be specified") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLI_Run_InvalidRegex(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	cli := &CLI{
		Path:  []string{dir},
		Regex: "[invalid",
	}

	err := cli.Run(nil)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid regex") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLI_Run_InvalidPath(t *testing.T) {
	t.Parallel()
	cli := &CLI{
		Path:  []string{"/nonexistent/path/that/does/not/exist"},
		Regex: defaultRegex,
	}

	err := cli.Run(nil)
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
	if !strings.Contains(err.Error(), "error walking path") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestCLI_Run_DryRun_FindsDuplicates(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	// Create original and duplicate files
	createTestFile(t, filepath.Join(dir, "book.pdf"), "original content")
	createTestFile(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1")
	createTestFile(t, filepath.Join(dir, "book (2).pdf"), "duplicate 2")

	cli := &CLI{
		Path:   []string{dir},
		DryRun: true,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify files still exist (dry run shouldn't delete anything)
	if !fileExists(filepath.Join(dir, "book.pdf")) {
		t.Error("original file should still exist after dry run")
	}
	if !fileExists(filepath.Join(dir, "book (1).pdf")) {
		t.Error("duplicate 1 should still exist after dry run")
	}
	if !fileExists(filepath.Join(dir, "book (2).pdf")) {
		t.Error("duplicate 2 should still exist after dry run")
	}
}

func TestCLI_Run_DryRun_NoDuplicates(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	// Create only original files (no duplicates)
	createTestFile(t, filepath.Join(dir, "book.pdf"), "original content")
	createTestFile(t, filepath.Join(dir, "another.mobi"), "another file")

	cli := &CLI{
		Path:   []string{dir},
		DryRun: true,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCLI_Run_Delete_RemovesDuplicates(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	// Create original and duplicate files
	createTestFile(t, filepath.Join(dir, "book.pdf"), "original content")
	createTestFile(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1")
	createTestFile(t, filepath.Join(dir, "book (2).pdf"), "duplicate 2")

	outFile := filepath.Join(dir, "results.txt")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Out:    outFile,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original still exists
	if !fileExists(filepath.Join(dir, "book.pdf")) {
		t.Error("original file should still exist")
	}

	// Verify duplicates are deleted
	if fileExists(filepath.Join(dir, "book (1).pdf")) {
		t.Error("duplicate 1 should be deleted")
	}
	if fileExists(filepath.Join(dir, "book (2).pdf")) {
		t.Error("duplicate 2 should be deleted")
	}

	// Verify output file was created
	if !fileExists(outFile) {
		t.Error("output file should exist")
	}
}

func TestCLI_Run_Delete_Inverse_KeepsNewest(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	now := time.Now()

	// Create original (oldest) and duplicate files with different mod times
	createTestFileWithModTime(t, filepath.Join(dir, "book.pdf"), "original", now.Add(-2*time.Hour))
	createTestFileWithModTime(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1", now.Add(-1*time.Hour))
	createTestFileWithModTime(t, filepath.Join(dir, "book (2).pdf"), "newest duplicate", now) // newest

	outFile := filepath.Join(dir, "results.txt")

	cli := &CLI{
		Path:    []string{dir},
		Delete:  true,
		Inverse: true,
		Out:     outFile,
		Regex:   defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original is deleted (inverse mode)
	if fileExists(filepath.Join(dir, "book.pdf")) {
		t.Error("original file should be deleted in inverse mode")
	}

	// Verify oldest duplicate is deleted
	if fileExists(filepath.Join(dir, "book (1).pdf")) {
		t.Error("older duplicate should be deleted")
	}

	// Verify newest duplicate is kept
	if !fileExists(filepath.Join(dir, "book (2).pdf")) {
		t.Error("newest duplicate should be kept")
	}
}

func TestCLI_Run_Delete_InverseAndRename(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	now := time.Now()

	// Create original (oldest) and duplicate files with different mod times
	createTestFileWithModTime(t, filepath.Join(dir, "book.pdf"), "original", now.Add(-2*time.Hour))
	createTestFileWithModTime(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1", now.Add(-1*time.Hour))
	createTestFileWithModTime(t, filepath.Join(dir, "book (2).pdf"), "newest content", now) // newest

	outFile := filepath.Join(dir, "results.txt")

	cli := &CLI{
		Path:             []string{dir},
		Delete:           true,
		InverseAndRename: true,
		Out:              outFile,
		Regex:            defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify duplicates are deleted
	if fileExists(filepath.Join(dir, "book (1).pdf")) {
		t.Error("duplicate 1 should be deleted")
	}
	if fileExists(filepath.Join(dir, "book (2).pdf")) {
		t.Error("duplicate 2 should be renamed and no longer exist at original path")
	}

	// Verify the newest was renamed to the original name
	if !fileExists(filepath.Join(dir, "book.pdf")) {
		t.Error("newest file should be renamed to original name")
	}

	// Verify content is from the newest file
	content, err := os.ReadFile(filepath.Join(dir, "book.pdf"))
	if err != nil {
		t.Fatalf("failed to read renamed file: %v", err)
	}
	if string(content) != "newest content" {
		t.Errorf("expected content 'newest content', got '%s'", string(content))
	}
}

func TestCLI_Run_OutputToFile(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	createTestFile(t, filepath.Join(dir, "book.pdf"), "original content")
	createTestFile(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1")

	outFile := filepath.Join(dir, "custom-output.txt")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Out:    outFile,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fileExists(outFile) {
		t.Error("custom output file should exist")
	}

	content, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(content), "Deleted") {
		t.Errorf("output file should contain deletion info, got: %s", string(content))
	}
}

func TestCLI_Run_DefaultOutputFile(t *testing.T) {
	// Do not run in parallel because it changes the process working directory
	dir := setupTestDir(t)

	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}
	// restore working directory even if test fails
	defer func() { _ = os.Chdir(originalWd) }()

	createTestFile(t, filepath.Join(dir, "book.pdf"), "original content")
	createTestFile(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Default output file should be results.txt in current directory
	if !fileExists(filepath.Join(dir, "results.txt")) {
		t.Error("default results.txt file should exist")
	}
}

func TestCLI_Run_MultiplePaths(t *testing.T) {
	t.Parallel()
	dir1 := setupTestDir(t)
	dir2 := setupTestDir(t)

	// Create files in first directory
	createTestFile(t, filepath.Join(dir1, "book1.pdf"), "original 1")
	createTestFile(t, filepath.Join(dir1, "book1 (1).pdf"), "duplicate 1")

	// Create files in second directory
	createTestFile(t, filepath.Join(dir2, "book2.mobi"), "original 2")
	createTestFile(t, filepath.Join(dir2, "book2 (1).mobi"), "duplicate 2")

	outFile := filepath.Join(dir1, "results.txt")

	cli := &CLI{
		Path:   []string{dir1, dir2},
		Delete: true,
		Out:    outFile,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify originals still exist
	if !fileExists(filepath.Join(dir1, "book1.pdf")) {
		t.Error("original 1 should still exist")
	}
	if !fileExists(filepath.Join(dir2, "book2.mobi")) {
		t.Error("original 2 should still exist")
	}

	// Verify duplicates are deleted
	if fileExists(filepath.Join(dir1, "book1 (1).pdf")) {
		t.Error("duplicate 1 should be deleted")
	}
	if fileExists(filepath.Join(dir2, "book2 (1).mobi")) {
		t.Error("duplicate 2 should be deleted")
	}
}

func TestCLI_Run_CustomRegex(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	// Create files matching a custom regex pattern (e.g., file_copy1.txt)
	createTestFile(t, filepath.Join(dir, "document.txt"), "original")
	createTestFile(t, filepath.Join(dir, "document_copy1.txt"), "copy 1")
	createTestFile(t, filepath.Join(dir, "document_copy2.txt"), "copy 2")

	outFile := filepath.Join(dir, "results.txt")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Out:    outFile,
		Regex:  `(.+)_copy(\d+)\.(txt)$`,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify original still exists
	if !fileExists(filepath.Join(dir, "document.txt")) {
		t.Error("original should still exist")
	}

	// Verify copies are deleted
	if fileExists(filepath.Join(dir, "document_copy1.txt")) {
		t.Error("copy 1 should be deleted")
	}
	if fileExists(filepath.Join(dir, "document_copy2.txt")) {
		t.Error("copy 2 should be deleted")
	}
}

func TestCLI_Run_DuplicateWithoutOriginal(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	// Create only duplicate files (no original)
	createTestFile(t, filepath.Join(dir, "book (1).pdf"), "duplicate 1")
	createTestFile(t, filepath.Join(dir, "book (2).pdf"), "duplicate 2")

	cli := &CLI{
		Path:   []string{dir},
		DryRun: true,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Files should still exist since there's no original to match against
	if !fileExists(filepath.Join(dir, "book (1).pdf")) {
		t.Error("duplicate 1 should still exist (no original found)")
	}
	if !fileExists(filepath.Join(dir, "book (2).pdf")) {
		t.Error("duplicate 2 should still exist (no original found)")
	}
}

func TestCLI_Run_NestedDirectories(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	subdir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(subdir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Create files in root directory
	createTestFile(t, filepath.Join(dir, "book.pdf"), "original")
	createTestFile(t, filepath.Join(dir, "book (1).pdf"), "duplicate")

	// Create files in subdirectory
	createTestFile(t, filepath.Join(subdir, "movie.mp4"), "original movie")
	createTestFile(t, filepath.Join(subdir, "movie (1).mp4"), "duplicate movie")

	outFile := filepath.Join(dir, "results.txt")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Out:    outFile,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify originals still exist
	if !fileExists(filepath.Join(dir, "book.pdf")) {
		t.Error("original book should still exist")
	}
	if !fileExists(filepath.Join(subdir, "movie.mp4")) {
		t.Error("original movie should still exist")
	}

	// Verify duplicates are deleted
	if fileExists(filepath.Join(dir, "book (1).pdf")) {
		t.Error("duplicate book should be deleted")
	}
	if fileExists(filepath.Join(subdir, "movie (1).mp4")) {
		t.Error("duplicate movie should be deleted")
	}
}

func TestCLI_Run_AllSupportedExtensions(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	extensions := []string{"pdf", "mobi", "mp4", "epub", "wav", "mp3"}

	for _, ext := range extensions {
		createTestFile(t, filepath.Join(dir, "file."+ext), "original")
		createTestFile(t, filepath.Join(dir, "file (1)."+ext), "duplicate")
	}

	outFile := filepath.Join(dir, "results.txt")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Out:    outFile,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, ext := range extensions {
		if !fileExists(filepath.Join(dir, "file."+ext)) {
			t.Errorf("original .%s should still exist", ext)
		}
		if fileExists(filepath.Join(dir, "file (1)."+ext)) {
			t.Errorf("duplicate .%s should be deleted", ext)
		}
	}
}

func TestCLI_Run_UnsupportedExtension(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	// Create files with unsupported extension
	createTestFile(t, filepath.Join(dir, "document.docx"), "original")
	createTestFile(t, filepath.Join(dir, "document (1).docx"), "duplicate")

	cli := &CLI{
		Path:   []string{dir},
		Delete: true,
		Regex:  defaultRegex,
	}

	if err := cli.Run(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Both files should still exist (unsupported extension)
	if !fileExists(filepath.Join(dir, "document.docx")) {
		t.Error("original .docx should still exist (unsupported extension)")
	}
	if !fileExists(filepath.Join(dir, "document (1).docx")) {
		t.Error("duplicate .docx should still exist (unsupported extension)")
	}
}

func TestOutputResults(t *testing.T) {
	t.Parallel()
	dir := setupTestDir(t)

	outFile := filepath.Join(dir, "output.txt")
	content := "Line 1\nLine 2\nLine 3"

	if err := outputResults(outFile, content); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fileExists(outFile) {
		t.Fatal("output file should exist")
	}

	readContent, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(readContent) != content {
		t.Errorf("expected content %q, got %q", content, string(readContent))
	}
}

func TestOutputResults_InvalidPath(t *testing.T) {
	t.Parallel()
	tmp := setupTestDir(t)
	invalidPath := filepath.Join(tmp, "nonexistent", "file.txt")

	err := outputResults(invalidPath, "content")
	if err == nil {
		t.Fatal("expected error for invalid path")
	}
	if !strings.Contains(err.Error(), "failed to write results") {
		t.Errorf("unexpected error message: %v", err)
	}
}
