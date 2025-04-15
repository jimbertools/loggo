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
//
// Loggo is a log viewer and analyzer that can read from files or stdin.
// It provides a rich UI for viewing and filtering logs.
//
// Basic usage:
//
//	import "github.com/marawanxmamdouh/loggo/pkg"
//
//	func main() {
//		app := pkg.NewLoggoApp("path/to/logfile.log", "path/to/template.yaml")
//		app.Run()
//	}
//
// To create a custom reader:
//
//	reader := pkg.NewReader("path/to/logfile.log", nil)
//	err := reader.StreamInto()
//	if err != nil {
//		// handle error
//	}
//	defer reader.Close()
//
//	for line := range reader.ChanReader() {
//		// process each line
//	}
package pkg
