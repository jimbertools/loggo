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

package main

import (
	"github.com/gdamore/tcell/v2"
	"github.com/jimbertools/loggo/color"
	"github.com/jimbertools/loggo/loggo"
	"github.com/rivo/tview"
)

func main() {
	app := loggo.NewApp("config-sample/gcp.yaml")
	main := tview.NewFlex().SetDirection(tview.FlexRow)
	main.AddItem(loggo.NewFilterView(app, nil), 4, 1, true).
		AddItem(loggo.NewHorizontalSeparator(color.FieldStyle, loggo.LineHThick, "test", tcell.ColorYellow), 1, 2, false)
	app.Run(main)
}
