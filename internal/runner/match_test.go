package runner

import (
	"bytes"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestLineMatchWriterMatchesAcrossChunksWithoutNewline(t *testing.T) {
	var output bytes.Buffer
	events := make(chan MatchEvent, 1)
	sink := &matchSink{onMatch: func(event MatchEvent) {
		events <- event
	}}
	writer := &lineMatchWriter{
		writer:  &output,
		stream:  "stdout",
		name:    "Agent task",
		pattern: regexp.MustCompile(`approve|waiting`),
		sink:    sink,
	}

	for _, chunk := range []string{"still ", "wait"} {
		if _, err := writer.Write([]byte(chunk)); err != nil {
			t.Fatalf("Write(%q) error = %v", chunk, err)
		}
		select {
		case event := <-events:
			t.Fatalf("matched incomplete pattern after %q: %#v", chunk, event)
		default:
		}
	}
	if _, err := writer.Write([]byte("ing")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case event := <-events:
		if event.Name != "Agent task" || event.Stream != "stdout" || event.Pattern != "approve|waiting" || event.Line != "still waiting" {
			t.Fatalf("event = %#v, want complete match metadata", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for live unterminated-line match")
	}

	writer.flush()
	if output.String() != "still waiting" {
		t.Fatalf("output = %q, want unchanged output", output.String())
	}
}

func TestMatchSinkSendsOnlyFirstMatchingLine(t *testing.T) {
	events := make(chan MatchEvent, 2)
	sink := &matchSink{onMatch: func(event MatchEvent) {
		events <- event
	}}
	writer := &lineMatchWriter{
		writer:  &bytes.Buffer{},
		stream:  "stderr",
		name:    "tests",
		pattern: regexp.MustCompile(`failed`),
		sink:    sink,
	}

	if _, err := writer.Write([]byte("first failed\nsecond failed\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case event := <-events:
		if event.Line != "first failed" || event.Stream != "stderr" {
			t.Fatalf("event = %#v, want first stderr match", event)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for first match")
	}
	select {
	case event := <-events:
		t.Fatalf("unexpected duplicate match: %#v", event)
	default:
	}
}

func TestLineMatchWriterIncludesMatchFromLongLine(t *testing.T) {
	events := make(chan MatchEvent, 1)
	writer := &lineMatchWriter{
		writer:  &bytes.Buffer{},
		stream:  "stdout",
		name:    "build",
		pattern: regexp.MustCompile(`approval required$`),
		sink: &matchSink{onMatch: func(event MatchEvent) {
			events <- event
		}},
	}
	line := strings.Repeat("x", 70*1024) + " approval required\n"

	if _, err := writer.Write([]byte(line)); err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	select {
	case event := <-events:
		if !strings.Contains(event.Line, "approval required") || len(event.Line) > maxMatchEventLineBytes+len("……") {
			t.Fatalf("event line length = %d, value %q; want short matching excerpt", len(event.Line), event.Line)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for long-line match")
	}
}

func TestLineMatchWriterDoesNotMatchOtherLines(t *testing.T) {
	events := make(chan MatchEvent, 1)
	writer := &lineMatchWriter{
		writer:  &bytes.Buffer{},
		stream:  "stdout",
		name:    "tests",
		pattern: regexp.MustCompile(`failed`),
		sink: &matchSink{onMatch: func(event MatchEvent) {
			events <- event
		}},
	}

	if _, err := writer.Write([]byte("all clear\n")); err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	writer.flush()

	select {
	case event := <-events:
		t.Fatalf("unexpected match: %#v", event)
	default:
	}
}
