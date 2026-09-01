package wcalive

import "sync"

type Broadcaster struct {
	mu          sync.Mutex
	nextID      int
	subscribers map[string]map[int]chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subscribers: make(map[string]map[int]chan struct{})}
}

func (b *Broadcaster) Subscribe(key string) (<-chan struct{}, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID

	events := make(chan struct{}, 1)
	if b.subscribers[key] == nil {
		b.subscribers[key] = make(map[int]chan struct{})
	}
	b.subscribers[key][id] = events

	return events, func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		delete(b.subscribers[key], id)
		if len(b.subscribers[key]) == 0 {
			delete(b.subscribers, key)
		}
	}
}

func (b *Broadcaster) Notify(key string) {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, events := range b.subscribers[key] {
		select {
		case events <- struct{}{}:
		default:
		}
	}
}

func (b *Broadcaster) SubscriberCount(key string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subscribers[key])
}
