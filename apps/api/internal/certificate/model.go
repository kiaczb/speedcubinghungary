package certificate

import (
	"encoding/json"
	"time"

	"speedcubinghungary/internal/types"
)

type Mode string

const (
	ModePodium      Mode = "podium"
	ModeCompetitors Mode = "competitors"
	ModeRegistrants Mode = "registrants"
)

type Source string

const (
	SourceRegistrations Source = "registrations"
	SourceResults       Source = "results"
	SourceLive          Source = "live"
	SourceNone          Source = "none"
)

type EventStatus string

const (
	EventFinished EventStatus = "finished"
	EventRunning  EventStatus = "running"
	EventPending  EventStatus = "pending"
)

type CompetitionInfo struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	StartDate       string   `json:"startDate"`
	EndDate         string   `json:"endDate"`
	City            string   `json:"city"`
	CountryISO2     string   `json:"countryIso2"`
	URL             string   `json:"url"`
	EventIDs        []string `json:"eventIds"`
	ResultsPostedAt string   `json:"resultsPostedAt,omitempty"`
}

type Event struct {
	EventID string      `json:"eventId"`
	Status  EventStatus `json:"status"`
}

type Recipient struct {
	Name        string `json:"name"`
	WCAID       string `json:"wcaId,omitempty"`
	CountryISO2 string `json:"countryIso2,omitempty"`
	EventID     string `json:"eventId,omitempty"`
	Place       int    `json:"place,omitempty"`
	Result      string `json:"result,omitempty"`
	ResultType  string `json:"resultType,omitempty"`
}

type RecipientList struct {
	Competition CompetitionInfo `json:"competition"`
	Mode        Mode            `json:"mode"`
	Source      Source          `json:"source"`
	FetchedAt   *time.Time      `json:"fetchedAt,omitempty"`
	Events      []Event         `json:"events"`
	Recipients  []Recipient     `json:"recipients"`
}

type Template struct {
	ID            string                        `json:"id"`
	Name          string                        `json:"name"`
	Description   string                        `json:"description"`
	Kind          types.CertificateTemplateKind `json:"kind"`
	Document      json.RawMessage               `json:"document"`
	Version       int                           `json:"version"`
	UpdatedByName string                        `json:"updatedByName"`
	CreatedAt     time.Time                     `json:"createdAt"`
	UpdatedAt     time.Time                     `json:"updatedAt"`
}

type TemplateInput struct {
	Name        string                        `json:"name"`
	Description string                        `json:"description"`
	Kind        types.CertificateTemplateKind `json:"kind"`
	Document    json.RawMessage               `json:"document"`
}

type Asset struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Mime      string    `json:"mime"`
	Width     int       `json:"width"`
	Height    int       `json:"height"`
	Size      int       `json:"size"`
	CreatedAt time.Time `json:"createdAt"`
}

type AssetData struct {
	Mime string
	Data []byte
}
