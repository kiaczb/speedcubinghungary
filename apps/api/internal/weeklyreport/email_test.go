package weeklyreport

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func recorder(name string) Competitor {
	return Competitor{Name: name, Events: []EventRecords{{EventID: "333", Records: []Record{{Type: "single", Result: 642, Badge: "NR"}}}}}
}

func competitionOf(name string, hungarian bool, competitors ...Competitor) Competition {
	iso := "AT"
	if hungarian {
		iso = "HU"
	}
	day := time.Date(2026, 7, 18, 0, 0, 0, 0, time.UTC)
	return Competition{Name: name, CountryISO2: iso, From: day, To: day, Hungarian: hungarian, Competitors: competitors}
}

func TestEmailCompetitions_HungarianCompetitionKeepsOnlyRecordHolders(t *testing.T) {
	hu := competitionOf("Miskolc Open 2026", true, recorder("Rekorder"), Competitor{Name: "Résztvevő"})

	selected := EmailCompetitions([]Competition{hu})

	require.Len(t, selected, 1)
	require.Len(t, selected[0].Competitors, 1)
	assert.Equal(t, "Rekorder", selected[0].Competitors[0].Name)
	assert.Len(t, hu.Competitors, 2)
}

func TestEmailCompetitions_DropsHungarianCompetitionWithoutRecords(t *testing.T) {
	hu := competitionOf("Miskolc Open 2026", true, Competitor{Name: "Résztvevő"})

	assert.Empty(t, EmailCompetitions([]Competition{hu}))
}

func TestEmailCompetitions_ForeignCompetitionKeepsEveryHungarian(t *testing.T) {
	foreign := competitionOf("Vienna Open 2026", false, Competitor{Name: "Résztvevő"})

	assert.Equal(t, []Competition{foreign}, EmailCompetitions([]Competition{foreign}))
}

func TestEmailSummary(t *testing.T) {
	summary := EmailSummary([]Competition{
		competitionOf("Miskolc Open 2026", true, recorder("A"), recorder("B"), recorder("C")),
		competitionOf("Vienna Open 2026", false, Competitor{Name: "D"}),
	})

	assert.Equal(t, "Az elmúlt hétvégén A, B és C a Miskolc Open 2026 versenyen vettek részt,\n"+
		"D a Vienna Open 2026 versenyen vett részt.\n\nMindenkinek gratulálunk!", summary)
}

func TestRenderEmail_RendersFilteredCompetitionsAndMilestones(t *testing.T) {
	html, err := RenderEmail(ReportData{
		Competitions: []Competition{
			competitionOf("Miskolc Open 2026", true, recorder("Rekorder"), Competitor{Name: "Csak Résztvevő"}),
			competitionOf("Vienna Open 2026", false, Competitor{Name: "Külföldön Járt"}),
		},
		CompetitorMilestones: []CompetitorMilestone{{Name: "Nagy Anna", CompetitionCount: 100, EstimatedCompetition: "Budapest Open", Achieved: true}},
		DelegateMilestones:   []DelegateMilestone{{Name: "Kiss Péter", DelegatedCount: 49}},
		SolveMilestones:      []SolveMilestone{{Name: "Teszt Elek", SolveCount: 4800, Milestone: 5000, Remaining: 200}},
	})
	require.NoError(t, err)

	assert.Contains(t, html, "Miskolc Open 2026 🇭🇺")
	assert.Contains(t, html, "Rekorder")
	assert.NotContains(t, html, "Csak Résztvevő")
	assert.Contains(t, html, "Külföldön Járt")
	assert.Contains(t, html, `record-badge-type badge-nr`)
	assert.Contains(t, html, "6.42")
	assert.Contains(t, html, "✓ Sikeresen teljesítve:")
	assert.Contains(t, html, "49 verseny")
	assert.Contains(t, html, "Már csak 200 sikeres kirakás a(z) 5000-hoz")
}

func TestRenderEmail_EscapesNames(t *testing.T) {
	html, err := RenderEmail(ReportData{Competitions: []Competition{
		competitionOf("<script>", false, Competitor{Name: "<b>x</b>"}),
	}})
	require.NoError(t, err)

	assert.NotContains(t, html, "<script>")
	assert.NotContains(t, html, "<b>x</b>")
}

func TestRenderEmail_EmptyWeek(t *testing.T) {
	html, err := RenderEmail(ReportData{})
	require.NoError(t, err)

	assert.Contains(t, html, "Nem volt ezen a héten magyar érdekeltségű versenyeredmény.")
	assert.NotContains(t, html, `class="milestone-container"`)
}

func TestBuildMessage_IsBase64HTMLWithHeaders(t *testing.T) {
	msg := string(buildMessage("from@example.com", []string{"a@example.com", "b@example.com"}, "2026-07-24 Weekly Hungarian competitors dump", "<p>Szia ő</p>"))

	headers, body, found := strings.Cut(msg, "\r\n\r\n")
	require.True(t, found)
	assert.Contains(t, headers, "From: Bonsz <from@example.com>")
	assert.Contains(t, headers, "To: a@example.com, b@example.com")
	assert.Contains(t, headers, "Subject: 2026-07-24 Weekly Hungarian competitors dump")
	assert.Contains(t, headers, `Content-Type: text/html; charset="utf-8"`)

	decoded, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(body, "\r\n", ""))
	require.NoError(t, err)
	assert.Equal(t, "<p>Szia ő</p>", string(decoded))
}

func TestNewSMTPMailer_SplitsRecipients(t *testing.T) {
	m := NewSMTPMailer("from@example.com", "pw", "a@example.com, b@example.com,")

	assert.Equal(t, []string{"a@example.com", "b@example.com"}, m.to)
}
