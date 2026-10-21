package stats

import (
	"context"
	"errors"
	"testing"
	"time"

	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgeAt(t *testing.T) {
	on := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name      string
		birthdate time.Time
		years     int
		days      int
	}{
		{"pont a születésnapján", time.Date(2016, 10, 23, 0, 0, 0, 0, time.UTC), 10, 0},
		{"egy nappal a születésnapja előtt", time.Date(2016, 10, 24, 0, 0, 0, 0, time.UTC), 9, 364},
		{"szökőnapon született", time.Date(2012, 2, 29, 0, 0, 0, 0, time.UTC), 14, 236},
	}
	for _, tc := range cases {
		years, days := ageAt(tc.birthdate, on)
		assert.Equal(t, tc.years, years, tc.name)
		assert.Equal(t, tc.days, days, tc.name)
	}
}

func TestComputeAges(t *testing.T) {
	registrants := []competition.Registrant{
		{WCAUserID: 1, Name: "Kicsi Kata", CountryISO2: "HU"},
		{WCAUserID: 2, Name: "Közép Kornél", CountryISO2: "HU"},
		{WCAUserID: 3, Name: "Öreg Ödön", CountryISO2: "AT"},
		{WCAUserID: 4, Name: "Ismeretlen Imre", CountryISO2: "HU"},
		{WCAUserID: 5, Name: "Jövőbeli János", CountryISO2: "HU"},
	}
	birthdates := map[int]string{
		1: "2019-05-01",
		2: "2000-10-24",
		3: "1960-01-15",
		5: "2030-01-01",
	}

	ages := computeAges(testInfo(), registrants, birthdates)

	assert.Equal(t, AgesReady, ages.Status)
	assert.Equal(t, 3, ages.Known, "hiányzó és értelmetlen születési dátum kimarad")
	require.Len(t, ages.Youngest, 3)
	assert.Equal(t, AgedPerson{Person: Person{Name: "Kicsi Kata", CountryISO2: "HU"}, Years: 7, Days: 175}, ages.Youngest[0])
	assert.Equal(t, "Öreg Ödön", ages.Oldest[0].Person.Name)
	assert.Equal(t, 66, ages.Oldest[0].Years)
	assert.Equal(t, 25, ages.Youngest[1].Years, "a születésnapja előtti napon még 25")

	assert.InDelta(t, 33.4, ages.Average, 0.05)
	assert.InDelta(t, 26.0, ages.Median, 0.1)
}

type fakeTokens struct {
	token string
	err   error
}

func (f fakeTokens) WCAAccessToken(context.Context, int) (string, error) { return f.token, f.err }

type birthdateSource struct {
	CompetitionSource
	calls int
	err   error
}

func (b *birthdateSource) GetBirthdates(competitionID, accessToken string) (map[int]string, error) {
	b.calls++
	if b.err != nil {
		return nil, b.err
	}
	return map[int]string{1: "2010-01-01"}, nil
}

func TestBirthdatesFor_Statuses(t *testing.T) {
	ctx := context.Background()

	noGrant := &Service{tokens: fakeTokens{err: auth.ErrNoGrant}, competitions: &birthdateSource{}, birthdates: map[string]cachedBirthdates{}}
	_, status := noGrant.birthdatesFor(ctx, "Teszt2026", 1)
	assert.Equal(t, AgesNoGrant, status)

	forbiddenSource := &birthdateSource{err: competition.ErrForbidden}
	forbidden := &Service{tokens: fakeTokens{token: "t"}, competitions: forbiddenSource, birthdates: map[string]cachedBirthdates{}}
	_, status = forbidden.birthdatesFor(ctx, "Teszt2026", 1)
	assert.Equal(t, AgesForbidden, status)
	_, status = forbidden.birthdatesFor(ctx, "Teszt2026", 1)
	assert.Equal(t, AgesForbidden, status)
	assert.Equal(t, 1, forbiddenSource.calls, "a 403-at is gyorsítótárazzuk, a 4 mp-es lekérdezés ne hívja újra a WCA-t")

	okSource := &birthdateSource{}
	ready := &Service{tokens: fakeTokens{token: "t"}, competitions: okSource, birthdates: map[string]cachedBirthdates{}}
	birthdates, status := ready.birthdatesFor(ctx, "Teszt2026", 1)
	assert.Equal(t, AgesReady, status)
	assert.Equal(t, "2010-01-01", birthdates[1])
	_, _ = ready.birthdatesFor(ctx, "Teszt2026", 2)
	assert.Equal(t, 2, okSource.calls, "a gyorsítótár felhasználónként külön van")

	ready.tokens = fakeTokens{err: auth.ErrNoGrant}
	_, status = ready.birthdatesFor(ctx, "Teszt2026", 1)
	assert.Equal(t, AgesNoGrant, status, "visszavont engedély után a gyorsítótárból sem jön adat")

	failing := &Service{tokens: fakeTokens{token: "t"}, competitions: &birthdateSource{err: errors.New("status 500")}, birthdates: map[string]cachedBirthdates{}}
	_, status = failing.birthdatesFor(ctx, "Teszt2026", 1)
	assert.Equal(t, AgesUnavailable, status)
}
