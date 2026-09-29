package competition

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ⚠️ A WCA API kb. 15 párhuzamos kérés után 429-et ad. Egy tippverseny-oldal
// versenyszámonként kér egy psych sheetet, és a singleflight csak AZONOS
// kulcsokat von össze — hideg cache mellett egy tucat különböző versenyszám
// egyszerre ment ki. A korlátnak a kimenő oldalon kell ülnie, mert ott
// független attól, ki és honnan kérdez.
func TestWCARepository_BoundsConcurrentUpstreamRequests(t *testing.T) {
	var inFlight, peak atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := inFlight.Add(1)
		defer inFlight.Add(-1)

		for {
			old := peak.Load()
			if now <= old || peak.CompareAndSwap(old, now) {
				break
			}
		}

		time.Sleep(40 * time.Millisecond)
		fmt.Fprint(w, `{"sorted_rankings":[]}`)
	}))
	t.Cleanup(server.Close)

	repo := newWCARepository(server.URL)

	var wg sync.WaitGroup
	for i := 0; i < 3*maxConcurrentUpstream; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, err := repo.GetPsychSheet("Verseny2026", fmt.Sprintf("event%d", i))
			assert.NoError(t, err)
		}(i)
	}
	wg.Wait()

	require.Positive(t, peak.Load())
	assert.LessOrEqual(t, int(peak.Load()), maxConcurrentUpstream,
		"egyszerre legfeljebb maxConcurrentUpstream kérés lehet úton a WCA felé")
}

// A túl nagy válasz nem kerülhet teljes egészében a memóriába.
func TestWCARepository_CapsTheResponseBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"persons":[`)
		chunk := []byte(`{"name":"xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"},`)
		for written := 0; written <= maxUpstreamBody; written += len(chunk) {
			if _, err := w.Write(chunk); err != nil {
				return
			}
		}
	}))
	t.Cleanup(server.Close)

	_, err := newWCARepository(server.URL).GetWCIF("Verseny2026")
	assert.Error(t, err, "a korlát fölötti választ nem szabad dekódolni")
}

// A GetCompetition bármely dátumú versenyt visszaad ID szerint; a WCA 404-e
// saját sentinel hibává válik, hogy a hívó meg tudja különböztetni a nem
// létező versenyt a WCA kiesésétől.
func TestWCARepository_GetCompetition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/BudapestOpen2025" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"id":"BudapestOpen2025","name":"Budapest Open 2025","start_date":"2025-03-01","end_date":"2025-03-02","city":"Budapest","country_iso2":"HU"}`)
	}))
	t.Cleanup(server.Close)

	repo := newWCARepository(server.URL)

	competition, err := repo.GetCompetition("BudapestOpen2025")
	require.NoError(t, err)
	assert.Equal(t, "Budapest Open 2025", competition.Name)
	assert.Equal(t, "HU", competition.CountryISO2)

	_, err = repo.GetCompetition("NemLetezo2099")
	assert.ErrorIs(t, err, ErrNotFound)
}
