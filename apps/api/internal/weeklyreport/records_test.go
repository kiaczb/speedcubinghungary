package weeklyreport

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func intPtr(v int) *int { return &v }

func sampleRecords() RecordTable {
	return RecordTable{
		regionWorld:   {"333": {Single: intPtr(300), Average: intPtr(400)}},
		regionEurope:  {"333": {Single: intPtr(350), Average: intPtr(450)}},
		regionHungary: {"333": {Single: intPtr(500), Average: intPtr(600)}},
	}
}

func TestRecordBook_PicksTheHighestBadge(t *testing.T) {
	book := newRecordBook(sampleRecords())

	badge, ok := book.check("2020TEST01", "333", recordSingle, 340)

	assert.True(t, ok)
	assert.Equal(t, "CR", badge)
}

func TestRecordBook_EqualingTheRecordCounts(t *testing.T) {
	book := newRecordBook(sampleRecords())

	badge, ok := book.check("2020TEST01", "333", recordAverage, 600)

	assert.True(t, ok)
	assert.Equal(t, "NR", badge)
}

func TestRecordBook_UpdatesTheRecordSoASlowerResultIsNotARecord(t *testing.T) {
	book := newRecordBook(sampleRecords())

	_, first := book.check("2020TEST01", "333", recordSingle, 450)
	_, second := book.check("2020TEST02", "333", recordSingle, 480)

	assert.True(t, first)
	assert.False(t, second)
}

func TestRecordBook_ExcludedCompetitorGetsNoNationalRecord(t *testing.T) {
	book := newRecordBook(sampleRecords())

	_, ok := book.check("2017TSVE02", "333", recordSingle, 450)

	assert.False(t, ok)
}

func TestRecordBook_InvalidOrUnknownResultsAreIgnored(t *testing.T) {
	book := newRecordBook(sampleRecords())

	_, dnf := book.check("2020TEST01", "333", recordSingle, -1)
	_, unknownEvent := book.check("2020TEST01", "555bf", recordSingle, 100)

	assert.False(t, dnf)
	assert.False(t, unknownEvent)
}
