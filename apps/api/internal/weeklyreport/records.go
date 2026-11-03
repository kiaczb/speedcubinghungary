package weeklyreport

import "sync"

const (
	regionWorld   = "world"
	regionEurope  = "europe"
	regionHungary = "hungary"

	recordSingle  = "single"
	recordAverage = "average"
)

var recordRegions = []struct {
	region string
	badge  string
}{
	{regionWorld, "WR"},
	{regionEurope, "CR"},
	{regionHungary, "NR"},
}

var excludedCompetitorWCAIDs = map[string]bool{"2017TSVE02": true}

func isTrackedPerson(countryISO2, wcaID string) bool {
	return countryISO2 == "HU" || excludedCompetitorWCAIDs[wcaID]
}

type recordBook struct {
	mu    sync.Mutex
	table RecordTable
}

func newRecordBook(table RecordTable) *recordBook {
	return &recordBook{table: table}
}

func (b *recordBook) check(wcaID, eventID, recordType string, value int) (string, bool) {
	if value <= 0 {
		return "", false
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	for _, r := range recordRegions {
		if r.region == regionHungary && excludedCompetitorWCAIDs[wcaID] {
			continue
		}
		values, ok := b.table[r.region][eventID]
		if !ok {
			continue
		}
		current := values.Single
		if recordType == recordAverage {
			current = values.Average
		}
		if current == nil || value > *current {
			continue
		}

		updated := value
		if recordType == recordAverage {
			values.Average = &updated
		} else {
			values.Single = &updated
		}
		b.table[r.region][eventID] = values
		return r.badge, true
	}
	return "", false
}
