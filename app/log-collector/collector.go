package logcollector

import (
	"bytes"
	"io"

	thtml "github.com/buildkite/terminal-to-html/v3"
)

type HTMLLogCollector struct {
	buf       *bytes.Buffer
	ansiChans [](chan string)
	htmlChans [](chan string)
}

func NewLogCollector() *HTMLLogCollector {
	return &HTMLLogCollector{
		buf: bytes.NewBuffer([]byte{}),
	}
}

var _ io.ReadWriteCloser = &HTMLLogCollector{}

func (h *HTMLLogCollector) Close() error {
	for i := range h.ansiChans {
		close(h.ansiChans[i])
	}
	for i := range h.htmlChans {
		close(h.htmlChans[i])
	}
	return nil
}

func (h *HTMLLogCollector) Write(b []byte) (n int, err error) {
	n, err = h.buf.Write(b)
	if err != nil {
		return
	}
	for i := range h.ansiChans {
		h.ansiChans[i] <- string(b)
	}
	for i := range h.htmlChans {
		h.htmlChans[i] <- string(thtml.Render(b))
	}

	return
}

func (h *HTMLLogCollector) Read(p []byte) (int, error) {
	return h.buf.Read(p)
}

func (h *HTMLLogCollector) ANSI() chan string {
	ch := make(chan string)
	h.ansiChans = append(h.ansiChans, ch)
	return ch
}

func (h *HTMLLogCollector) HTML() chan string {
	ch := make(chan string)
	h.htmlChans = append(h.htmlChans, ch)
	return ch
}
