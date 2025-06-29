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

package reader

type reader struct {
	strChan    chan string
	readerType Type
	onError    func(err error)
}

type Type = int64

type Option func(*readerConfig)

type readerConfig struct {
	strChan chan string
	offset  int64
}

func WithStrChan(strChan chan string) Option {
	return func(c *readerConfig) {
		c.strChan = strChan
	}
}

func WithOffset(offset int64) Option {
	return func(c *readerConfig) {
		c.offset = offset
	}
}

const (
	TypeFile = Type(iota)
	TypePipe
	TypeMultiFile
)

// MakeReader builds a continues file/pipe streamer used to feed the logger. If
// fileName is not provided, it will attempt to consume the input from the stdin.
func MakeReader(fileName string, opts ...Option) Reader {
	c := readerConfig{
		strChan: make(chan string, 1),
		offset:  0,
	}

	for _, opt := range opts {
		opt(&c)
	}

	if len(fileName) > 0 {
		return &fileStream{
			reader: reader{
				strChan:    c.strChan,
				readerType: TypeFile,
			},
			fileName: fileName,
			offset:   c.offset,
		}
	}

	return &readPipeStream{
		reader: reader{
			strChan:    c.strChan,
			readerType: TypePipe,
		},
	}
}

func MakeMultiReader(fileNames []string, strChan chan string) Reader {
	if len(fileNames) == 0 {
		return &readPipeStream{
			reader: reader{
				strChan:    strChan,
				readerType: TypePipe,
			},
		}
	}

	if len(fileNames) == 1 {
		return MakeReader(fileNames[0])
	}

	return MakeMultiFileReader(fileNames, strChan)
}

func (s *reader) ChanReader() <-chan string {
	return s.strChan
}

func (s *reader) ErrorNotifier(onError func(err error)) {
	s.onError = onError
}

func (s *reader) Type() Type {
	return s.readerType
}

type Reader interface {
	// StreamInto feeds the strChan channel for every streamed line.
	StreamInto() error
	// Close finalises and invalidates this stream reader.
	Close()
	// ChanReader returns the outbound channel reader
	ChanReader() <-chan string
	// ErrorNotifier registers a callback func that's called upon fatal streaming log.
	ErrorNotifier(onError func(err error))
}
