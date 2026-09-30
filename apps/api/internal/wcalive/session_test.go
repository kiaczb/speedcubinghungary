package wcalive

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	waitTimeout = 2 * time.Second
	waitTick    = 10 * time.Millisecond
	neverWindow = 150 * time.Millisecond
)

type fakeWatcher struct {
	mu      sync.Mutex
	reloads int
	events  []SubscriptionEvent
}

func (w *fakeWatcher) Reload(context.Context) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.reloads++
}

func (w *fakeWatcher) WantedSubscriptions() map[string]Subscription {
	return map[string]Subscription{}
}

func (w *fakeWatcher) HandleEvent(event SubscriptionEvent) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
}

func (w *fakeWatcher) reloadCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.reloads
}

func newTestSession(t *testing.T) (*Session, *fakeWatcher) {
	t.Helper()

	watcher := &fakeWatcher{}
	session := NewSession("teszt", NewSocketClient("http://example.invalid"), time.Hour)
	session.Watch(watcher)

	return session, watcher
}

func TestSession_RunsOnlyWhileViewersArePresent(t *testing.T) {
	session, watcher := newTestSession(t)

	assert.False(t, session.Running(), "néző nélkül nem fut session")

	releaseFirst, startedFirst := session.AddViewer()
	assert.True(t, startedFirst, "az első néző indítja a session-t")

	require.Eventually(t, func() bool {
		return session.Running() && watcher.reloadCount() >= 1
	}, waitTimeout, waitTick, "indulás után azonnal be kell töltenie")

	releaseSecond, startedSecond := session.AddViewer()
	assert.False(t, startedSecond,
		"a másodiknak tudnia kell, hogy nem ő indította — különben nem töltené be, amit néz")

	releaseFirst()
	assert.Never(t, func() bool {
		return !session.Running()
	}, neverWindow, waitTick, "amíg van néző, fut a session")

	releaseSecond()
	assert.Eventually(t, func() bool {
		return !session.Running()
	}, waitTimeout, waitTick, "az utolsó néző távozásakor le kell állni")
}

func TestSession_ReleaseIsIdempotent(t *testing.T) {
	session, _ := newTestSession(t)

	releaseFirst, _ := session.AddViewer()
	releaseSecond, _ := session.AddViewer()

	require.Eventually(t, session.Running, waitTimeout, waitTick)

	releaseFirst()
	releaseFirst()

	assert.Never(t, func() bool {
		return !session.Running()
	}, neverWindow, waitTick, "a kétszer lezárt nézet nem állíthatja le a session-t")

	releaseSecond()
	assert.Eventually(t, func() bool {
		return !session.Running()
	}, waitTimeout, waitTick)
}

func TestSession_RestartsAfterFullStop(t *testing.T) {
	session, watcher := newTestSession(t)

	release, _ := session.AddViewer()
	require.Eventually(t, func() bool {
		return watcher.reloadCount() >= 1
	}, waitTimeout, waitTick)

	release()
	require.Eventually(t, func() bool {
		return !session.Running()
	}, waitTimeout, waitTick)

	before := watcher.reloadCount()

	release2, started := session.AddViewer()
	defer release2()

	assert.True(t, started, "a leállás után az új néző megint indít")
	assert.Eventually(t, func() bool {
		return watcher.reloadCount() > before
	}, waitTimeout, waitTick, "a második session is lefut")
}

func TestSession_ContextFollowsTheSession(t *testing.T) {
	session, _ := newTestSession(t)

	assert.NoError(t, session.Context().Err(), "session nélkül háttér-context jár")

	release, _ := session.AddViewer()
	sessionCtx := session.Context()
	require.NoError(t, sessionCtx.Err())

	release()

	assert.Eventually(t, func() bool {
		return sessionCtx.Err() != nil
	}, waitTimeout, waitTick, "a leállás a context-et is lezárja")
}

func TestRoundSubscriptionKeyRoundTrip(t *testing.T) {
	roundID, err := RoundIDFromKey(RoundSubscriptionKey(147464))

	require.NoError(t, err)
	assert.Equal(t, 147464, roundID)

	_, err = RoundIDFromKey("nem-ilyen-kulcs")
	assert.Error(t, err)
}
