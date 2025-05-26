/*
Copyright © 2022 Aurelio Calegari, et al.

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package reader

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestMultiFileReader(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := ioutil.TempDir("", "multi_file_reader_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create two test files
	file1Path := filepath.Join(tempDir, "file1.log")
	file2Path := filepath.Join(tempDir, "file2.log")

	// Write initial content to files
	if err := ioutil.WriteFile(file1Path, []byte("file1 line1\n"), 0644); err != nil {
		t.Fatalf("Failed to write to file1: %v", err)
	}
	if err := ioutil.WriteFile(file2Path, []byte("file2 line1\n"), 0644); err != nil {
		t.Fatalf("Failed to write to file2: %v", err)
	}

	// Create a channel to receive the streamed lines
	strChan := make(chan string, 10)

	// Create a multi-file reader
	reader := MakeMultiFileReader([]string{file1Path, file2Path}, strChan)

	// Start streaming
	if err := reader.StreamInto(); err != nil {
		t.Fatalf("Failed to start streaming: %v", err)
	}

	// Collect lines from both files
	var receivedLines []string
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		// Read initial lines
		timeout := time.After(2 * time.Second)
		for i := 0; i < 2; i++ {
			select {
			case line := <-reader.ChanReader():
				receivedLines = append(receivedLines, line)
			case <-timeout:
				t.Errorf("Timeout waiting for initial lines")
				return
			}
		}

		// Append to both files
		f1, err := os.OpenFile(file1Path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Errorf("Failed to open file1 for append: %v", err)
			return
		}
		defer f1.Close()
		if _, err := f1.WriteString("file1 line2\n"); err != nil {
			t.Errorf("Failed to append to file1: %v", err)
			return
		}

		f2, err := os.OpenFile(file2Path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			t.Errorf("Failed to open file2 for append: %v", err)
			return
		}
		defer f2.Close()
		if _, err := f2.WriteString("file2 line2\n"); err != nil {
			t.Errorf("Failed to append to file2: %v", err)
			return
		}

		// Read appended lines
		for i := 0; i < 2; i++ {
			select {
			case line := <-reader.ChanReader():
				receivedLines = append(receivedLines, line)
			case <-timeout:
				t.Errorf("Timeout waiting for appended lines")
				return
			}
		}
	}()

	// Wait for all lines to be read
	wg.Wait()

	// Close the reader
	reader.Close()

	// Verify we got all 4 lines (2 initial + 2 appended)
	if len(receivedLines) != 4 {
		t.Errorf("Expected 4 lines, got %d: %v", len(receivedLines), receivedLines)
	}

	// Check that we have lines from both files
	file1Count := 0
	file2Count := 0
	for _, line := range receivedLines {
		if line == "file1 line1" || line == "file1 line2" {
			file1Count++
		} else if line == "file2 line1" || line == "file2 line2" {
			file2Count++
		}
	}

	if file1Count != 2 {
		t.Errorf("Expected 2 lines from file1, got %d", file1Count)
	}
	if file2Count != 2 {
		t.Errorf("Expected 2 lines from file2, got %d", file2Count)
	}
}
