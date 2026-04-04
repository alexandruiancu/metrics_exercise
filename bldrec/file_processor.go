package bldrec

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	capnp "capnproto.org/go/capnp/v3"
)

// FileProcessor encapsulates file processing operations.
type FileProcessor struct{}

func (fp *FileProcessor) readLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func (fp *FileProcessor) capnpRecord(fields []string) (Record, error) {
	// Create a new message arena
	arena := capnp.SingleSegment(nil)
	_, seg, err := capnp.NewMessage(arena)
	if err != nil {
		panic(err)
	}
	// Create a new Record
	rec, err := NewRootRecord(seg)
	if err != nil {
		return rec, err
	}

	// Set the fields from the string array
	if len(fields) > 0 {
		const layout = "02/01/2006"
		// Parse fields[0] as a date and convert to Unix timestamp (Int64)
		t, err := time.Parse(layout, fields[0])
		if err != nil {
			// Try alternative formats or fall back to current time
			t = time.Now()
		}
		rec.SetUDateTime(t.Unix())
	}
	if len(fields) > 1 {
		rec.SetSDescription(fields[1])
	}
	if len(fields) > 2 {
		amount := strings.ReplaceAll(fields[2], ",", ".")
		f, err := strconv.ParseFloat(amount, 32)
		if err != nil {
			f = 0.0
		}
		rec.SetFValue(float32(f))
	}
	if len(fields) > 3 {
		rec.SetSDontCare(fields[3])
	}

	return rec, nil
}

func (fp *FileProcessor) ProcessFiles(inDir, historyDir string) ([]Record, error) {
	var records []Record

	// a) Open each text file in in_dir subdirectory
	files, err := os.ReadDir(inDir)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".txt") {
			continue
		}

		filePath := filepath.Join(inDir, file.Name())

		renameCurrentFile := func() error {
			return os.Rename(filePath, filepath.Join(historyDir, file.Name()))
		}
		// b) Read current file line by line and add to lines slice
		recs, err := fp.processFile(filePath, renameCurrentFile)
		if err != nil {
			return records, err
		}
		records = append(records, recs...)
	}

	return records, nil
}

func (fp *FileProcessor) processFile(filePath string, moveToHistory func() error) ([]Record, error) {

	lines, err := fp.readLines(filePath)
	if err != nil {
		// Handle error and return empty slice if there is an error reading the file
		return []Record{}, err
	}
	if err := moveToHistory(); err != nil {
		return []Record{}, err
	}

	return fp.processLines(lines)
}

func (fp *FileProcessor) processLines(lines []string) ([]Record, error) {
	var records []Record
	var aggregates [][]string

	for _, line := range lines {
		aggregate := regexp.MustCompile(`[ \t]{3,}`).Split(line, -1)
		if len(aggregate) > 0 {
			aggregates = append(aggregates, aggregate)
		}
	}

	var record []string
	for _, aggregate := range aggregates {
		if len(aggregate[0]) > 0 {
			if len(record) > 0 {
				rec, err := fp.capnpRecord(record)
				if err == nil {
					records = append(records, rec)
				}
				record = nil
			}
			record = aggregate
		} else {
			for i, field := range aggregate {
				if len(field) == 0 {
					continue
				}
				record[i] += " " + field
			}
		}
	}
	if record != nil {
		rec, err := fp.capnpRecord(record)
		if err == nil {
			records = append(records, rec)
		}
	}
	return records, nil
}
