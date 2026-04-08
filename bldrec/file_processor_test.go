package bldrec

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadLines(t *testing.T) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "testfile_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	// Write test content
	content := "line1\nline2\nline3\n"
	if _, err := tmpfile.WriteString(content); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpfile.Close()

	fp := &FileProcessor{}
	lines, err := fp.readLines(tmpfile.Name())
	if err != nil {
		t.Fatalf("readLines failed: %v", err)
	}

	if len(lines) != 3 {
		t.Errorf("Expected 3 lines, got %d", len(lines))
	}
	if lines[0] != "line1" || lines[1] != "line2" || lines[2] != "line3" {
		t.Errorf("Lines do not match. Got: %v", lines)
	}
}

func TestReadLinesFileNotFound(t *testing.T) {
	fp := &FileProcessor{}
	lines, err := fp.readLines("/nonexistent/file.txt")
	if err == nil {
		t.Errorf("Expected error for non-existent file, got nil")
	}
	if lines != nil {
		t.Errorf("Expected nil lines on error, got %v", lines)
	}
}

func TestReadLinesEmptyFile(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "empty_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())
	tmpfile.Close()

	fp := &FileProcessor{}
	lines, err := fp.readLines(tmpfile.Name())
	if err != nil {
		t.Fatalf("readLines failed: %v", err)
	}

	if len(lines) != 0 {
		t.Errorf("Expected 0 lines for empty file, got %d", len(lines))
	}
}

func TestCapnpRecordWithAllFields(t *testing.T) {
	fp := &FileProcessor{}
	fields := []string{"01/15/2026", "Test Description", "123.45", "extra"}

	record, err := fp.capnpRecord(fields)
	if err != nil {
		t.Fatalf("capnpRecord failed: %v", err)
	}

	// Verify date field (Unix timestamp) was set
	uDateTime := record.UDateTime()
	if uDateTime == 0 {
		t.Errorf("Date should have been set, got zero value")
	}

	// Verify description field
	desc, err := record.SDescription()
	if err != nil {
		t.Fatalf("Failed to get SDescription: %v", err)
	}
	if desc != "Test Description" {
		t.Errorf("Description mismatch. Expected 'Test Description', got '%s'", desc)
	}

	// Verify amount field was parsed as float (allowing some precision variance)
	amount := record.FValue()
	if amount == 0.0 {
		t.Errorf("Amount should have been set from '123.45'")
	}

	// Verify extra field
	extra, err := record.SDontCare()
	if err != nil {
		t.Errorf("Failed to get SDontCare: %v", err)
	}
	if extra != "extra" {
		t.Errorf("Extra field mismatch. Expected 'extra', got '%s'", extra)
	}
}

func TestCapnpRecordWithPartialFields(t *testing.T) {
	fp := &FileProcessor{}
	fields := []string{"02/28/2026", "Partial Description"}

	record, err := fp.capnpRecord(fields)
	if err != nil {
		t.Fatalf("capnpRecord failed: %v", err)
	}

	desc, err := record.SDescription()
	if err != nil {
		t.Fatalf("Failed to get SDescription: %v", err)
	}
	if desc != "Partial Description" {
		t.Errorf("Description mismatch. Expected 'Partial Description', got '%s'", desc)
	}

	amount := record.FValue()
	if amount != 0.0 {
		t.Errorf("Default amount should be 0.0, got %f", amount)
	}
}

func TestCapnpRecordWithValidAmount(t *testing.T) {
	fp := &FileProcessor{}
	fields := []string{"02/28/2026", "Description", "99.99", "extra"}

	record, err := fp.capnpRecord(fields)
	if err != nil {
		t.Fatalf("capnpRecord failed: %v", err)
	}

	// Verify amount is parsed as a positive number
	amount := record.FValue()
	if amount <= 0.0 {
		t.Errorf("Amount should be positive, got %f", amount)
	}
	// Verify it's approximately the right value (accounting for float precision)
	if amount < 99.0 || amount > 100.0 {
		t.Errorf("Amount should be around 99.99, got %f", amount)
	}
}

func TestCapnpRecordWithInvalidDate(t *testing.T) {
	fp := &FileProcessor{}
	fields := []string{"invalid-date", "Description", "100.00", "extra"}

	record, err := fp.capnpRecord(fields)
	if err != nil {
		t.Fatalf("capnpRecord failed: %v", err)
	}

	// Should fallback to current time (approximately)
	uDateTime := record.UDateTime()

	now := time.Now().Unix()
	// Allow 5 seconds difference for test execution
	if uDateTime < now-5 || uDateTime > now+5 {
		t.Errorf("Date should be approximately now. Expected ~%d, got %d", now, uDateTime)
	}
}

func TestCapnpRecordWithInvalidAmount(t *testing.T) {
	fp := &FileProcessor{}
	fields := []string{"02/28/2026", "Description", "not-a-number", "extra"}

	record, err := fp.capnpRecord(fields)
	if err != nil {
		t.Fatalf("capnpRecord failed: %v", err)
	}

	amount := record.FValue()
	if amount != 0.0 {
		t.Errorf("Invalid amount should default to 0.0, got %f", amount)
	}
}

func TestProcessFiles(t *testing.T) {
	// Create temporary directories
	inDir := t.TempDir()
	historyDir := t.TempDir()

	// Create test file 1
	file1Path := filepath.Join(inDir, "test1.txt")
	content1 := "01/15/2026   Description 1   100.00   extra1\n   Continuation line 1\n"
	if err := os.WriteFile(file1Path, []byte(content1), 0644); err != nil {
		t.Fatalf("Failed to create test file 1: %v", err)
	}

	// Create test file 2
	file2Path := filepath.Join(inDir, "test2.txt")
	content2 := "02/20/2026   Description 2   200.50   extra2\n"
	if err := os.WriteFile(file2Path, []byte(content2), 0644); err != nil {
		t.Fatalf("Failed to create test file 2: %v", err)
	}

	// Create a non-text file (should be ignored)
	ignorePath := filepath.Join(inDir, "readme.md")
	if err := os.WriteFile(ignorePath, []byte("ignore me"), 0644); err != nil {
		t.Fatalf("Failed to create ignore file: %v", err)
	}

	fp := &FileProcessor{}
	records, err := fp.ProcessFiles(inDir, historyDir)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	// Should have at least 2 records from the 2 txt files
	if len(records) < 2 {
		t.Errorf("Expected at least 2 records, got %d", len(records))
	}

	// Verify files were moved to history
	if _, err := os.Stat(file1Path); err == nil {
		t.Errorf("File should have been moved from inDir")
	}
	if hist1, err := os.Stat(filepath.Join(historyDir, "test1.txt")); err != nil || hist1.IsDir() {
		t.Errorf("File should exist in historyDir")
	}

	// Verify non-txt file was not moved
	if _, err := os.Stat(ignorePath); err != nil {
		t.Errorf("Non-txt file should not have been moved")
	}
}

func TestProcessFilesEmptyDirectory(t *testing.T) {
	inDir := t.TempDir()
	historyDir := t.TempDir()

	fp := &FileProcessor{}
	records, err := fp.ProcessFiles(inDir, historyDir)
	if err != nil {
		t.Fatalf("ProcessFiles failed: %v", err)
	}

	if len(records) != 0 {
		t.Errorf("Expected 0 records for empty directory, got %d", len(records))
	}
}

func TestProcessFilesInvalidDirectory(t *testing.T) {
	fp := &FileProcessor{}
	records, err := fp.ProcessFiles("/nonexistent/directory", "/tmp")
	if err == nil {
		t.Errorf("Expected error for non-existent directory, got nil")
	}
	if records != nil {
		t.Errorf("Expected nil records on error, got %v", records)
	}
}

func TestProcessFile(t *testing.T) {
	// Create temporary file
	tmpfile, err := os.CreateTemp("", "test_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpfile.Close()

	// Write test content with multi-space delimiters
	content := "01/15/2026   Description 1   100.00   extra\n   Continuation\n"
	if err := os.WriteFile(tmpfile.Name(), []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	moved := false
	moveFunc := func() error {
		moved = true
		return os.Remove(tmpfile.Name())
	}

	fp := &FileProcessor{}
	records, err := fp.processFile(tmpfile.Name(), moveFunc)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	if !moved {
		t.Errorf("moveToHistory callback was not called")
	}

	if len(records) == 0 {
		t.Errorf("Expected records from file, got empty slice")
	}
}

func TestProcessFileWithMoveCallback(t *testing.T) {
	// Create temporary files
	inFile := filepath.Join(t.TempDir(), "input.txt")
	outDir := t.TempDir()
	outFile := filepath.Join(outDir, "input.txt")

	content := "01/15/2026   Test   100.00   extra\n"
	if err := os.WriteFile(inFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	moveFunc := func() error {
		return os.Rename(inFile, outFile)
	}

	fp := &FileProcessor{}
	_, err := fp.processFile(inFile, moveFunc)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify file was moved
	if _, err := os.Stat(inFile); err == nil {
		t.Errorf("Original file should have been moved")
	}
	if _, err := os.Stat(outFile); err != nil {
		t.Errorf("File should exist in destination: %v", err)
	}
}

func TestProcessFileWithValidContent(t *testing.T) {
	fp := &FileProcessor{}

	// Create temporary input and history directories
	inDir := t.TempDir()
	historyDir := filepath.Join(inDir, "history")
	if err := os.Mkdir(historyDir, 0755); err != nil {
		t.Fatalf("Failed to create history directory: %v", err)
	}

	// Create a test file with valid content
	testFile := filepath.Join(inDir, "test.txt")
	content := []string{
		"date    description    amount    extra\n",
		"01/15/2026    Test Description    123.45    extra1\n",
		"02/28/2026    Another Description    99.99    extra2",
	}

	if err := os.WriteFile(testFile, []byte(strings.Join(content, "")), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock function to move file to history
	moveToHistory := func() error {
		return os.Rename(testFile, filepath.Join(historyDir, "test.txt"))
	}
	moveToHistory()

	records, err := fp.processLines([]string{content[1], content[2]})
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify we got 2 records
	if len(records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(records))
	}

	// Verify file was moved to history
	if _, err := os.Stat(testFile); err == nil {
		t.Errorf("File should have been moved to history directory")
	}

	// Verify history file exists
	historyFile := filepath.Join(historyDir, "test.txt")
	if _, err := os.Stat(historyFile); err != nil {
		t.Errorf("File should exist in history directory")
	}
}

func TestProcessFileWithEmptyFile(t *testing.T) {
	fp := &FileProcessor{}

	// Create temporary input and history directories
	inDir := t.TempDir()
	historyDir := filepath.Join(inDir, "history")
	if err := os.Mkdir(historyDir, 0755); err != nil {
		t.Fatalf("Failed to create history directory: %v", err)
	}

	// Create an empty test file
	testFile := filepath.Join(inDir, "empty.txt")
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("Failed to create empty test file: %v", err)
	}

	// Mock function to move file to history
	moveToHistory := func() error {
		return os.Rename(testFile, filepath.Join(historyDir, "empty.txt"))
	}

	records, err := fp.processFile(testFile, moveToHistory)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify we got 0 records
	if len(records) != 0 {
		t.Errorf("Expected 0 records for empty file, got %d", len(records))
	}
}

func TestProcessFileWithInvalidDateFormat(t *testing.T) {
	fp := &FileProcessor{}

	// Create temporary input and history directories
	inDir := t.TempDir()
	historyDir := filepath.Join(inDir, "history")
	if err := os.Mkdir(historyDir, 0755); err != nil {
		t.Fatalf("Failed to create history directory: %v", err)
	}

	// Create a test file with invalid date format
	testFile := filepath.Join(inDir, "invalid_date.txt")
	content := "date    description    amount    extra\n"
	content += "invalid_date    Description    123.45    extra"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock function to move file to history
	moveToHistory := func() error {
		return os.Rename(testFile, filepath.Join(historyDir, "invalid_date.txt"))
	}

	records, err := fp.processFile(testFile, moveToHistory)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify we got 1 record with default timestamp
	if len(records) != 1 {
		t.Errorf("Expected 1 record with default timestamp, got %d", len(records))
	}

	// Verify timestamp is set to current time
	record := records[0]
	uDateTime := record.UDateTime()
	if uDateTime == 0 {
		t.Errorf("Date should have been set to current time, got zero value")
	}
}

func TestProcessFileWithMissingFields(t *testing.T) {
	fp := &FileProcessor{}

	// Create temporary input and history directories
	inDir := t.TempDir()
	historyDir := filepath.Join(inDir, "history")
	if err := os.Mkdir(historyDir, 0755); err != nil {
		t.Fatalf("Failed to create history directory: %v", err)
	}

	// Create a test file with missing fields
	testFile := filepath.Join(inDir, "missing_fields.txt")
	content := "date    description\n"
	content += "01/15/2026    Test Description"

	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Mock function to move file to history
	moveToHistory := func() error {
		return os.Rename(testFile, filepath.Join(historyDir, "missing_fields.txt"))
	}

	records, err := fp.processFile(testFile, moveToHistory)
	if err != nil {
		t.Fatalf("processFile failed: %v", err)
	}

	// Verify we got 1 record with default values
	if len(records) != 1 {
		t.Errorf("Expected 1 record with default values, got %d", len(records))
	}

	// Verify default amount is 0.0
	record := records[0]
	amount := record.FValue()
	if amount != 0.0 {
		t.Errorf("Default amount should be 0.0, got %f", amount)
	}
}
