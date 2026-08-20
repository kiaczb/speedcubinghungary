package special

import "sync"

type Broadcaster struct {
	mu          sync.Mutex
	nextID      int
	subscribers map[string]map[int]chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subscribers: make(map[string]map[int]chan struct{})}
}

func (b *Broadcaster) Subscribe(competitionWCAID string) (<-chan struct{}, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID

	events := make(chan struct{}, 1)
	if b.subscribers[competitionWCAID] == nil {
		b.subscribers[competitionWCAID] = make(map[int]chan struct{})
	}
	b.subscribers[competitionWCAID][id] = events

	return events, func() {
		b.mu.Lock()
		defer b.mu.Unlock()

		delete(b.subscribers[competitionWCAID], id)
		if len(b.subscribers[competitionWCAID]) == 0 {
			delete(b.subscribers, competitionWCAID)
		}
	}
}

func (b *Broadcaster) Notify(competitionWCAID string) {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, events := range b.subscribers[competitionWCAID] {
		select {
		case events <- struct{}{}:
		default:
		}
	}
}

func (b *Broadcaster) SubscriberCount(competitionWCAID string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subscribers[competitionWCAID])
}
