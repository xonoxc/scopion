package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTraceInfo(t *testing.T) {
	now := time.Now()
	trace := TraceInfo{
		ID:        "trace-1",
		Name:      "test-trace",
		Duration:  100,
		Spans:     5,
		Timestamp: now,
		HasError:  false,
	}

	if trace.ID != "trace-1" {
		t.Errorf("expected ID 'trace-1', got %q", trace.ID)
	}
	if trace.Name != "test-trace" {
		t.Errorf("expected Name 'test-trace', got %q", trace.Name)
	}
}

func TestTraceInfoNoServiceField(t *testing.T) {
	trace := TraceInfo{
		ID:       "trace-1",
		Name:     "test",
		Duration: 100,
		Spans:    5,
	}

	data, err := json.Marshal(trace)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := m["service"]; ok {
		t.Error("TraceInfo JSON should not have 'service' field")
	}
}
