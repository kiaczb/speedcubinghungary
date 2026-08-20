package special

import "sync"

// Broadcaster továbbítja a böngészőknek, hogy egy verseny adata változott.
//
// Szándékosan minimális: NEM az adatot küldjük szét, csak egy "változott"
// jelzést. Az igazság forrása marad a WCA Live (+ a mi rögzített időnk), a
// kliens a jelzésre újrakérdez. Így nem kell külön szinkronban tartani egy
// második adatfolyamot, és a lemaradt/duplikált jelzés sem okoz gondot: a
// legrosszabb eset egy fölösleges lekérdezés.
//
// A jelzés VERSENY SZERINT kulcsolt: a versenyt a néző választja ki, így
// egyszerre több verseny is nyitva lehet különböző böngészőkben, és egyik
// frissülése nem ránthatja újra a másikat.
type Broadcaster struct {
	mu          sync.Mutex
	nextID      int
	subscribers map[string]map[int]chan struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{subscribers: make(map[string]map[int]chan struct{})}
}

// Subscribe feliratkozik egy verseny változásaira. A visszaadott függvénnyel
// kell leiratkozni (a hívó defer-rel).
func (b *Broadcaster) Subscribe(competitionWCAID string) (<-chan struct{}, func()) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.nextID++
	id := b.nextID

	// 1 kapacitású puffer: ha a kliens épp dolgozik, egy összevont jelzés vár rá.
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

// Notify jelzi, hogy az adott verseny adata változott. Sosem blokkol.
func (b *Broadcaster) Notify(competitionWCAID string) {
	if b == nil {
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, events := range b.subscribers[competitionWCAID] {
		select {
		case events <- struct{}{}:
		default: // már vár egy jelzés, elég egy
		}
	}
}

// SubscriberCount csak teszteléshez/diagnosztikához.
func (b *Broadcaster) SubscriberCount(competitionWCAID string) int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subscribers[competitionWCAID])
}
