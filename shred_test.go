package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMain(m *testing.M) {
	os.MkdirAll("wtests", 0755)

	code := m.Run()

	os.RemoveAll("wtests")

	os.Exit(code)
}

func testPath(name string) string {
	return filepath.Join("wtests", name)
}

func createLargeFile(path string, size int64, content string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	contentBytes := []byte(content)
	interval := int64(1024 * 1024) // Write content every 1MB

	for offset := int64(0); offset < size; offset += interval {
		_, err := file.Seek(offset, 0)
		if err != nil {
			return err
		}
		_, err = file.Write(contentBytes)
		if err != nil {
			return err
		}
	}

	return file.Truncate(size)
}

func TestShredRemove(t *testing.T) {
	testFile := testPath("sample_test_file.txt")
	content := "This is test content that should be shredded"

	err := os.WriteFile(testFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = Shred(testFile)
	if err != nil {
		t.Fatalf("Shred failed: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("File still exists after shredding")
	}
}

func TestShredNonExistentFile(t *testing.T) {
	err := Shred("non_existent_file.txt")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestShredEmptyFile(t *testing.T) {
	testFile := testPath("empty_test.txt")

	err := os.WriteFile(testFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}

	err = Shred(testFile)
	if err != nil {
		t.Fatalf("Shred failed on empty file: %v", err)
	}

	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("Empty file still exists after shredding")
	}
}

// one of these days this test will fail
// https://xkcd.com/221/
func TestShredOverwritesContent(t *testing.T) {
	testFile := testPath("overwrite_test.txt")
	originalContent := "SENSITIVE_DATA_TO_BE_OVERWRITTEN_1234567890"

	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read original content: %v", err)
	}
	if string(content) != originalContent {
		t.Fatalf("Original content not written correctly")
	}

	// Overwrite but don't truncate or delete to verify content change
	err = shredFileWithOptions(testFile, false, false)
	if err != nil {
		t.Fatalf("Test shred failed: %v", err)
	}

	overwrittenContent, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read overwritten content: %v", err)
	}

	// verify the original data is gone
	if string(overwrittenContent) == originalContent {
		t.Errorf("Original content still present after shredding")
	}

	os.Remove(testFile)
}

//
func TestShredReadOnlyFile(t *testing.T) {
	testFile := testPath("readonly_test.txt")
	content := "read-only content"

	err := os.WriteFile(testFile, []byte(content), 0444)
	if err != nil {
		t.Fatalf("Failed to create read-only file: %v", err)
	}

	err = Shred(testFile)
	if err == nil {
		t.Error("Expected error when shredding read-only file")
	}

	os.Chmod(testFile, 0644)
	os.Remove(testFile)
}

func TestShredDirectory(t *testing.T) {
	testDir := testPath("sample_test_dir")

	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err = Shred(testDir)
	if err == nil {
		t.Error("Expected error when shredding directory")
	}

	os.Remove(testDir)
}

func TestShredSymlink(t *testing.T) {
	symlinkname := "test_symlink.txt"
	targetname := "test_symlink_target.txt"

	targetFile := testPath(targetname)
	symlinkFile := testPath(symlinkname)

	err := os.WriteFile(targetFile, []byte("target content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create target file: %v", err)
	}
	defer os.Remove(targetFile)

	err = os.Symlink(targetname, symlinkFile)
	if err != nil {
		t.Fatalf("Failed to create symlink: %v", err)
	}
	defer os.Remove(symlinkFile)

	// Check if symlink exists before calling Shred
	if _, err := os.Lstat(symlinkFile); err != nil {
		t.Fatalf("Symlink doesn't exist: %v", err)
	}
	fmt.Println("Symlink exists, calling Shred...")

	err = Shred(symlinkFile)
	if err != nil {
		t.Fatalf("Shred failed: %v", err)
	}

	// Symlink should be removed
	if _, err := os.Lstat(symlinkFile); !os.IsNotExist(err) {
		t.Error("Symlink should be removed after shredding")
	}

	// Target file should still exist but be truncated (0 bytes)
	stat, err := os.Stat(targetFile)
	if err != nil {
		t.Error("Target file should still exist after shredding symlink")
	} else if stat.Size() != 0 {
		t.Errorf("Target file should be truncated to 0 bytes, got %d bytes", stat.Size())
	}
}

func TestShredDeviceFiles(t *testing.T) {
	deviceFiles := []string{
		"/dev/null",
		"/dev/zero",
		"/dev/urandom",
	}

	for _, device := range deviceFiles {
		t.Run(device, func(t *testing.T) {
			if _, err := os.Stat(device); os.IsNotExist(err) {
				t.Skipf("Device %s not available", device)
			}

			err := Shred(device)
			if err == nil {
				t.Errorf("Expected error when shredding device file %s", device)
			}
		})
	}
}

func TestShredBlockDevice(t *testing.T) {
	t.Skip("Block device test not implemented - S_ISBLK")
}

func TestShredCharacterDevice(t *testing.T) {
	t.Skip("Character device test not implemented - S_ISCHR")
}

func TestShredSocket(t *testing.T) {
	t.Skip("Socket test not implemented - S_ISSOCK")
}

func TestShredFIFO(t *testing.T) {
	t.Skip("FIFO test not implemented - S_ISFIFO (covered by TestShredNamedPipe)")
}

func TestShredNativeObject(t *testing.T) {
	t.Skip("Native object test not implemented - S_ISNATIVE")
}

func TestShredLargeFileOverwrite(t *testing.T) {
	testFile := testPath("large_overwrite_test.txt")
	pattern := "I_AM_A_PATTERN_1234567890_ABCDABCD"
	size := int64(10 * 1024 * 1024) // 10MB

	err := createLargeFile(testFile, size, pattern)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	// Test overwrite
	err = shredFileWithOptions(testFile, false, false)
	if err != nil {
		t.Fatalf("Failed to overwrite large file: %v", err)
	}

	overwrittenData, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("Failed to read overwritten large file: %v", err)
	}

	// Verify that original pattern is destroyed
	if bytes.Contains(overwrittenData, []byte(pattern)) {
		t.Errorf("Original pattern still found after overwriting")
	}

	os.Remove(testFile)
}

func TestShredLargeFileTruncate(t *testing.T) {
	testFile := testPath("large_truncate_test.txt")
	size := int64(5 * 1024 * 1024) // 5MB
	content := "LARGE_FILE_CONTENT_TO_BE_TRUNCATED_AFTER_SHRED"

	err := createLargeFile(testFile, size, content)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	// Verify original file size
	stat, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat large file: %v", err)
	}
	if stat.Size() != size {
		t.Fatalf("File size mismatch: expected %d, got %d", size, stat.Size())
	}

	// Test truncate
	err = shredFile(testFile, false)
	if err != nil {
		t.Fatalf("Failed to truncate large file: %v", err)
	}

	// Verify file is truncated to 0 bytes
	stat, err = os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat truncated file: %v", err)
	}
	if stat.Size() != 0 {
		t.Errorf("File not truncated properly: expected 0 bytes, got %d", stat.Size())
	}

	os.Remove(testFile)
}

func TestShredLargeFile(t *testing.T) {
	testFile := testPath("large_fullshred_test.txt")
	size := int64(8 * 1024 * 1024) // 8MB
	content := "CONTENT_TO_BE_DELETED_AFTER_SHRED"

	err := createLargeFile(testFile, size, content)
	if err != nil {
		t.Fatalf("Failed to create large test file: %v", err)
	}

	// Verify file exists and has correct size
	stat, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat large file: %v", err)
	}
	if stat.Size() != size {
		t.Fatalf("File size mismatch: expected %d, got %d", size, stat.Size())
	}

	err = Shred(testFile)
	if err != nil {
		t.Fatalf("Failed to fully shred large file: %v", err)
	}

	// Verify file is deleted
	_, err = os.Stat(testFile)
	if !os.IsNotExist(err) {
		t.Errorf("File still exists after full shred")
	}
}
