package live

import (
	"sync"

	"github.com/xonoxc/scopion/internal/model"
)

type Message struct {
	Type string
	Data any
}

type subscriber struct {
	ch   chan Message
	done chan struct{}
}

type Broadcaster struct {
	register    chan subscriber
	unregister  chan chan Message
	publish     chan model.Event
	publishSpan chan model.Span
	subscribers map[chan Message]subscriber
	mu          sync.RWMutex
	wg          sync.WaitGroup
	stopped     chan struct{}
}

func New() *Broadcaster {
	b := &Broadcaster{
		register:    make(chan subscriber),
		unregister:  make(chan chan Message),
		publish:     make(chan model.Event, 1024),
		publishSpan: make(chan model.Span, 1024),
		subscribers: make(map[chan Message]subscriber),
		stopped:     make(chan struct{}),
	}
	b.wg.Add(1)
	go b.run()
	return b
}

func (b *Broadcaster) run() {
	defer b.wg.Done()
	for {
		select {
		case <-b.stopped:
			return
		case s := <-b.register:
			b.mu.Lock()
			b.subscribers[s.ch] = s
			b.mu.Unlock()
		case c := <-b.unregister:
			b.mu.Lock()
			if s, ok := b.subscribers[c]; ok {
				close(s.done)
				delete(b.subscribers, c)
			}
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
	for c, s := range b.subscribers {
		select {
		case c <- msg:
		case <-s.done:
			go func(ch chan Message) {
				b.unregister <- ch
			}(c)
		default:
		}
	}
}

func (b *Broadcaster) Subscribe(ch chan Message) {
	s := subscriber{ch: ch, done: make(chan struct{})}
	b.register <- s
}

func (b *Broadcaster) Unsubscribe(ch chan Message) {
	b.unregister <- ch
}

func (b *Broadcaster) Publish(e model.Event) {
	select {
	case b.publish <- e:
	case <-b.stopped:
	}
}

func (b *Broadcaster) PublishSpan(span model.Span) {
	select {
	case b.publishSpan <- span:
	case <-b.stopped:
	}
}

func (b *Broadcaster) Stop() {
	close(b.stopped)
	b.wg.Wait()
}
