// Package special a "Budapest Special" formátumú versenyek nyilvános
// eredménytáblája, ahol a versenyző pontszáma az ÖSSZES versenyszámban elért
// ideje összege.
//
// Fontos, hogy tisztán lássuk, mit ad és mit nem ad a WCA Live:
//
//   - Ad: minden hivatalos eredményt, és azt is, hogy melyik versenyző melyik
//     körének hányadik kísérlete lett DNF, néhány másodperccel az adatrögzítés
//     után (roundUpdated subscription).
//   - NEM ad: a DNF solve tényleges hosszát. A WCA Live Attempt modellje
//     mindössze {result, reconstruction} (lásd lib/wca_live/scoretaking/attempt.ex),
//     és a Result.changeset a limitet átlépő időt még mentés ELŐTT DNF-re
//     cseréli, tehát az eredeti idő sehol nem marad meg. Egyetlen időbélyeg a
//     Result.enteredAt, ami az adatrögzítés (nem a solve) másodperc pontos ideje.
//
// Ebből következik a felelősségek szétválasztása, és az egész csomag alakja:
//
//   - A hivatalos eredmény a WCA Live-é. NEM tároljuk — kérésre kérdezzük le,
//     és amíg valaki nézi az oldalt, a subscription tartja frissen (tracker.go).
//   - A DNF tényleges hossza a MI adatunk, mert sehol máshol nem létezik. Ezt
//     ember rögzíti, és EZ AZ EGYETLEN, amit adatbázisba írunk (RecordedTime).
//
// A WCA Szabályzat A1a3 szerint kumulatív limitnél épp ez az eltelt idő számít
// bele a limitbe — ezt a hivatalos eredmény szándékosan nem tárolja.
package special

import "time"

// WCIF AttemptResult speciális értékei (WCIF specification#attemptresult, és
// wca-live lib/wca_live/scoretaking/attempt_result.ex).
const (
	AttemptSkipped = 0  // még nem történt meg, vagy cutoff miatt kimarad
	AttemptDNF     = -1 // Did Not Finish
	AttemptDNS     = -2 // Did Not Start
)

// Competition egy WCA Live-on futó verseny alapadatai.
//
// A versenyt a NÉZŐ választja ki (nincs szerveroldali "figyelt verseny"):
// minden kérés magával hozza a WCA verseny-azonosítót, a numerikus WCA Live
// id-t pedig mindig mi oldjuk fel — így elgépelt vagy hamis azonosítóval nem
// kérdezünk le semmit.
//
// Ugyanez a típus szolgálja a verseny-választót és a riport fejlécét is.
type Competition struct {
	CompetitionWCAID  string `json:"competitionWcaId"`
	LiveCompetitionID int    `json:"liveCompetitionId"`
	Name              string `json:"name"`
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	// CountryISO2 a helyszín országa (a WCA Live venues mezőjéből), ha ismert.
	CountryISO2 string `json:"countryIso2,omitempty"`
}

// RecordedTime egy DNF kísérlet TÉNYLEGES hossza — emberi megfigyelés, és az
// egyetlen dolog, amit ebből a csomagból tárolunk.
//
// A kulcs a WCA Live result id-ja (verseny + kör + versenyző, globálisan
// egyedi) és a kísérlet 1-alapú sorszáma. Minden más — versenyző, versenyszám,
// limit, korlát — a friss WCA Live pillanatképből jön, ezért nincs eltárolva.
type RecordedTime struct {
	LiveResultID  int
	AttemptNumber int
	ElapsedCs     int
	Note          string
	RecordedBy    *int
	RecordedAt    time.Time
}

// ResultSnapshot egy versenyző eredménye egy körben, a WCA Live szerint.
// Kizárólag a memóriában él (lásd tracker.go) — a hivatalos adatot nem tároljuk.
type ResultSnapshot struct {
	LiveResultID int

	EventID     string
	RoundNumber int
	RoundWCIFID string

	RegistrantID    int
	CompetitorWCAID string
	CompetitorName  string

	// Attempts WCIF AttemptResult értékek sorrendben (>0 cs, -1 DNF, -2 DNS, 0 kihagyott).
	Attempts []int
	Best     *int
	Average  *int

	TimeLimitCs        *int
	CumulativeRoundIDs []string
}

// --- Versenyzői nézet ---

// CompetitorAttempt egy kísérlet a versenyzői nézetben.
type CompetitorAttempt struct {
	AttemptNumber int `json:"attemptNumber"`
	// Result a hivatalos WCIF AttemptResult (-1 DNF, -2 DNS, 0 kihagyott, >0 cs).
	Result int `json:"result"`
	// ElapsedCs a rögzített TÉNYLEGES idő (emberi megfigyelés), ha van.
	ElapsedCs *int `json:"elapsedCs,omitempty"`
	// MaxElapsedCs felső korlát, nem mérés — lásd attemptBound.
	MaxElapsedCs *int `json:"maxElapsedCs,omitempty"`
}

// CompetitorRound egy versenyző egy köre. A LiveResultID az írás kulcsa: az
// admin ezen és a kísérlet sorszámán keresztül rögzít időt.
type CompetitorRound struct {
	LiveResultID       int                 `json:"liveResultId"`
	RoundWCIFID        string              `json:"roundWcifId"`
	EventID            string              `json:"eventId"`
	RoundNumber        int                 `json:"roundNumber"`
	Attempts           []CompetitorAttempt `json:"attempts"`
	Best               *int                `json:"best,omitempty"`
	Average            *int                `json:"average,omitempty"`
	TimeLimitCs        *int                `json:"timeLimitCs,omitempty"`
	CumulativeRoundIDs []string            `json:"cumulativeRoundIds"`
	// CompletedCs a sikeres kísérletek összege ebben a körben.
	CompletedCs int `json:"completedCs"`
	// RecordedDNFCs a rögzített DNF-idők összege ebben a körben.
	RecordedDNFCs int `json:"recordedDnfCs"`
	// MissingDNFCount a még rögzítetlen DNF-ek száma ebben a körben.
	MissingDNFCount int `json:"missingDnfCount"`
}

// CumulativeBudget egy kumulatív limit-keret elhasználása. A keret több körre is
// szólhat (WCIF TimeLimit.cumulativeRoundIds) — pl. a "Budapest Special"
// formátumnál egyetlen keret fedi az összes versenyszámot.
type CumulativeBudget struct {
	// RoundWCIFIDs a keretbe tartozó körök (rendezve).
	RoundWCIFIDs []string `json:"roundWcifIds"`
	LimitCs      int      `json:"limitCs"`
	CompletedCs  int      `json:"completedCs"`
	// RecordedDNFCs a rögzített DNF-idők összege — Szabályzat A1a3 szerint
	// ezek is beleszámítanak a keretbe.
	RecordedDNFCs int `json:"recordedDnfCs"`
	UsedCs        int `json:"usedCs"`
	RemainingCs   int `json:"remainingCs"`
	// MissingDNFCount amíg > 0, a RemainingCs csak FELSŐ becslés.
	MissingDNFCount int `json:"missingDnfCount"`
}

// Competitor a versenyzői nézet egy kártyája: minden köre, összesítve.
type Competitor struct {
	RegistrantID    int                `json:"registrantId"`
	CompetitorWCAID string             `json:"competitorWcaId,omitempty"`
	Name            string             `json:"name"`
	Rounds          []CompetitorRound  `json:"rounds"`
	Budgets         []CumulativeBudget `json:"budgets"`
	// UsedCs az összes körben elhasznált idő (sikeres kísérletek + rögzített
	// DNF-idők), kumulatív limittől függetlenül.
	UsedCs int `json:"usedCs"`
	// DNFCount / MissingDNFCount a versenyző összes köréből.
	DNFCount        int `json:"dnfCount"`
	MissingDNFCount int `json:"missingDnfCount"`
}

// --- WCA Live GraphQL DTO-k (a válasz camelCase-ben érkezik) ---

type liveCompetition struct {
	ID                string                `json:"id"`
	WCAID             string                `json:"wcaId"`
	Name              string                `json:"name"`
	StartDate         string                `json:"startDate"`
	EndDate           string                `json:"endDate"`
	Venues            []liveVenue           `json:"venues"`
	CompetitionEvents []liveCompetitionEvnt `json:"competitionEvents"`
}

type liveCompetitionEvnt struct {
	Event  liveEvent   `json:"event"`
	Rounds []liveRound `json:"rounds"`
}

type liveEvent struct {
	ID string `json:"id"`
}

type liveRound struct {
	ID        string         `json:"id"`
	Number    int            `json:"number"`
	Name      string         `json:"name"`
	Open      bool           `json:"open"`
	Finished  bool           `json:"finished"`
	Active    bool           `json:"active"`
	Format    *liveFormat    `json:"format"`
	TimeLimit *liveTimeLimit `json:"timeLimit"`
	Cutoff    *liveCutoff    `json:"cutoff"`
	Results   []liveResult   `json:"results"`

	// CompetitionEvent csak a round(id:) lekérdezésben jön vissza — a
	// subscription payloadja szándékosan szűkebb, ott a tracker a saját
	// nyilvántartásából tölti ki az event/round azonosítót.
	CompetitionEvent *liveCompEventRef `json:"competitionEvent"`
}

type liveCompEventRef struct {
	Event liveEvent `json:"event"`
}

type liveFormat struct {
	ID               string `json:"id"`
	NumberOfAttempts int    `json:"numberOfAttempts"`
}

type liveTimeLimit struct {
	Centiseconds           int      `json:"centiseconds"`
	CumulativeRoundWCIFIDs []string `json:"cumulativeRoundWcifIds"`
}

type liveCutoff struct {
	AttemptResult    int `json:"attemptResult"`
	NumberOfAttempts int `json:"numberOfAttempts"`
}

type liveResult struct {
	ID        string        `json:"id"`
	EnteredAt *string       `json:"enteredAt"`
	Best      *int          `json:"best"`
	Average   *int          `json:"average"`
	Attempts  []liveAttempt `json:"attempts"`
	Person    livePerson    `json:"person"`
}

type liveAttempt struct {
	Result int `json:"result"`
}

type livePerson struct {
	ID           string `json:"id"`
	RegistrantID *int   `json:"registrantId"`
	WCAID        string `json:"wcaId"`
	Name         string `json:"name"`
}

// liveVenue / liveCountry csak a verseny-választóhoz kell: a versenyt az
// ország szerint szűrjük (a lista alapból a magyar versenyeket kínálja).
type liveVenue struct {
	Country liveCountry `json:"country"`
}

type liveCountry struct {
	ISO2 string `json:"iso2"`
	Name string `json:"name"`
}
