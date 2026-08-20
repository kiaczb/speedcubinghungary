package special

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBroadcaster_NotifiesSubscribersOfCompetition(t *testing.T) {
	broadcaster := NewBroadcaster()

	first, unsubscribeFirst := broadcaster.Subscribe("SpeedcubingHungary2026")
	defer unsubscribeFirst()

	second, unsubscribeSecond := broadcaster.Subscribe("SpeedcubingHungary2026")
	defer unsubscribeSecond()

	broadcaster.Notify("SpeedcubingHungary2026")

	for name, events := range map[string]<-chan struct{}{"első": first, "második": second} {
		select {
		case <-events:
		default:
			t.Fatalf("a(z) %s feliratkozó nem kapott jelzést", name)
		}
	}
}

func TestBroadcaster_NotifyIsScopedToCompetition(t *testing.T) {
	broadcaster := NewBroadcaster()

	watched, unsubscribe := broadcaster.Subscribe("SpeedcubingHungary2026")
	defer unsubscribe()

	other, unsubscribeOther := broadcaster.Subscribe("MasikVerseny2026")
	defer unsubscribeOther()

	broadcaster.Notify("SpeedcubingHungary2026")

	select {
	case <-watched:
	default:
		t.Fatal("a verseny nézője nem kapott jelzést")
	}

	select {
	case <-other:
		t.Fatal("másik verseny nézője nem kaphat jelzést")
	default:
	}
}

func TestBroadcaster_CoalescesNotifications(t *testing.T) {
	broadcaster := NewBroadcaster()

	events, unsubscribe := broadcaster.Subscribe("SpeedcubingHungary2026")
	defer unsubscribe()

	for i := 0; i < 100; i++ {
		broadcaster.Notify("SpeedcubingHungary2026")
	}

	received := 0
	for {
		select {
		case <-events:
			received++
			continue
		default:
		}
		break
	}

	assert.Equal(t, 1, received, "a torlódó jelzések egybeolvadnak")
}

func TestBroadcaster_Unsubscribe(t *testing.T) {
	broadcaster := NewBroadcaster()

	_, unsubscribe := broadcaster.Subscribe("SpeedcubingHungary2026")
	require.Equal(t, 1, broadcaster.SubscriberCount("SpeedcubingHungary2026"))

	unsubscribe()
	assert.Zero(t, broadcaster.SubscriberCount("SpeedcubingHungary2026"))

	broadcaster.Notify("SpeedcubingHungary2026")
}

func TestBroadcaster_NilIsSafe(t *testing.T) {
	var broadcaster *Broadcaster
	assert.NotPanics(t, func() { broadcaster.Notify("SpeedcubingHungary2026") })
}
