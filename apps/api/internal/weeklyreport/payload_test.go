package weeklyreport

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderPostContent_EmptyReportHasEveryKey(t *testing.T) {
	content, err := RenderPostContent(ReportData{})
	require.NoError(t, err)

	assert.JSONEq(t, `{"competitions":[],"competitorMilestones":[],"delegateMilestones":[],"solveMilestones":[]}`, content)
}

func TestRenderPostContent_MatchesTheFrontendShape(t *testing.T) {
	content, err := RenderPostContent(ReportData{
		Competitions: []Competition{{
			Name:        "Miskolc Open 2026",
			CountryISO2: "HU",
			From:        time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC),
			To:          time.Date(2026, 7, 19, 0, 0, 0, 0, time.UTC),
			Competitors: []Competitor{
				{Name: "Rekorder", Events: []EventRecords{{EventID: "333", Records: []Record{{Type: "single", Result: 642, Badge: "NR"}}}}},
				{Name: "Résztvevő"},
			},
		}},
		CompetitorMilestones: []CompetitorMilestone{{Name: "Nagy Anna", CompetitionCount: 99, EstimatedCompetition: "Budapest Open 2026", EstimatedDate: "2026. Aug. 01.", Achieved: false}},
		DelegateMilestones:   []DelegateMilestone{{Name: "Kiss Péter", DelegatedCount: 49}},
		SolveMilestones:      []SolveMilestone{{Name: "Teszt Elek", SolveCount: 4800, Milestone: 5000, Remaining: 200}},
	})
	require.NoError(t, err)

	assert.JSONEq(t, `{
		"competitions": [{
			"name": "Miskolc Open 2026", "countryIso": "HU", "from": "2026-07-18", "to": "2026-07-19",
			"competitors": [
				{"name": "Rekorder", "results": [{"event": "333", "type": "single", "result": "6.42", "badge": "NR"}]},
				{"name": "Résztvevő", "results": []}
			]
		}],
		"competitorMilestones": [{"name": "Nagy Anna", "count": 99, "estimated": "Budapest Open 2026", "estimatedDate": "2026. Aug. 01.", "achieved": false}],
		"delegateMilestones": [{"name": "Kiss Péter", "count": 49}],
		"solveMilestones": [{"name": "Teszt Elek", "count": 4800, "milestone": 5000, "remaining": 200, "achieved": false}]
	}`, content)
	assert.True(t, json.Valid([]byte(content)))
}

func TestPostTitle_ShowsTheActualPeriod(t *testing.T) {
	end := time.Date(2026, 7, 24, 8, 0, 0, 0, time.UTC)

	assert.Equal(t, "Heti riport – 2026. 07. 17. – 2026. 07. 24.", PostTitle(end.Add(-reportPeriod), end))
}

func TestWeekStart_ReturnsMonday(t *testing.T) {
	monday := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	for _, day := range []int{20, 22, 26} {
		assert.Equal(t, monday, WeekStart(time.Date(2026, 7, day, 23, 30, 0, 0, time.UTC)))
	}
	assert.Equal(t, monday.AddDate(0, 0, 7), WeekStart(time.Date(2026, 7, 27, 0, 5, 0, 0, time.UTC)))
}
