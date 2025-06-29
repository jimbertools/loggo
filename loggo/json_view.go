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
	"sort"
	"strings"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell/v2"
	"github.com/jimbertools/loggo/color"
	"github.com/jimbertools/loggo/search"
	"github.com/rivo/tview"
)

type JsonView struct {
	tview.Flex
	app                      Loggo
	textView                 *tview.TextView
	searchInput              *tview.InputField
	searchType               *tview.DropDown
	statusBar                *tview.TextView
	contextMenu              *tview.List
	jText                    []byte
	searchWord               string
	isSearching              bool
	indent                   string
	searchStrategy           search.Searchable
	withSearchTag            string
	wordWrap                 bool
	showQuit                 bool
	isCopyMode               bool
	toggleFullScreenCallback func()
	closeCallback            func()
}

func NewJsonView(app Loggo, showQuit bool,
	toggleFullScreenCallback, closeCallback func()) *JsonView {
	v := &JsonView{
		Flex:                     *tview.NewFlex(),
		app:                      app,
		indent:                   "  ",
		isCopyMode:               true,
		wordWrap:                 true,
		showQuit:                 showQuit,
		toggleFullScreenCallback: toggleFullScreenCallback,
		closeCallback:            closeCallback,
	}
	v.makeUIComponents()
	v.makeLayouts(false)
	v.makeContextMenu()

	return v
}

// SetJson sets a JSON and colourise accordingly, replacing any existing content. If it
// fails to parse the json, it displays the text as plain text.
func (j *JsonView) SetJson(jText []byte) *JsonView {
	j.jText = jText
	return j.setJson()
}

func (j *JsonView) makeUIComponents() {
	j.textView = tview.NewTextView().
		SetTextAlign(tview.AlignLeft).
		SetRegions(true).
		SetDynamicColors(true).
		SetWrap(j.wordWrap)

	j.textView.
		SetBackgroundColor(color.ColorBackgroundField).
		SetBorderPadding(0, 0, 1, 1).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if j.isSearching {
				switch event.Rune() {
				case 'n', 'N':
					j.next()
					return nil
				case 'p', 'P':
					j.prev()
					return nil
				case 'c', 'C':
					j.clearSearch()
					return nil
				}
				switch event.Key() {
				case tcell.KeyEsc:
					j.clearSearch()
					return nil
				case tcell.KeyTAB, tcell.KeyEnter:
					j.next()
					return nil
				}
			}
			switch event.Rune() {
			case '`':
				j.copyToClipboard()
			case 'f', 'F':
				if j.toggleFullScreenCallback != nil {
					j.toggleFullScreenCallback()
					return nil
				}
			case 's', 'S':
				j.prepareCaseInsensitiveSearch()
				return nil
			case 'r', 'R':
				j.prepareRegexSearch()
				return nil
			case 'x', 'X':
				if j.closeCallback != nil {
					j.closeCallback()
					return nil
				}
			case 'q':
				if j.showQuit {
					j.app.Stop()
					return nil
				}
			case 'w', 'W':
				j.wordWrap = !j.wordWrap
				j.textView.SetWrap(j.wordWrap)
				return nil
			}
			switch event.Key() {
			case tcell.KeyEsc:
				if j.closeCallback != nil {
					j.closeCallback()
					return nil
				} else if j.isSearching {
					j.clearSearch()
					return nil
				}
				return nil
			}
			return event
		})

	j.contextMenu = tview.NewList()
	j.contextMenu.
		SetBorder(true).
		SetTitle("Context Menu").
		SetBackgroundColor(color.ColorBackgroundField)

	j.searchInput = tview.NewInputField()
	j.searchInput.SetFieldStyle(color.FieldStyle).
		SetBorder(true).
		SetBackgroundColor(color.ColorBackgroundField)
	j.searchInput.SetChangedFunc(func(text string) {
		if len(text) == 0 {
			return
		}
		j.search(text)
	})
	j.searchInput.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEsc:
			j.clearSearch()
			return nil
		case tcell.KeyEnter:
			if len(j.searchInput.GetText()) > 0 {
				j.app.SetFocus(j.contextMenu)
				j.next()
				return nil
			} else {
				j.clearSearch()
				return nil
			}
		}
		return event
	})

	j.statusBar = tview.NewTextView()
	j.statusBar.SetBackgroundColor(color.ColorBackgroundField).SetBorder(true)
}

func (j *JsonView) makeLayouts(search bool) {
	mainContent := tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(j.contextMenu, 30, 1, false).
		AddItem(j.textView, 0, 2, false)

	j.Flex.Clear().SetDirection(tview.FlexRow)
	j.Flex.AddItem(mainContent, 0, 2, false)
	if search {
		j.Flex.AddItem(tview.NewFlex().
			SetDirection(tview.FlexColumn).
			AddItem(j.searchInput, 30, 1, false).
			AddItem(j.statusBar.Clear(), 0, 1, false),
			3, 1, false,
		)
	}
}

func (j *JsonView) makeContextMenu() {
	j.contextMenu.Clear().ShowSecondaryText(false).SetBorderPadding(0, 0, 1, 1)
	j.contextMenu.
		ShowSecondaryText(false)
	if j.isSearching {
		j.contextMenu.
			AddItem("Next Result", "", 'n', func() {
				j.next()
			}).
			AddItem("Previous Result", "", 'p', func() {
				j.prev()
			}).
			AddItem("Clear Search", "", 'c', func() {
				j.clearSearch()
			})
	}

	if j.toggleFullScreenCallback != nil {
		j.contextMenu.AddItem("Toggle Full Screen", "", 'f', func() {
			j.toggleFullScreenCallback()
		})
	}

	j.contextMenu.
		AddItem("Copy to Clipboard", "", '`', func() {
			j.copyToClipboard()
		}).
		AddItem("Search Word", "", 's', func() {
			j.prepareCaseInsensitiveSearch()
		}).
		AddItem("Search Regex", "", 'r', func() {
			j.prepareRegexSearch()
		}).
		AddItem("Go to Top", "", 'g', func() {
			j.textView.ScrollToBeginning()
		}).
		AddItem("Go to Bottom", "", 'G', func() {
			j.textView.ScrollToEnd()
		}).
		AddItem("Toggle word wrap", "", 'w', func() {
			j.wordWrap = !j.wordWrap
			j.textView.SetWrap(j.wordWrap)
		})

	if j.closeCallback != nil {
		j.contextMenu.AddItem("Close", "", 'x', func() {
			j.closeCallback()
		})
	}
	if j.showQuit {
		j.contextMenu.AddItem("Quit", "", 'q', func() {
			j.app.Stop()
		})
	}
}

func (j *JsonView) copyToClipboard() {
	size := ""
	l := float64(len(j.jText))
	if l > 1000000 {
		size = fmt.Sprintf(`[yellow::bu]%.2fMB[-::-]`, l/1000000.0)
	} else if l > 1000 {
		size = fmt.Sprintf(`[yellow::bu]%.2f KB[-::-]`, l/1000.0)
	} else {
		size = fmt.Sprintf(`[yellow::bu]%.0f bytes[-::-]`, l)
	}
	j.app.ShowPopMessage(fmt.Sprintf(`Copied %s to clipboard`, size), 2, j.textView)
	// Attempt formatting
	b := j.jText
	m := make(map[string]any)
	err := json.Unmarshal(b, &m)
	if err == nil {
		b2, err := json.MarshalIndent(m, "", "  ")
		if err == nil {
			b = b2
		}
	}
	// Paste to clipboard
	_ = clipboard.WriteAll(string(b))
}

func (j *JsonView) prepareCaseInsensitiveSearch() {
	if j.searchStrategy != nil {
		j.searchStrategy.Clear()
	}
	j.searchStrategy = search.MakeCaseInsensitiveSearch(j.statusBar)
	j.makeLayouts(true)
	j.searchInput.SetTitle("Search Word")
	j.app.SetFocus(j.searchInput)
	if len(j.searchInput.GetText()) > 0 {
		j.search(j.searchInput.GetText())
	}
}

func (j *JsonView) prepareRegexSearch() {
	if j.searchStrategy != nil {
		j.searchStrategy.Clear()
	}
	j.searchStrategy = search.MakeRegexSearch(j.statusBar)
	j.makeLayouts(true)
	j.searchInput.SetTitle("Search Regex")
	j.app.SetFocus(j.searchInput)
	if len(j.searchInput.GetText()) > 0 {
		j.search(j.searchInput.GetText())
	}
}

func (j *JsonView) search(word string) []int {
	j.isSearching = true
	j.isCopyMode = false
	j.makeContextMenu()
	j.searchStrategy.Clear()
	j.withSearchTag = word
	j.setJson()
	j.textView.
		Highlight(fmt.Sprintf(`%d`, j.searchStrategy.GetSearchPosition()-1)).
		ScrollToHighlight()

	j.searchStrategy.SetCurrentStatus()
	return nil
}

func (j *JsonView) next() {
	j.searchStrategy.Next()
	j.textView.
		Highlight(fmt.Sprintf(`%d`, j.searchStrategy.GetSearchPosition()-1)).
		ScrollToHighlight()

	j.searchStrategy.SetCurrentStatus()
}

func (j *JsonView) prev() {
	j.searchStrategy.Prev()
	j.textView.
		Highlight(fmt.Sprintf(`%d`, j.searchStrategy.GetSearchPosition()-1)).
		ScrollToHighlight()

	j.searchStrategy.SetCurrentStatus()
}

func (j *JsonView) clearSearch() {
	j.app.SetFocus(j.textView)
	j.searchInput.SetText("")
	j.isSearching = false
	j.isCopyMode = true
	j.withSearchTag = ""
	j.setJson()
	j.makeLayouts(false)
	j.makeContextMenu()
}

func (j *JsonView) setJson() *JsonView {
	jMap := make(map[string]interface{})
	if err := json.Unmarshal(j.jText, &jMap); err != nil {
		tex := string(j.jText)
		sb := strings.Builder{}
		wordList := strings.Split(tex, " ")
		for i, w := range wordList {
			if word := j.captureWordSection(w, j.withSearchTag); len(word) > 0 {
				sb.WriteString(word)
			} else {
				sb.WriteString(w)
			}
			if i < len(wordList)-1 {
				sb.WriteString(" ")
			}
		}
		j.wordWrap = true
		j.textView.SetWrap(j.wordWrap)
		j.textView.SetText(sb.String()).SetTextColor(tcell.ColorRed)
	} else {
		text := &strings.Builder{}
		text.WriteString("{" + j.newLine())
		kc := len(jMap)
		i := 0
		keys := j.extractKeys(jMap)
		for _, k := range keys {
			v := jMap[k]
			j.processNode(k, v, j.indent, text, i+1 == kc)
			text.WriteString(j.newLine())
			i++
		}
		text.WriteString("}")
		markedText := text.String()
		j.textView.SetText(markedText)
	}

	return j
}

func (j *JsonView) processNode(k, v interface{}, indent string, text *strings.Builder, last bool) {
	word := j.captureWordSection(k, j.withSearchTag)
	if word != "" {
		k = word
	}
	key := fmt.Sprintf(`%s%s"%v"%s: `, indent, color.ClField, k, color.ClWhite)
	text.WriteString(key)
	switch tp := v.(type) {
	case int,
		float64,
		bool:
		j.processNumeric(text, v, "")
	case string:
		j.processString(text, v, "")
	case map[string]interface{}:
		if last {
			j.processObject(text, v, j.indent+"")
		} else {
			j.processObject(text, v, j.computeIndent(indent[len(j.indent):])+"")
		}
	case []interface{}:
		j.processArray(text, tp, j.indent+indent)
	}
	if !last {
		text.WriteString(",")
	}
}

func (j *JsonView) processArray(text *strings.Builder, tp []interface{}, indent string) {
	if len(tp) == 0 {
		text.WriteString("[]")
		return
	}

	text.WriteString("[" + j.newLine())
	kc := len(tp)
	i := 0
	for _, n := range tp {
		j.processArrayItem(n, indent, text, i+1 == kc)
		text.WriteString(j.newLine())
		i++
	}
	text.WriteString(j.computeIndent(indent[len(j.indent):]) + "]")
}

func (j *JsonView) processObject(text *strings.Builder, val interface{}, indent string) {
	if len(val.(map[string]interface{})) == 0 {
		text.WriteString("{}")
		return
	}

	text.WriteString(color.ClString)
	text.WriteString(fmt.Sprintf(`[white::]{%s`, j.newLine()))

	vmap := val.(map[string]interface{})
	kc := len(vmap)
	i := 0

	keys := j.extractKeys(vmap)
	for _, k := range keys {
		v := vmap[k]
		j.processNode(k, v, indent+j.indent, text, i+1 == kc)
		text.WriteString(j.newLine())
		i++
	}
	text.WriteString(j.computeIndent(indent+j.indent) + `}`)
}

func (j *JsonView) processString(text *strings.Builder, v interface{}, indent string) {
	val := fmt.Sprintf(`%v`, v)
	//val = strings.ReplaceAll(val, "\"", "\\\"")
	//val = strings.ReplaceAll(val, "\n", "\\n")
	if word := j.captureWordSection(v, j.withSearchTag); len(word) > 0 {
		val = word
	}
	text.WriteString(color.ClString)
	text.WriteString(fmt.Sprintf(`%s"%v"`, j.computeIndent(indent), val))
	text.WriteString(color.ClWhite)
}

func (j *JsonView) processNumeric(text *strings.Builder, v interface{}, indent string) {
	if word := j.captureWordSection(v, j.withSearchTag); len(word) > 0 {
		v = word
	}
	text.WriteString(color.ClNumeric)
	text.WriteString(fmt.Sprintf("%s%v", j.computeIndent(indent), v))
	text.WriteString(color.ClWhite)
}

func (j *JsonView) processArrayItem(v interface{}, indent string, text *strings.Builder, last bool) {
	text.WriteString(j.computeIndent(indent + j.indent))

	switch tp := v.(type) {
	case int,
		float64,
		bool:
		j.processNumeric(text, v, "")
	case string:
		j.processString(text, v, "")
	case map[string]interface{}:
		j.processObject(text, v, indent)
	case []interface{}:
		j.processArray(text, tp, indent)
	}
	if !last {
		text.WriteString(",")
	}
}

func (j *JsonView) extractKeys(m map[string]interface{}) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

func (j *JsonView) computeIndent(indent string) string {
	if len(j.indent) > 0 {
		return indent
	}
	return ""
}

func (j *JsonView) newLine() string {
	if len(j.indent) > 0 {
		return "\n"
	}
	return ""
}

func (j *JsonView) captureWordSection(text interface{}, withTag string) string {
	val := fmt.Sprintf("%v", text)
	tagged := len(withTag) > 0
	sel := ""
	if tagged {
		sel = j.searchStrategy.TagWord(withTag, val)
	}
	return sel
}
