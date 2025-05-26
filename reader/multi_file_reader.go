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
	"fmt"
	"sync"

	"github.com/nxadm/tail"
)

type multiFileStream struct {
	reader
	fileNames []string
	tails     []*tail.Tail
	wg        sync.WaitGroup
	mu        sync.Mutex
}

// MakeMultiFileReader builds a reader that can stream from multiple files simultaneously
func MakeMultiFileReader(fileNames []string, strChan chan string) Reader {
	if strChan == nil {
		strChan = make(chan string, len(fileNames))
	}
	return &multiFileStream{
		reader: reader{
			strChan:    strChan,
			readerType: TypeMultiFile,
		},
		fileNames: fileNames,
	}
}

func (s *multiFileStream) StreamInto() error {
	s.tails = make([]*tail.Tail, 0, len(s.fileNames))

	for _, fileName := range s.fileNames {
		t, err := tail.TailFile(fileName, tail.Config{Follow: true, Poll: true})
		if err != nil {
			s.closeTails()
			return fmt.Errorf("failed to tail file %s: %w", fileName, err)
		}

		s.mu.Lock()
		s.tails = append(s.tails, t)
		s.mu.Unlock()

		s.wg.Add(1)
		go func(t *tail.Tail, fileName string) {
			defer s.wg.Done()
			for line := range t.Lines {
				// TODO: consider adding file name to the line text
				//s.strChan <- fmt.Sprintf("[%s] %s", fileName, line.Text)
				s.strChan <- line.Text
			}
		}(t, fileName)
	}

	return nil
}

func (s *multiFileStream) closeTails() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, t := range s.tails {
		t.Kill(fmt.Errorf("stopped by Close method"))
	}
}

func (s *multiFileStream) Close() {
	s.closeTails()
	s.wg.Wait()
	close(s.strChan)
}
