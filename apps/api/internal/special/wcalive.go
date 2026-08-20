package special

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DefaultWCALiveURL a WCA Live nyilvános példánya. A GraphQL végpont a /api,
// a subscription pedig a Phoenix socket — mindkettő autentikáció nélkül
// elérhető (wca_live UserSocket.connect/3 minden kapcsolatot elfogad).
const DefaultWCALiveURL = "https://live.worldcubeassociation.org"

// competitionIDFromPath a /link/... végpont átirányításából szedi ki a WCA Live
// numerikus verseny-azonosítóját: "/competitions/11151" → 11151.
var competitionIDFromPath = regexp.MustCompile(`/competitions/(\d+)`)

// LiveClient a WCA Live GraphQL API HTTP oldalát szólítja meg (lekérdezések).
// A subscription külön, WebSocketen megy — lásd socket.go.
type LiveClient struct {
	client  *http.Client
	baseURL string
}

func NewLiveClient(baseURL string) *LiveClient {
	if baseURL == "" {
		baseURL = DefaultWCALiveURL
	}
	return &LiveClient{
		// A /link/ végpont 302-vel válaszol, amit magunk dolgozunk fel.
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

func (c *LiveClient) query(ctx context.Context, doc string, variables map[string]any, out any) error {
	body, err := json.Marshal(graphQLRequest{Query: doc, Variables: variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WCA Live GraphQL returned status %d", resp.StatusCode)
	}

	var parsed graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}
	if len(parsed.Errors) > 0 {
		return fmt.Errorf("WCA Live GraphQL error: %s", parsed.Errors[0].Message)
	}

	return json.Unmarshal(parsed.Data, out)
}

// ResolveCompetitionID a WCA verseny-azonosítóból (pl. "HungarianAllRounder62026")
// megadja a WCA Live numerikus id-ját.
//
// A GraphQL sémában nincs wcaId szerinti lekérdezés (a competitions(filter:)
// csak névre keres), ezért a WCA Live saját, dokumentált /link/ átirányítását
// használjuk — ugyanezt teszi a WCA Live kliense is a megosztható linkekhez.
func (c *LiveClient) ResolveCompetitionID(ctx context.Context, competitionWCAID string) (int, error) {
	url := fmt.Sprintf("%s/link/competitions/%s", c.baseURL, competitionWCAID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		return 0, fmt.Errorf("competition %q not found on WCA Live (status %d)", competitionWCAID, resp.StatusCode)
	}

	match := competitionIDFromPath.FindStringSubmatch(resp.Header.Get("Location"))
	if match == nil {
		return 0, fmt.Errorf("unexpected WCA Live redirect for %q", competitionWCAID)
	}

	return strconv.Atoi(match[1])
}

const competitionQuery = `query Competition($id: ID!) {
  competition(id: $id) {
    id
    wcaId
    name
    startDate
    endDate
    competitionEvents {
      event { id }
      rounds {
        id
        number
        name
        open
        finished
        active
        format { id numberOfAttempts }
        timeLimit { centiseconds cumulativeRoundWcifIds }
        cutoff { attemptResult numberOfAttempts }
      }
    }
  }
}`

// GetCompetition a verseny köreit kéri le (eredmények nélkül).
func (c *LiveClient) GetCompetition(ctx context.Context, liveCompetitionID int) (*liveCompetition, error) {
	var data struct {
		Competition *liveCompetition `json:"competition"`
	}

	err := c.query(ctx, competitionQuery, map[string]any{
		"id": strconv.Itoa(liveCompetitionID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Competition == nil {
		return nil, fmt.Errorf("competition %d not found on WCA Live", liveCompetitionID)
	}

	return data.Competition, nil
}

// RoundQuery a kör teljes pillanatképe. Szándékosan ugyanazokat a mezőket kéri,
// mint a RoundSubscription, hogy az induló állapot és az élő frissítés azonos
// alakú legyen és ugyanaz a feldolgozás (snapshotResults) kezelhesse mindkettőt.
const RoundQuery = `query Round($id: ID!) {
  round(id: $id) {
    id
    number
    name
    open
    finished
    active
    format { id numberOfAttempts }
    timeLimit { centiseconds cumulativeRoundWcifIds }
    cutoff { attemptResult numberOfAttempts }
    competitionEvent { event { id } }
    results {
      id
      enteredAt
      best
      average
      attempts { result }
      person { id registrantId wcaId name }
    }
  }
}`

// RoundSubscription a WCA Live EGYETLEN subscription-je (scoretaking_subscription_types.ex).
// Minden esemény TELJES kör-pillanatképet küld, nem deltát.
const RoundSubscription = `subscription RoundUpdated($id: ID!) {
  roundUpdated(id: $id) {
    id
    results {
      id
      enteredAt
      best
      average
      attempts { result }
      person { id registrantId wcaId name }
    }
  }
}`

// GetRound egy kör aktuális állapotát kéri le eredményekkel együtt. Ezt
// használjuk induláskor és minden újracsatlakozás után, hogy a subscription
// alatt kimaradt eseményeket behozzuk.
func (c *LiveClient) GetRound(ctx context.Context, liveRoundID int) (*liveRound, error) {
	var data struct {
		Round *liveRound `json:"round"`
	}

	err := c.query(ctx, RoundQuery, map[string]any{
		"id": strconv.Itoa(liveRoundID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Round == nil {
		return nil, fmt.Errorf("round %d not found on WCA Live", liveRoundID)
	}

	return data.Round, nil
}

// competitionsQuery a verseny-választó listája.
//
// A WCA Live `competitions` lekérdezése névre szűr (filter) és a megadott
// naptól kezdve sorolja fel a versenyeket (from). Eredményeket NEM kérünk, így
// a komplexitási korlát közelébe sem ér: mért adat szerint ~185 verseny és a
// helyszínek országa együtt is ~27 KB.
const competitionsQuery = `query Competitions($filter: String, $from: Date, $limit: Int) {
  competitions(filter: $filter, from: $from, limit: $limit) {
    id
    wcaId
    name
    startDate
    endDate
    venues { country { iso2 name } }
  }
}`

// ListCompetitions a verseny-választó forrása. filter üresen hagyható (ekkor
// a from utáni összes versenyt adja), a from nulla értéke pedig azt jelenti:
// nincs dátumszűrés.
func (c *LiveClient) ListCompetitions(ctx context.Context, filter string, from time.Time, limit int) ([]Competition, error) {
	variables := map[string]any{"limit": limit}
	if filter != "" {
		variables["filter"] = filter
	}
	if !from.IsZero() {
		variables["from"] = from.Format(dateLayout)
	}

	var data struct {
		Competitions []liveCompetition `json:"competitions"`
	}
	if err := c.query(ctx, competitionsQuery, variables, &data); err != nil {
		return nil, err
	}

	competitions := make([]Competition, 0, len(data.Competitions))
	for _, live := range data.Competitions {
		competition, err := toCompetition(live)
		if err != nil {
			// Egy értelmezhetetlen sor ne vigye el az egész listát.
			continue
		}
		competitions = append(competitions, competition)
	}

	return competitions, nil
}

// competitionInfoQuery a verseny alapadatai, körök nélkül — a kliens által
// megadott azonosító feloldásához pontosan ennyi kell.
const competitionInfoQuery = `query CompetitionInfo($id: ID!) {
  competition(id: $id) {
    id
    wcaId
    name
    startDate
    endDate
    venues { country { iso2 name } }
  }
}`

func (c *LiveClient) GetCompetitionInfo(ctx context.Context, liveCompetitionID int) (*Competition, error) {
	var data struct {
		Competition *liveCompetition `json:"competition"`
	}

	err := c.query(ctx, competitionInfoQuery, map[string]any{
		"id": strconv.Itoa(liveCompetitionID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Competition == nil {
		return nil, fmt.Errorf("competition %d not found on WCA Live", liveCompetitionID)
	}

	competition, err := toCompetition(*data.Competition)
	if err != nil {
		return nil, err
	}

	return &competition, nil
}

// dateLayout a WCA Live Date skalárjának alakja.
const dateLayout = "2006-01-02"

// toCompetition a WCA Live DTO-jából a saját típusunkat képzi. Az országot az
// első helyszínből vesszük: több helyszín esetén is ugyanaz az ország.
func toCompetition(live liveCompetition) (Competition, error) {
	liveID, err := strconv.Atoi(live.ID)
	if err != nil {
		return Competition{}, fmt.Errorf("unexpected WCA Live competition id %q", live.ID)
	}

	competition := Competition{
		CompetitionWCAID:  live.WCAID,
		LiveCompetitionID: liveID,
		Name:              live.Name,
		StartDate:         live.StartDate,
		EndDate:           live.EndDate,
	}
	if len(live.Venues) > 0 {
		competition.CountryISO2 = live.Venues[0].Country.ISO2
	}

	return competition, nil
}
