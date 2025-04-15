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

// Package pkg provides the main API for using loggo as an importable package.
package pkg

import (
	"github.com/marawanxmamdouh/loggo/loggo"
	"github.com/marawanxmamdouh/loggo/reader"
)

// LoggoApp represents the main loggo application.
type LoggoApp struct {
	app *loggo.LoggoApp
}

// NewLoggoApp creates a new loggo application instance.
// If fileName is provided, it will read from that file.
// If fileName is empty, it will read from stdin.
// If templateFile is provided, it will use that template for rendering.
func NewLoggoApp(fileName, templateFile string) *LoggoApp {
	reader := reader.MakeReader(fileName, nil)
	app := loggo.NewLoggoApp(reader, templateFile)
	return &LoggoApp{
		app: app,
	}
}

// Run starts the loggo application.
func (a *LoggoApp) Run() {
	a.app.Run()
}

// Version returns the current version of loggo.
func Version() string {
	return loggo.BuildVersion
}

// SetVersion sets the version of loggo.
func SetVersion(version string) {
	loggo.BuildVersion = version
}
