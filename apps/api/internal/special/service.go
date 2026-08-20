package special

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrInvalidCompetitionID = errors.New("invalid competition id")
	ErrInvalidElapsed       = errors.New("elapsed time must be between 1 and 8640000 centiseconds")
	ErrInvalidAttempt       = errors.New("invalid attempt reference")
)

// maxElapsedCs 24 óra századmásodpercben — józan felső határ az emberi bevitelre.
const maxElapsedCs = 24 * 60 * 60 * 100

// competitionIDPattern a WCA verseny-azonosító alakja (csak betű és szám).
// Azért ellenőrizzük, mert az azonosító a WCA Live /link/ URL-jébe kerül, és
// kérésenként a kliens adja — hibás alakút meg se próbálunk feloldani.
var competitionIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

const (
	// A verseny-választó alapból a magyar versenyeket kínálja; kereséssel
	// bármelyik WCA Live-on futó verseny elérhető.
	defaultCountryISO2 = "HU"
	// Ennyi napra visszamenőleg soroljuk fel a versenyeket: egy Special
	// eredménytábláját a verseny napja után is nézik.
	competitionLookbackDays = 60
	// A WCA Live a from utáni ÖSSZES versenyt adja (világszerte, mért adat
	// szerint ~185 db ~27 KB-ban), az ország szerinti szűrés itt történik.
	competitionListLimit   = 500
	competitionSearchLimit = 25

	// A verseny alapadatai ritkán változnak, a feloldás viszont két WCA Live
	// kérés (átirányítás + lekérdezés) — ezért gyorsítótárazzuk.
	competitionCacheTTL = time.Hour
	// A lista rövid ideig él: verseny közben is nyílhat új WCA Live verseny.
	competitionListCacheTTL = time.Minute
)

// Service a riport összeállítása és az emberi időrögzítés.
//
// A versenyt kérésenként a NÉZŐ adja meg — nincs szerveroldali "figyelt
// verseny". A kapott azonosítót mindig a WCA Live-nál oldjuk fel, hogy a
// numerikus id ne a kliens szava legyen.
type Service struct {
	repo        Repository
	client      *LiveClient
	tracker     *Tracker
	broadcaster *Broadcaster

	// A feloldott versenyek és a választó listája gyorsítótárban; a WCA Live-ot
	// önkéntesek üzemeltetik, ne kérdezzük fölöslegesen.
	mu       sync.Mutex
	resolved map[string]cachedCompetition
	list     []Competition
	listedAt time.Time
}

type cachedCompetition struct {
	competition Competition
	at          time.Time
}

func NewService(repo Repository, client *LiveClient, tracker *Tracker, broadcaster *Broadcaster) *Service {
	return &Service{
		repo:        repo,
		client:      client,
		tracker:     tracker,
		broadcaster: broadcaster,
		resolved:    make(map[string]cachedCompetition),
	}
}

// Broadcaster a nyitott felületek élő értesítéséhez (SSE).
func (s *Service) Broadcaster() *Broadcaster { return s.broadcaster }

// Tracker a WCA Live oldali állapot; az SSE kezelő ezen keresztül jelzi, hogy
// van nyitott nézet.
func (s *Service) Tracker() *Tracker { return s.tracker }

// Competitions a verseny-választó találatai.
//
// Üres kereséssel a magyar versenyeket adja vissza (az elmúlt
// competitionLookbackDays nap + a jövőbeliek), keresésre pedig a WCA Live
// névre szűrő lekérdezését használjuk, ország szerinti szűkítés nélkül — így
// egy külföldi verseny is megnyitható, ha valaki keresi.
func (s *Service) Competitions(ctx context.Context, query string) ([]Competition, error) {
	query = strings.TrimSpace(query)

	if query != "" {
		found, err := s.client.ListCompetitions(ctx, query, time.Time{}, competitionSearchLimit)
		if err != nil {
			return nil, err
		}
		sortCompetitions(found, today())
		return found, nil
	}

	if cached, ok := s.cachedList(); ok {
		return cached, nil
	}

	from := time.Now().AddDate(0, 0, -competitionLookbackDays)
	all, err := s.client.ListCompetitions(ctx, "", from, competitionListLimit)
	if err != nil {
		return nil, err
	}

	local := make([]Competition, 0, 8)
	for _, competition := range all {
		if competition.CountryISO2 == defaultCountryISO2 {
			local = append(local, competition)
		}
	}
	sortCompetitions(local, today())

	s.mu.Lock()
	s.list = local
	s.listedAt = time.Now()
	s.mu.Unlock()

	return local, nil
}

func (s *Service) cachedList() ([]Competition, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.list == nil || time.Since(s.listedAt) > competitionListCacheTTL {
		return nil, false
	}
	return s.list, true
}

// ResolveCompetition a WCA verseny-azonosítóból a verseny alapadatait adja.
//
// A numerikus WCA Live id-t SOHA nem a klienstől vesszük: a /link/ átirányítás
// és a lekérdezés a WCA Live-tól jön, így elgépelt vagy hamis azonosítóval nem
// kérdezünk le semmit.
func (s *Service) ResolveCompetition(ctx context.Context, competitionWCAID string) (Competition, error) {
	competitionWCAID = strings.TrimSpace(competitionWCAID)
	if !competitionIDPattern.MatchString(competitionWCAID) {
		return Competition{}, ErrInvalidCompetitionID
	}

	s.mu.Lock()
	cached, known := s.resolved[competitionWCAID]
	s.mu.Unlock()

	if known && time.Since(cached.at) < competitionCacheTTL {
		return cached.competition, nil
	}

	liveID, err := s.client.ResolveCompetitionID(ctx, competitionWCAID)
	if err != nil {
		return Competition{}, err
	}

	competition, err := s.client.GetCompetitionInfo(ctx, liveID)
	if err != nil {
		return Competition{}, err
	}

	s.mu.Lock()
	s.resolved[competitionWCAID] = cachedCompetition{competition: *competition, at: time.Now()}
	s.mu.Unlock()

	return *competition, nil
}

// Report a versenyzőnkénti nézet: minden idő, felhasznált/maradék keret.
//
// A kumulatív keret számítása KIZÁRÓLAG a Competitor.Budgets mezőben él
// (buildBudgets) — a Szabályzat A1a3 értelmezését szándékosan egyetlen helyen
// tartjuk, hogy ne csúszhasson szét két implementáció.
type Report struct {
	// Competition a riport tárgya: a néző által választott verseny.
	Competition Competition `json:"competition"`
	// AvailableRounds a versenyen előforduló körsorszámok (növekvő sorrendben),
	// MINDIG a teljes adathalmazból — a kör-választó ebből épül fel, ezért nem
	// szűkülhet a kiválasztott körre.
	AvailableRounds []int `json:"availableRounds"`
	// Round a szűrés tárgya; nil, ha minden kör benne van.
	Round       *int         `json:"round"`
	Competitors []Competitor `json:"competitors"`
}

// GetReport a megadott verseny riportját adja vissza.
//
// Az eredmények a WCA Live-ból jönnek (tracker), a DNF-ek tényleges hossza az
// adatbázisból; a kettő itt találkozik.
//
// roundNumber != nil esetén CSAK az adott sorszámú körökre szűkít, minden
// versenyszámban (pl. 1 → 333-r1, 222-r1, …). A szűrés a BuildCompetitors ELŐTT
// történik, hogy a felhasznált/maradék idő és a kumulatív keret ugyanazzal az
// egyetlen implementációval (buildBudgets) számolódjon — kliensoldali
// újraszámolás nélkül.
func (s *Service) GetReport(ctx context.Context, competitionWCAID string, roundNumber *int) (*Report, error) {
	competition, err := s.ResolveCompetition(ctx, competitionWCAID)
	if err != nil {
		return nil, err
	}

	results, err := s.tracker.Results(ctx, competition)
	if err != nil {
		return nil, err
	}

	recorded, err := s.repo.ListRecordedTimes(competition.CompetitionWCAID)
	if err != nil {
		return nil, err
	}

	// A választható körök a SZŰRÉS ELŐTTI halmazból jönnek.
	available := availableRoundNumbers(results)

	if roundNumber != nil {
		results = filterResultsByRound(results, *roundNumber)
	}

	return &Report{
		Competition:     competition,
		AvailableRounds: available,
		Round:           roundNumber,
		Competitors:     BuildCompetitors(results, recorded),
	}, nil
}

func availableRoundNumbers(results []ResultSnapshot) []int {
	seen := make(map[int]bool)
	for _, result := range results {
		seen[result.RoundNumber] = true
	}

	numbers := make([]int, 0, len(seen))
	for number := range seen {
		numbers = append(numbers, number)
	}
	sort.Ints(numbers)

	return numbers
}

func filterResultsByRound(results []ResultSnapshot, roundNumber int) []ResultSnapshot {
	filtered := make([]ResultSnapshot, 0, len(results))
	for _, result := range results {
		if result.RoundNumber == roundNumber {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

// sortCompetitions a választó sorrendje: előbb a most futó versenyek, utána a
// közelgők (dátum szerint növekvő), végül a lezártak (a legfrissebb elöl).
func sortCompetitions(competitions []Competition, today string) {
	sort.SliceStable(competitions, func(i, j int) bool {
		left, right := competitions[i], competitions[j]

		leftRank, rightRank := competitionRank(left, today), competitionRank(right, today)
		if leftRank != rightRank {
			return leftRank < rightRank
		}

		// A lezártak közül a legfrissebb az érdekes, a többinél a soron
		// következő. Az ISO dátumok sorrendje megegyezik a szövegessel.
		if leftRank == rankPast {
			return left.StartDate > right.StartDate
		}
		return left.StartDate < right.StartDate
	})
}

const (
	rankOngoing = iota
	rankUpcoming
	rankPast
)

func competitionRank(competition Competition, today string) int {
	switch {
	case competition.StartDate <= today && today <= competition.EndDate:
		return rankOngoing
	case competition.StartDate > today:
		return rankUpcoming
	default:
		return rankPast
	}
}

func today() string {
	return time.Now().Format(dateLayout)
}

// RecordElapsed rögzíti egy DNF kísérlet TÉNYLEGES megoldási idejét.
//
// Ez emberi megfigyelés: a WCA Live nem tudja megmondani, meddig tartott a
// solve (az Attempt modellben nincs időbélyeg, a limitet átlépő időt pedig a
// mentés előtt DNF-re cseréli). nil elapsedCs törli a bejegyzést.
func (s *Service) RecordElapsed(
	ctx context.Context,
	competitionWCAID string,
	liveResultID, attemptNumber int,
	elapsedCs *int,
	note string,
	userID int,
) error {
	if liveResultID <= 0 || attemptNumber <= 0 {
		return ErrInvalidAttempt
	}

	competitionWCAID = strings.TrimSpace(competitionWCAID)
	if !competitionIDPattern.MatchString(competitionWCAID) {
		return ErrInvalidCompetitionID
	}

	if elapsedCs == nil {
		// Törléshez nem kell a WCA Live: a (live_result_id, attempt_number)
		// kulcs önmagában azonosítja a bejegyzést.
		if err := s.repo.DeleteElapsed(liveResultID, attemptNumber); err != nil {
			return err
		}
		s.broadcaster.Notify(competitionWCAID)
		return nil
	}

	if *elapsedCs <= 0 || *elapsedCs > maxElapsedCs {
		return ErrInvalidElapsed
	}

	// A verseny sora a rögzített idő idegen kulcsa — csak itt keletkezik, tehát
	// akkor, ha ember tényleg rögzített valamit.
	competition, err := s.ResolveCompetition(ctx, competitionWCAID)
	if err != nil {
		return err
	}
	if err := s.repo.EnsureCompetition(competition, userID); err != nil {
		return err
	}

	if err := s.repo.SetElapsed(
		competition.CompetitionWCAID, liveResultID, attemptNumber,
		*elapsedCs, strings.TrimSpace(note), userID, time.Now().UTC(),
	); err != nil {
		return err
	}

	s.broadcaster.Notify(competition.CompetitionWCAID)

	return nil
}
