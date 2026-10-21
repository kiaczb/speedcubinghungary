package stats

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatResult(t *testing.T) {
	cases := []struct {
		name    string
		eventID string
		value   int
		average bool
		want    string
	}{
		{"másodperc alatt", "333", 642, false, "6.42"},
		{"perc fölött", "666", 8834, false, "1:28.34"},
		{"óra fölött", "555bf", 372345, false, "1:02:03.45"},
		{"DNF", "333", -1, false, "DNF"},
		{"DNS", "333", -2, false, "DNS"},
		{"üres", "333", 0, false, ""},
		{"FMC single lépésszám", "333fm", 24, false, "24"},
		{"FMC átlag két tizedessel", "333fm", 2633, true, "26.33"},
		{"MBLD 8/9 52:31", "333mbf", 920315101, false, "8/9 52:31"},
		{"MBLD ismeretlen idővel", "333mbf", 929999901, false, "8/9"},
		{"régi MBLD új kódolással", "333mbo", 970107500, false, "2/2 17:55"},
		{"régi MBLD régi kódolással", "333mbo", 1960303600, false, "3/3 60:00"},
	}

	for _, tc := range cases {
		assert.Equal(t, tc.want, FormatResult(tc.eventID, tc.value, tc.average), tc.name)
	}
}
