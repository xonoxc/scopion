package live

import (
	"sync"

	"github.com/xonoxc/scopion/internal/model"
)

type Message struct {
	Type string
	Data any
}

type Broadcaster struct {
	register    chan chan Message
	unregister  chan chan Message
	publish     chan model.Event
	publishSpan chan model.Span
	subscribers map[chan Message]struct{}
	mu          sync.RWMutex
}

func New() *Broadcaster {
	b := &Broadcaster{
		register:    make(chan chan Message),
		unregister:  make(chan chan Message),
		publish:     make(chan model.Event, 1024),
		publishSpan: make(chan model.Span, 1024),
		subscribers: make(map[chan Message]struct{}),
	}
	go b.run()
	return b
}

func (b *Broadcaster) run() {
	for {
		select {
		case c := <-b.register:
			b.mu.Lock()
			b.subscribers[c] = struct{}{}
			b.mu.Unlock()
		case c := <-b.unregister:
			b.mu.Lock()
			delete(b.subscribers, c)
			b.mu.Unlock()
			close(c)
		case e := <-b.publish:
			msg := Message{Type: "event", Data: e}
			b.broadcast(msg)
		case sp := <-b.publishSpan:
			msg := Message{Type: "span", Data: sp}
			b.broadcast(msg)
		}
	}
}

func (b *Broadcaster) broadcast(msg Message) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for c := range b.subscribers {
		select {
		case c <- msg:
		default:
		}
	}
}

func (b *Broadcaster) Subscribe(ch chan Message) {
	b.register <- ch
}

func (b *Broadcaster) Unsubscribe(ch chan Message) {
	b.unregister <- ch
}

func (b *Broadcaster) Publish(e model.Event) {
	b.publish <- e
}

func (b *Broadcaster) PublishSpan(span model.Span) {
	b.publishSpan <- span
}

func (b *Broadcaster) Stop() {
	close(b.publish)
	close(b.publishSpan)
}
