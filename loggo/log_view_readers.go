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

package loggo

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"

	"github.com/marawanxmamdouh/loggo/filter"

	"github.com/marawanxmamdouh/loggo/config"
	"github.com/rivo/tview"
)

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
		} else {
			if len(l.config.LastSavedName) > 0 {
				l.keyMap = l.config.KeyMap()
			}
			for {
				t := <-l.chanReader.ChanReader()
				if len(t) > 0 {
					m := make(map[string]interface{})
					err := json.Unmarshal([]byte(t), &m)
					if err != nil {
						m[config.ParseErr] = err.Error()
						m[config.TextPayload] = t
					}
					l.inSlice = append(l.inSlice, m)
				}
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
