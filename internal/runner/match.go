package runner

import (
	"bytes"
	"io"
	"regexp"
	"sync"
)

const (
	maxMatchEventLineBytes  = 512
	matchExcerptContext     = 80
	maxRetainedMatchLineCap = 64 * 1024
)

type matchSink struct {
	mu      sync.Mutex
	sent    bool
	onMatch func(MatchEvent)
}

func (sink *matchSink) alreadySent() bool {
	sink.mu.Lock()
	defer sink.mu.Unlock()
	return sink.sent
}

func (sink *matchSink) notify(event MatchEvent) {
	sink.mu.Lock()
	if sink.sent {
		sink.mu.Unlock()
		return
	}
	sink.sent = true
	sink.mu.Unlock()
	sink.onMatch(event)
}

type lineMatchWriter struct {
	writer  io.Writer
	stream  string
	name    string
	pattern *regexp.Regexp
	sink    *matchSink
	mu      sync.Mutex
	line    []byte
}

func (writer *lineMatchWriter) Write(p []byte) (int, error) {
	n, err := writer.writer.Write(p)
	if n > 0 {
		writer.observe(p[:n])
	}
	return n, err
}

func (writer *lineMatchWriter) observe(p []byte) {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	for len(p) > 0 {
		newline := bytes.IndexByte(p, '\n')
		if newline < 0 {
			writer.line = append(writer.line, p...)
			writer.notifyMatch(bytes.TrimSuffix(writer.line, []byte{'\r'}))
			return
		}

		writer.line = append(writer.line, p[:newline]...)
		writer.matchLine()
		p = p[newline+1:]
	}
}

func (writer *lineMatchWriter) flush() {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if len(writer.line) > 0 {
		writer.matchLine()
	}
}

func (writer *lineMatchWriter) matchLine() {
	writer.notifyMatch(bytes.TrimSuffix(writer.line, []byte{'\r'}))
	if cap(writer.line) > maxRetainedMatchLineCap {
		writer.line = nil
	} else {
		writer.line = writer.line[:0]
	}
}

func (writer *lineMatchWriter) notifyMatch(line []byte) {
	if writer.sink.alreadySent() {
		return
	}
	if match := writer.pattern.FindIndex(line); match != nil {
		writer.sink.notify(MatchEvent{
			Name:    writer.name,
			Pattern: writer.pattern.String(),
			Stream:  writer.stream,
			Line:    matchedLineExcerpt(line, match),
		})
	}
}

func matchedLineExcerpt(line []byte, match []int) string {
	if len(line) <= maxMatchEventLineBytes {
		return string(line)
	}

	start := match[0] - matchExcerptContext
	if start < 0 {
		start = 0
	}
	end := start + maxMatchEventLineBytes
	if end > len(line) {
		end = len(line)
	}

	excerpt := string(line[start:end])
	if start > 0 {
		excerpt = "…" + excerpt
	}
	if end < len(line) {
		excerpt += "…"
	}
	return excerpt
}
