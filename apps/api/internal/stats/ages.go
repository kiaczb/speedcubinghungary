package stats

import (
	"math"
	"sort"
	"time"

	"speedcubinghungary/internal/competition"
)

func emptyAges(status AgeStatus) AgeStats {
	return AgeStats{
		Status:   status,
		Youngest: []AgedPerson{},
		Oldest:   []AgedPerson{},
	}
}

func ageAt(birthdate, on time.Time) (int, int) {
	years := on.Year() - birthdate.Year()
	lastBirthday := birthdate.AddDate(years, 0, 0)
	if lastBirthday.After(on) {
		years--
		lastBirthday = birthdate.AddDate(years, 0, 0)
	}
	return years, int(on.Sub(lastBirthday).Hours() / 24)
}

func computeAges(info CompetitionInfo, registrants []competition.Registrant, birthdates map[int]string) AgeStats {
	stats := emptyAges(AgesReady)

	on, err := time.Parse(time.DateOnly, info.StartDate)
	if err != nil {
		return emptyAges(AgesUnavailable)
	}

	aged := make([]AgedPerson, 0, len(registrants))
	exact := make([]float64, 0, len(registrants))
	for _, r := range registrants {
		raw, ok := birthdates[r.WCAUserID]
		if !ok {
			continue
		}
		birthdate, err := time.Parse(time.DateOnly, raw)
		if err != nil || birthdate.After(on) {
			continue
		}
		years, days := ageAt(birthdate, on)
		aged = append(aged, AgedPerson{Person: personOf(r), Years: years, Days: days})
		exact = append(exact, on.Sub(birthdate).Hours()/24/365.25)
	}
	stats.Known = len(aged)
	if len(aged) == 0 {
		return stats
	}

	byAge := func(a AgedPerson) int { return a.Years*400 + a.Days }
	sort.SliceStable(aged, func(i, j int) bool { return byAge(aged[i]) < byAge(aged[j]) })
	stats.Youngest = append([]AgedPerson{}, aged[:min(topCount, len(aged))]...)
	for i := len(aged) - 1; i >= 0 && len(stats.Oldest) < topCount; i-- {
		stats.Oldest = append(stats.Oldest, aged[i])
	}

	sort.Float64s(exact)
	sum := 0.0
	for _, age := range exact {
		sum += age
	}
	stats.Average = math.Round(sum/float64(len(exact))*10) / 10
	middle := len(exact) / 2
	median := exact[middle]
	if len(exact)%2 == 0 {
		median = (exact[middle-1] + exact[middle]) / 2
	}
	stats.Median = math.Round(median*10) / 10
	return stats
}
