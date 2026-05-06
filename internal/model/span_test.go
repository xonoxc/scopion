package model

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSpanFields(t *testing.T) {
	start := time.Now()
	end := start.Add(100 * time.Millisecond)
	parentID := "parent-123"

	span := Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ParentSpanID: &parentID,
		Name:        "http.request",
		Service:     "api",
		StartTime:   start,
		EndTime:     end,
		Status:      "OK",
	}

	if span.TraceID != "trace-1" {
		t.Errorf("expected TraceID 'trace-1', got %q", span.TraceID)
	}
	if span.SpanID != "span-1" {
		t.Errorf("expected SpanID 'span-1', got %q", span.SpanID)
	}
	if span.ParentSpanID == nil || *span.ParentSpanID != "parent-123" {
		t.Errorf("expected ParentSpanID 'parent-123', got %v", span.ParentSpanID)
	}
	if span.Name != "http.request" {
		t.Errorf("expected Name 'http.request', got %q", span.Name)
	}
	if span.Service != "api" {
		t.Errorf("expected Service 'api', got %q", span.Service)
	}
	if !span.StartTime.Equal(start) {
		t.Errorf("expected StartTime %v, got %v", start, span.StartTime)
	}
	if !span.EndTime.Equal(end) {
		t.Errorf("expected EndTime %v, got %v", end, span.EndTime)
	}
	if span.Status != "OK" {
		t.Errorf("expected Status 'OK', got %q", span.Status)
	}
}

func TestSpanJSONTags(t *testing.T) {
	start := time.Now()
	end := start.Add(100 * time.Millisecond)

	span := Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Name:      "test",
		Service:   "svc",
		StartTime: start,
		EndTime:   end,
		Status:    "OK",
	}

	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("failed to marshal span: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedKeys := []string{"trace_id", "span_id", "name", "service", "start_time", "end_time", "status"}
	for _, key := range expectedKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	if _, ok := m["parent_span_id"]; ok {
		t.Error("parent_span_id should not be present when nil")
	}
}

func TestSpanJSONTagsWithParent(t *testing.T) {
	start := time.Now()
	end := start.Add(100 * time.Millisecond)
	parentID := "parent-123"

	span := Span{
		TraceID:     "trace-1",
		SpanID:      "span-1",
		ParentSpanID: &parentID,
		Name:        "test",
		Service:     "svc",
		StartTime:   start,
		EndTime:     end,
		Status:      "OK",
	}

	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("failed to marshal span: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if _, ok := m["parent_span_id"]; !ok {
		t.Error("parent_span_id should be present when set")
	}
}

func TestSpanZeroValues(t *testing.T) {
	span := Span{}

	if span.TraceID != "" {
		t.Errorf("expected zero TraceID, got %q", span.TraceID)
	}
	if span.SpanID != "" {
		t.Errorf("expected zero SpanID, got %q", span.SpanID)
	}
	if span.ParentSpanID != nil {
		t.Errorf("expected nil ParentSpanID, got %v", span.ParentSpanID)
	}
	if span.Name != "" {
		t.Errorf("expected zero Name, got %q", span.Name)
	}
	if span.Service != "" {
		t.Errorf("expected zero Service, got %q", span.Service)
	}
	if span.Status != "" {
		t.Errorf("expected zero Status, got %q", span.Status)
	}
}
