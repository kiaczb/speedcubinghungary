package weeklyreport

import (
	"encoding/json"
	"fmt"
	"time"
)

type reportPayload struct {
	Competitions         []payloadCompetition         `json:"competitions"`
	CompetitorMilestones []payloadCompetitorMilestone `json:"competitorMilestones"`
	DelegateMilestones   []payloadDelegateMilestone   `json:"delegateMilestones"`
	SolveMilestones      []payloadSolveMilestone      `json:"solveMilestones"`
}

type payloadCompetition struct {
	Name        string              `json:"name"`
	CountryISO  string              `json:"countryIso"`
	From        string              `json:"from"`
	To          string              `json:"to"`
	Competitors []payloadCompetitor `json:"competitors"`
}

type payloadCompetitor struct {
	Name    string          `json:"name"`
	Results []payloadResult `json:"results"`
}

type payloadResult struct {
	Event  string `json:"event"`
	Type   string `json:"type"`
	Result string `json:"result"`
	Badge  string `json:"badge"`
}

type payloadCompetitorMilestone struct {
	Name          string `json:"name"`
	Count         int    `json:"count"`
	Estimated     string `json:"estimated"`
	EstimatedDate string `json:"estimatedDate"`
	Achieved      bool   `json:"achieved"`
}

type payloadDelegateMilestone struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type payloadSolveMilestone struct {
	Name      string `json:"name"`
	Count     int    `json:"count"`
	Milestone int    `json:"milestone"`
	Remaining int    `json:"remaining"`
	Achieved  bool   `json:"achieved"`
}

func RenderPostContent(data ReportData) (string, error) {
	payload := reportPayload{
		Competitions:         make([]payloadCompetition, 0, len(data.Competitions)),
		CompetitorMilestones: make([]payloadCompetitorMilestone, 0, len(data.CompetitorMilestones)),
		DelegateMilestones:   make([]payloadDelegateMilestone, 0, len(data.DelegateMilestones)),
		SolveMilestones:      make([]payloadSolveMilestone, 0, len(data.SolveMilestones)),
	}

	for _, comp := range data.Competitions {
		pc := payloadCompetition{
			Name:        comp.Name,
			CountryISO:  comp.CountryISO2,
			From:        comp.From.Format(time.DateOnly),
			To:          comp.To.Format(time.DateOnly),
			Competitors: make([]payloadCompetitor, 0, len(comp.Competitors)),
		}
		for _, competitor := range comp.Competitors {
			results := make([]payloadResult, 0)
			for _, event := range competitor.Events {
				for _, record := range event.Records {
					results = append(results, payloadResult{
						Event:  event.EventID,
						Type:   record.Type,
						Result: formatRecord(event.EventID, record),
						Badge:  record.Badge,
					})
				}
			}
			pc.Competitors = append(pc.Competitors, payloadCompetitor{Name: competitor.Name, Results: results})
		}
		payload.Competitions = append(payload.Competitions, pc)
	}

	for _, m := range data.CompetitorMilestones {
		payload.CompetitorMilestones = append(payload.CompetitorMilestones, payloadCompetitorMilestone{
			Name:          m.Name,
			Count:         m.CompetitionCount,
			Estimated:     m.EstimatedCompetition,
			EstimatedDate: m.EstimatedDate,
			Achieved:      m.Achieved,
		})
	}
	for _, m := range data.DelegateMilestones {
		payload.DelegateMilestones = append(payload.DelegateMilestones, payloadDelegateMilestone{
			Name:  m.Name,
			Count: m.DelegatedCount,
		})
	}
	for _, m := range data.SolveMilestones {
		payload.SolveMilestones = append(payload.SolveMilestones, payloadSolveMilestone{
			Name:      m.Name,
			Count:     m.SolveCount,
			Milestone: m.Milestone,
			Remaining: m.Remaining,
			Achieved:  m.Achieved,
		})
	}

	content, err := json.Marshal(payload)
	return string(content), err
}

func PostTitle(periodStart, periodEnd time.Time) string {
	return fmt.Sprintf("Heti riport – %s – %s", periodStart.Format("2006. 01. 02."), periodEnd.Format("2006. 01. 02."))
}

func WeekStart(day time.Time) time.Time {
	offset := (int(day.Weekday()) + 6) % 7
	date := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.UTC)
	return date.AddDate(0, 0, -offset)
}
