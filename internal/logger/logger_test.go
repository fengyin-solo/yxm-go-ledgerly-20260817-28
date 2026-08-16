package logger

import (
	"strings"
	"testing"
)

func TestLevelString(t *testing.T) {
	if LevelDebug.String() != "DEBUG" {
		t.Fatal("unexpected debug string")
	}
	if LevelInfo.String() != "INFO" {
		t.Fatal("unexpected info string")
	}
	if LevelWarn.String() != "WARN" {
		t.Fatal("unexpected warn string")
	}
	if LevelError.String() != "ERROR" {
		t.Fatal("unexpected error string")
	}
	if Level(99).String() != "UNKNOWN" {
		t.Fatal("unexpected unknown string")
	}
}

func TestLoggerFiltering(t *testing.T) {
	var buf strings.Builder
	l := New(&buf, LevelWarn)
	l.Info("should be filtered", "key", "value")
	if buf.Len() != 0 {
		t.Fatalf("expected no output for filtered level, got %q", buf.String())
	}
	l.Warn("should appear", "key", "value")
	if buf.Len() == 0 {
		t.Fatal("expected output for warn level")
	}
}

func TestLoggerWith(t *testing.T) {
	var buf strings.Builder
	l := New(&buf, LevelInfo)
	child := l.With("account_id", 42)
	child.Info("test message")
	out := buf.String()
	if !strings.Contains(out, "account_id=42") {
		t.Fatalf("expected account_id in output, got %q", out)
	}
	if !strings.Contains(out, "test message") {
		t.Fatalf("expected message in output, got %q", out)
	}
}

func TestLoggerWithMultipleFields(t *testing.T) {
	var buf strings.Builder
	l := New(&buf, LevelInfo)
	l.Info("fields test", "x", 1, "y", "hello", "z", true)
	out := buf.String()
	if !strings.Contains(out, "x=1") {
		t.Fatalf("expected x=1 in output, got %q", out)
	}
	if !strings.Contains(out, "y=hello") {
		t.Fatalf("expected y=hello in output, got %q", out)
	}
}

func TestNewForTest(t *testing.T) {
	var buf strings.Builder
	l := NewForTest(&buf, LevelInfo)
	l.Info("from test logger")
	out := buf.String()
	if !strings.Contains(out, "from test logger") {
		t.Fatalf("expected message, got %q", out)
	}
}

func TestLoggerChildDoesNotAffectParent(t *testing.T) {
	var parentOut strings.Builder
	parent := New(&parentOut, LevelInfo)
	parent.Info("parent message")
	// Parent writes to parentOut
	if parentOut.Len() == 0 {
		t.Fatal("expected parent output")
	}
	// Child writes to same out but uses fresh buffer
	child := parent.With("trace_id", "abc")
	child.Info("child message")
	// Both should appear in parentOut
	out := parentOut.String()
	if !strings.Contains(out, "parent message") {
		t.Fatalf("expected parent message, got %q", out)
	}
	if !strings.Contains(out, "child message") {
		t.Fatalf("expected child message, got %q", out)
	}
	if !strings.Contains(out, "trace_id=abc") {
		t.Fatalf("expected trace_id from child, got %q", out)
	}
}

func TestDefault(t *testing.T) {
	l := Default()
	if l == nil {
		t.Fatal("expected non-nil default logger")
	}
	if l.level != LevelInfo {
		t.Fatalf("expected LevelInfo, got %v", l.level)
	}
}

func TestLoggerErrorLevel(t *testing.T) {
	var buf strings.Builder
	l := New(&buf, LevelError)
	l.Info("should not appear")
	l.Warn("should not appear")
	l.Error("should appear", "err", "something broke")
	out := buf.String()
	if !strings.Contains(out, "should appear") {
		t.Fatalf("expected error message, got %q", out)
	}
	if strings.Contains(out, "should not appear") {
		t.Fatalf("did not expect info/warn messages, got %q", out)
	}
}
