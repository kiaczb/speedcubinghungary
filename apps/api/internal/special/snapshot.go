package special

import (
	"fmt"
	"strconv"
)

// RoundIdentity azonosítja azt a kört, amelyhez egy pillanatkép tartozik.
// A roundUpdated subscription payloadja szándékosan szűk (id + results), az
// eventId/roundNumber a tracker nyilvántartásából jön — lásd tracker.go.
type RoundIdentity struct {
	LiveRoundID int
	EventID     string
	RoundNumber int
}

// RoundWCIFID a WCIF Round.id / ActivityCode formátuma: "{eventId}-r{number}".
func RoundWCIFID(eventID string, roundNumber int) string {
	return fmt.Sprintf("%s-r%d", eventID, roundNumber)
}

// snapshotResults egy kör WCA Live pillanatképét alakítja a saját nézetünk
// bemenetévé.
//
// A WCA Live minden eseménye TELJES kör-pillanatkép, nem delta — ezért a
// feldolgozás állapotalapú: ugyanarra a bemenetre mindig ugyanaz jön ki, és a
// duplikált vagy sorrenden kívüli események önmaguktól helyreállnak.
func snapshotResults(id RoundIdentity, round liveRound) []ResultSnapshot {
	roundWCIF := RoundWCIFID(id.EventID, id.RoundNumber)

	var limitCs *int
	cumulative := []string{}
	if round.TimeLimit != nil {
		cs := round.TimeLimit.Centiseconds
		limitCs = &cs
		if round.TimeLimit.CumulativeRoundWCIFIDs != nil {
			cumulative = round.TimeLimit.CumulativeRoundWCIFIDs
		}
	}

	snapshots := make([]ResultSnapshot, 0, len(round.Results))

	for _, result := range round.Results {
		if result.Person.RegistrantID == nil {
			continue
		}
		resultID, err := strconv.Atoi(result.ID)
		if err != nil {
			continue
		}
		// Üres eredmény (a versenyző még nem szerepelt): nem vesszük fel, csak
		// zajt vinne a nézetbe.
		if len(result.Attempts) == 0 {
			continue
		}

		attempts := make([]int, 0, len(result.Attempts))
		for _, attempt := range result.Attempts {
			attempts = append(attempts, attempt.Result)
		}

		snapshots = append(snapshots, ResultSnapshot{
			LiveResultID:       resultID,
			EventID:            id.EventID,
			RoundNumber:        id.RoundNumber,
			RoundWCIFID:        roundWCIF,
			RegistrantID:       *result.Person.RegistrantID,
			CompetitorWCAID:    result.Person.WCAID,
			CompetitorName:     result.Person.Name,
			Attempts:           attempts,
			Best:               result.Best,
			Average:            result.Average,
			TimeLimitCs:        limitCs,
			CumulativeRoundIDs: cumulative,
		})
	}

	return snapshots
}

// attemptBound felső KORLÁTOT ad a kísérlet eltelt idejére — sosem mérést.
//
// Szabályzat A1a: "The time limit for a particular attempt is the time limit or
// the cumulative time limit minus the total time in other attempts counting
// towards the time limit so far, whichever is lower." A bíró a limit elérésekor
// megállítja a solve-ot (A1a1), tehát az eltelt idő legfeljebb ennyi lehet.
// Büntetésből kapott DNF-nél az eltelt idő ennél bármennyivel kevesebb lehet,
// ezért ez kizárólag korlát.
//
// Nincs limit (333fm, 333mbf) → nincs korlát.
func attemptBound(limitCs *int, cumulativeRoundIDs []string, attempts []int, idx int) *int {
	if limitCs == nil {
		return nil
	}

	if len(cumulativeRoundIDs) == 0 {
		cs := *limitCs
		return &cs
	}

	// Kumulatív limit: a korábbi kísérletek elhasznált idejét levonjuk.
	// Csak a pozitív eredményeket ismerjük; a korábbi DNF-ek eltelt idejét a
	// WCA Live nem adja meg, azokat 0-nak vesszük — így a korlát biztosan nem
	// lesz kisebb a valóságosnál (konzervatív marad).
	//
	// Megjegyzés: a kör-közi (cross-round) kumulatív limitet a WCA Live maga is
	// körönként külön kezeli, ezt a viselkedést követjük.
	used := 0
	for j := 0; j < idx && j < len(attempts); j++ {
		if attempts[j] > 0 {
			used += attempts[j]
		}
	}

	remaining := *limitCs - used
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}
