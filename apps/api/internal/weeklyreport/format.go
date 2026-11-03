package weeklyreport

import (
	"strings"
	"time"

	"speedcubinghungary/internal/wcaformat"
)

func formatRecord(eventID string, record Record) string {
	return wcaformat.Result(eventID, record.Result, record.Type == recordAverage)
}

func FormatDateRange(from, to time.Time) string {
	if from.Equal(to) {
		return from.Format("2006. Jan. 02.")
	}
	return from.Format("2006. Jan. 02-") + to.Format("02.")
}

func FlagEmoji(countryISO2 string) string {
	if len(countryISO2) != 2 {
		return countryISO2
	}
	var b strings.Builder
	for _, c := range strings.ToUpper(countryISO2) {
		b.WriteRune(127397 + c)
	}
	return b.String()
}
