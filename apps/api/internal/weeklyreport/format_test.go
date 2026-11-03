package weeklyreport

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatRecord_PassesTheRecordTypeToTheFormatter(t *testing.T) {
	assert.Equal(t, "28", formatRecord("333fm", Record{Type: recordSingle, Result: 28}))
	assert.Equal(t, "28.33", formatRecord("333fm", Record{Type: recordAverage, Result: 2833}))
	assert.Equal(t, "1:05.50", formatRecord("444", Record{Type: recordSingle, Result: 6550}))
	assert.Equal(t, "9/11 52:31", formatRecord("333mbf", Record{Type: recordSingle, Result: 920315102}))
}

func TestFormatDateRange(t *testing.T) {
	day := func(d int) time.Time { return time.Date(2026, 7, d, 0, 0, 0, 0, time.UTC) }

	assert.Equal(t, "2026. Jul. 18.", FormatDateRange(day(18), day(18)))
	assert.Equal(t, "2026. Jul. 18-19.", FormatDateRange(day(18), day(19)))
}

func TestFlagEmoji(t *testing.T) {
	assert.Equal(t, "🇭🇺", FlagEmoji("HU"))
	assert.Equal(t, "🇦🇹", FlagEmoji("at"))
	assert.Equal(t, "XA1", FlagEmoji("XA1"))
}
