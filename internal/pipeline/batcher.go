package pipeline

import (
	"context"
	"log/slog"
	"time"

	"github.com/xonoxc/scopion/internal/live"
	"github.com/xonoxc/scopion/internal/model"
	"github.com/xonoxc/scopion/internal/store"
)

const (
	maxBatchSize   = 1000
	flushInterval  = 100 * time.Millisecond
	channelBufSize = 10000
)

type Batcher struct {
	events      chan model.Event
	spans       chan model.Span
	store       store.Storage
	broadcaster *live.Broadcaster
	log         *slog.Logger
}

func New(s store.Storage, b *live.Broadcaster, log *slog.Logger) *Batcher {
	return &Batcher{
		events:      make(chan model.Event, channelBufSize),
		spans:       make(chan model.Span, channelBufSize),
		store:       s,
		broadcaster: b,
		log:         log,
	}
}

func (bp *Batcher) SubmitEvent(e model.Event) {
	select {
	case bp.events <- e:
	default:
		bp.log.Warn("event channel full, dropping event")
	}
}

func (bp *Batcher) SubmitSpan(s model.Span) {
	select {
	case bp.spans <- s:
	default:
		bp.log.Warn("span channel full, dropping span")
	}
}

func (bp *Batcher) Run(ctx context.Context) {
	eventBatch := make([]model.Event, 0, maxBatchSize)
	spanBatch := make([]model.Span, 0, maxBatchSize)
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			bp.flushAll(eventBatch, spanBatch)
			return

		case e := <-bp.events:
			eventBatch = append(eventBatch, e)
			if len(eventBatch) >= maxBatchSize {
				bp.flushEvents(eventBatch)
				eventBatch = eventBatch[:0]
			}

		case s := <-bp.spans:
			spanBatch = append(spanBatch, s)
			if len(spanBatch) >= maxBatchSize {
				bp.flushSpans(spanBatch)
				spanBatch = spanBatch[:0]
			}

		case <-ticker.C:
			if len(eventBatch) > 0 {
				bp.flushEvents(eventBatch)
				eventBatch = eventBatch[:0]
			}
			if len(spanBatch) > 0 {
				bp.flushSpans(spanBatch)
				spanBatch = spanBatch[:0]
			}
		}
	}
}

func (bp *Batcher) flushEvents(batch []model.Event) {
	for _, e := range batch {
		if err := bp.store.Append(e); err != nil {
			bp.log.Error("failed to append event", "error", err)
			continue
		}
		if bp.broadcaster != nil {
			bp.broadcaster.Publish(e)
		}
	}
}

func (bp *Batcher) flushSpans(batch []model.Span) {
	for _, s := range batch {
		if err := bp.store.AppendSpan(s); err != nil {
			bp.log.Error("failed to append span", "error", err)
			continue
		}
		if bp.broadcaster != nil {
			bp.broadcaster.PublishSpan(s)
		}
	}
}

func (bp *Batcher) flushAll(events []model.Event, spans []model.Span) {
	if len(events) > 0 {
		bp.flushEvents(events)
	}
	if len(spans) > 0 {
		bp.flushSpans(spans)
	}
}
