/*
Copyright 2022 Aurelio Calegari, et al.

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

package loggo

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/marawanxmamdouh/loggo/filter"

	"github.com/marawanxmamdouh/loggo/config"
	"github.com/rivo/tview"
)

var bytePool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 1024) // Initial capacity 1KB
	},
}

// LogEntry represents a structured log entry parsed from plain text
type LogEntry struct {
	Timestamp string      `json:"timestamp"`
	Level     string      `json:"level"`
	File      string      `json:"file"`
	Line      int         `json:"line"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	ParseErr  string      `json:"parse_err,omitempty"`
}

// ToMap converts a LogEntry to a map[string]interface{} for compatibility with existing code
func (le *LogEntry) ToMap() map[string]interface{} {
	m := make(map[string]interface{})

	m["timestamp"] = le.Timestamp
	m["level"] = le.Level
	m["file"] = le.File
	m["line"] = le.Line
	m["message"] = le.Message

	// Add Data fields directly to the map
	if le.Data != nil {
		m["data"] = le.Data
	}

	// Add parse error if exists
	if le.ParseErr != "" {
		m[config.ParseErr] = le.ParseErr
	}

	return m
}

// parseTextLogLine parses a text log line with the format:
// timestamp ; level ; file ; line ; message ;;; data
func parseTextLogLine(line []byte) (map[string]interface{}, error) {
	logEntry := &LogEntry{}
	lineStr := string(line)

	// Check if the log line is in JSON format first
	var jsonMap map[string]interface{}
	if err := json.Unmarshal(line, &jsonMap); err == nil {
		// This is a valid JSON log, return it directly
		return jsonMap, nil
	}

	// Try to parse as text log format
	// First, find the data section which starts with ";;;"
	parts := strings.SplitN(lineStr, ";;;", 2)
	mainSection := parts[0]
	var dataSection string
	if len(parts) > 1 {
		dataSection = strings.TrimSpace(parts[1])
	}

	// Parse main section (fields separated by ";")
	fields := strings.Split(mainSection, ";")
	if len(fields) < 5 {
		return map[string]interface{}{
			config.ParseErr:    "Invalid text log format: not enough fields",
			config.TextPayload: lineStr,
		}, fmt.Errorf("invalid text log format: not enough fields")
	}

	// Assign fields to logEntry
	logEntry.Timestamp = strings.TrimSpace(fields[0])
	logEntry.Level = strings.TrimSpace(fields[1])
	logEntry.File = strings.TrimSpace(fields[2])

	// Parse line number
	lineNum, err := strconv.Atoi(strings.TrimSpace(fields[3]))
	if err == nil {
		logEntry.Line = lineNum
	} else {
		logEntry.ParseErr = fmt.Sprintf("Invalid line number: %s", err.Error())
	}

	logEntry.Message = strings.TrimSpace(fields[4])

	// Parse data section as JSON if it exists
	if dataSection != "" {
		var dataValue interface{}
		if err := json.Unmarshal([]byte(dataSection), &dataValue); err == nil {
			logEntry.Data = dataValue
		} else {
			// If data section is not valid JSON, store the error and raw text
			logEntry.ParseErr = fmt.Sprintf("Invalid JSON in data section: %s", err.Error())
			// Store the raw text as is
			logEntry.Data = dataSection
		}
	}

	return logEntry.ToMap(), nil
}

// ParseTextLogLine is the exported version of parseTextLogLine for use in examples and tests
// It parses a text log line with the format: timestamp ; level ; file ; line ; message ;;; data
func ParseTextLogLine(line []byte) (map[string]interface{}, error) {
	return parseTextLogLine(line)
}

func (l *LogView) read() {
	go func() {
		if err := l.chanReader.StreamInto(); err != nil {
			l.app.ShowPrefabModal(fmt.Sprintf("Unable to start stream: %v", err), 40, 10,
				func(event *tcell.EventKey) *tcell.EventKey {
					switch event.Key() {
					case tcell.KeyEnter, tcell.KeyEsc:
						l.app.Stop()
						return nil
					}
					switch event.Rune() {
					case 'Q', 'q':
						l.app.Stop()
						return nil
					}
					return event
				},
				tview.NewButton("[darkred::bu]Q[-::-]uit").SetSelectedFunc(func() {
					l.app.Stop()
				}))
			return
		}

		if len(l.config.LastSavedName) > 0 {
			l.keyMap = l.config.KeyMap()
		}

		// Set initial following state
		l.isFollowing = true
		l.updateLineView()

		// Process logs line by line
		for data := range l.chanReader.ChanReader() {
			if len(data) == 0 {
				continue
			}

			// Get a buffer from pool and copy data
			buf := bytePool.Get().([]byte)
			buf = append(buf[:0], data...) // Reset and copy

			// Parse the log line
			m, err := parseTextLogLine(buf)
			if err != nil {
				m = make(map[string]interface{})
				m[config.ParseErr] = err.Error()
				m[config.TextPayload] = string(buf) // Only convert to string when needed
			}

			// Return buffer to pool
			bytePool.Put(buf)

			// Add to inSlice with lock
			l.filterLock.Lock()
			l.inSlice = append(l.inSlice, m)
			l.filterLock.Unlock()

			// Apply filter if needed
			select {
			case exp := <-l.filterChannel:
				l.filterLock.Lock()
				l.finSlice = l.finSlice[:0]
				l.globalCount = 0
				l.filterLock.Unlock()
				l.filterLine(exp, len(l.inSlice)-1)
			default:
				// No filter change, just process the new line
				l.filterLine(nil, len(l.inSlice)-1)
			}

			// Batch UI updates
			if l.isFollowing && len(l.inSlice)%10 == 0 { // Update every 10 lines
				l.app.app.QueueUpdate(func() {
					l.table.ScrollToEnd()
				})
			}
		}
	}()
}

func (l *LogView) processSampleForConfig(sampling []map[string]interface{}) {
	if len(l.config.LastSavedName) > 0 || l.isTemplateViewShown() {
		return
	}
	l.config, l.keyMap = config.MakeConfigFromSample(sampling, l.config.Keys...)
	l.app.config = l.config
}

func (l *LogView) filter() {
	go func() {
		for {
			l.rebufferFilter = false
			exp := <-l.filterChannel
			l.clearFilterBuffer()
			l.globalCount = 0
			l.updateLineView()
			l.app.Draw()
			for i := 0; ; {
				lastUpdate := time.Now().Add(-time.Minute)
				if l.rebufferFilter {
					break
				}
				size := len(l.inSlice)
				if i < size {
					if err := l.filterLine(exp, i); err != nil {
						break
					}
					i++
				} else {
					time.Sleep(100 * time.Millisecond)
					continue
				}
				now := time.Now()
				if now.Sub(lastUpdate)*time.Millisecond > 500 {
					lastUpdate = now
					l.app.Draw()
					if l.isFollowing {
						l.table.ScrollToEnd()
					}
				}
			}
		}
	}()
}

func (l *LogView) clearFilterBuffer() {
	l.filterLock.Lock()
	defer l.filterLock.Unlock()
	l.finSlice = l.finSlice[:0]
}

func (l *LogView) sampleAndCount() {
	if len(l.config.LastSavedName) == 0 {
		if len(l.finSlice) > 20 {
			l.processSampleForConfig(l.finSlice[len(l.finSlice)-20:])
		} else {
			l.processSampleForConfig(l.finSlice)
		}
	}
	l.updateLineView()
}

func (l *LogView) filterLine(e *filter.Expression, index int) error {
	l.filterLock.Lock()
	defer l.filterLock.Unlock()
	row := l.inSlice[index]
	if e == nil {
		l.finSlice = append(l.finSlice, row)
		l.globalCount++
		l.sampleAndCount()
		return nil
	}
	a, err := e.Apply(row, l.keyMap)
	if err != nil {
		l.app.ShowPrefabModal(fmt.Sprintf("[yellow::b]Error interpreting filter expression:[-::-]\n"+
			"Filter stream has reset. Please adjust the filter expression"+
			"\n[::i]%v", err), 50, 12,
			func(event *tcell.EventKey) *tcell.EventKey {
				switch event.Key() {
				case tcell.KeyEnter, tcell.KeyEsc:
					l.app.DismissModal(l.table)
					return nil
				}
				switch event.Rune() {
				case 'C', 'c':
					l.app.DismissModal(l.table)
					return nil
				}
				return event
			},
			tview.NewButton("[darkred::bu]C[-::-]ancel").SetSelectedFunc(func() {
				l.app.DismissModal(l.table)
			}))
		l.filterChannel <- nil
		return err
	}
	if a {
		l.finSlice = append(l.finSlice, row)
		l.globalCount++
		l.sampleAndCount()
	}
	return nil
}
