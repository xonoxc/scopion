package live

import "github.com/xonoxc/scopion/internal/model"

type Broadcaster struct {
	register   chan chan model.Event
	unregister chan chan model.Event
	publish    chan model.Event
}

func New() *Broadcaster {
	b := &Broadcaster{
		register:   make(chan chan model.Event),
		unregister: make(chan chan model.Event),
		publish:    make(chan model.Event, 1024),
	}
	go b.run()
	return b
}

func (b *Broadcaster) run() {
	clients := map[chan model.Event]struct{}{}
	for {
		select {
		case c := <-b.register:
			clients[c] = struct{}{}
		case c := <-b.unregister:
			delete(clients, c)
			close(c)
		case e := <-b.publish:
			for c := range clients {
				select {
				case c <- e:
				default:
					delete(clients, c)
				}
			}
		}
	}
}

func (b *Broadcaster) Publish(e model.Event) {
	b.publish <- e
}
