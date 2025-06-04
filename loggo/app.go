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
	"github.com/gdamore/tcell/v2"
	"github.com/jimbertools/loggo/config"
	"github.com/jimbertools/loggo/reader"
	"github.com/jimbertools/loggo/util"
	"github.com/rivo/tview"
)

type ViewerOption func(*viewerConfig)

type viewerConfig struct {
	templateFile string
	offset       int64
}

func WithTemplate(templateFile string) ViewerOption {
	return func(c *viewerConfig) {
		c.templateFile = templateFile
	}
}

// WithOffset specifies the offset to start reading from
func WithOffset(offset int64) ViewerOption {
	return func(c *viewerConfig) {
		c.offset = offset
	}
}

type LoggoApp struct {
	appScaffold
	chanReader reader.Reader
	logView    *LogView
}

type Loggo interface {
	Draw()
	SetInputCapture(cap func(event *tcell.EventKey) *tcell.EventKey)
	Stop()
	SetFocus(primitive tview.Primitive)
	ShowPopMessage(text string, waitSecs int64, resetFocusTo tview.Primitive)
	ShowPrefabModal(text string, width, height int, capture inputCapture, buttons ...*tview.Button)
	ShowModal(p tview.Primitive, width, height int, bgColor tcell.Color, capture inputCapture)
	DismissModal(resetFocusTo tview.Primitive)
	Config() *config.Config
	StackView(p tview.Primitive)
	PopView()
}

func StartLogViewer(fileName string, opts ...ViewerOption) {
	c := viewerConfig{}

	for _, opt := range opts {
		opt(&c)
	}

	myReader := reader.MakeReader(fileName, reader.WithOffset(c.offset))
	app := NewLoggoApp(myReader, c.templateFile)
	app.Run()
}

func StartMultiFileLogViewer(fileNames []string, templateFile string) {
	myReader := reader.MakeMultiReader(fileNames, nil)
	app := NewLoggoApp(myReader, templateFile)
	app.Run()
}

func NewLoggoApp(reader reader.Reader, configFile string) *LoggoApp {
	app := NewApp(configFile)
	lapp := &LoggoApp{
		appScaffold: *app,
		chanReader:  reader,
	}

	lapp.logView = NewLogReader(lapp, reader)

	lapp.pages = tview.NewPages().
		AddPage("background", lapp.logView, true, true)

	return lapp
}

func (a *LoggoApp) Run() {
	if err := a.app.
		SetRoot(a.pages, true).
		EnableMouse(true).
		Run(); err != nil {
		util.Log().Error(err)
		panic(err)
	}
}
