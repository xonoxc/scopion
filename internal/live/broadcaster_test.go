package live

import (
	"testing"
	"time"

	"github.com/xonoxc/scopion/internal/model"
)

func TestPublishSpan(t *testing.T) {
	b := New()
	defer b.Stop()

	ch := make(chan Message, 1)
	b.Subscribe(ch)
	defer b.Unsubscribe(ch)

	span := model.Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Name:      "test-span",
		Service:   "test-svc",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(100 * time.Millisecond),
		Status:    "OK",
	}

	b.PublishSpan(span)

	select {
	case msg := <-ch:
		if msg.Type != "span" {
			t.Errorf("Expected message type 'span', got '%s'", msg.Type)
		}
		spanData, ok := msg.Data.(model.Span)
		if !ok {
			t.Fatalf("Expected span data to be model.Span, got %T", msg.Data)
		}
		if spanData.TraceID != "trace-1" {
			t.Errorf("Expected trace_id 'trace-1', got '%s'", spanData.TraceID)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for span message")
	}
}

func TestPublishSpanMultipleSubscribers(t *testing.T) {
	b := New()
	defer b.Stop()

	ch1 := make(chan Message, 1)
	ch2 := make(chan Message, 1)
	b.Subscribe(ch1)
	b.Subscribe(ch2)
	defer b.Unsubscribe(ch1)
	defer b.Unsubscribe(ch2)

	span := model.Span{
		TraceID: "trace-1",
		SpanID:  "span-1",
		Name:    "test-span",
	}

	b.PublishSpan(span)

	for i, ch := range []chan Message{ch1, ch2} {
		select {
		case msg := <-ch:
			if msg.Type != "span" {
				t.Errorf("Subscriber %d: Expected message type 'span', got '%s'", i+1, msg.Type)
			}
		case <-time.After(time.Second):
			t.Fatalf("Subscriber %d: Timed out waiting for span message", i+1)
		}
	}
}

func TestPublishEvent(t *testing.T) {
	b := New()
	defer b.Stop()

	ch := make(chan Message, 1)
	b.Subscribe(ch)
	defer b.Unsubscribe(ch)

	event := model.Event{
		ID:        "event-1",
		Level:     "info",
		Service:   "test-svc",
		Name:      "test-event",
		Timestamp: time.Now(),
	}

	b.Publish(event)

	select {
	case msg := <-ch:
		if msg.Type != "event" {
			t.Errorf("Expected message type 'event', got '%s'", msg.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("Timed out waiting for event message")
	}
}
