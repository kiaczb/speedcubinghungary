package learn

// A tanulófelület hierarchiája: Event → Case → Algorithm (+ opcionális videó).
// A DTO-k camelCase-t használnak, mint a `special`/`allrounder` webes válaszai.

type Event struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	// CaseCount csak listázásnál töltődik ki (COUNT a lekérdezésben), hogy a
	// kártyákon ne kelljen külön kérés az esetek számához.
	CaseCount int `json:"caseCount"`
}

type Case struct {
	ID          string `json:"id"`
	EventID     string `json:"eventId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       int    `json:"order"`
	// AlgorithmCount csak listázásnál töltődik ki, lásd Event.CaseCount.
	AlgorithmCount int `json:"algorithmCount"`
}

type Algorithm struct {
	ID          string `json:"id"`
	CaseID      string `json:"caseId"`
	Name        string `json:"name"`
	Notation    string `json:"notation"`
	Description string `json:"description"`
	// VideoURL üres string, ha nincs videó — a DB oszlop nullable.
	VideoURL string `json:"videoUrl"`
	Order    int    `json:"order"`
}

// EventDetail: egy versenyszám a hozzá tartozó esetekkel (esetek oldala).
type EventDetail struct {
	Event Event  `json:"event"`
	Cases []Case `json:"cases"`
}

// CaseDetail: egy eset az algoritmusaival és a szülő versenyszámmal — a
// morzsamenühöz a szülő is kell, hogy az oldal egyetlen kérésből felépüljön.
type CaseDetail struct {
	Event      Event       `json:"event"`
	Case       Case        `json:"case"`
	Algorithms []Algorithm `json:"algorithms"`
}

// AlgorithmDetail: a tanulónézet payloadja. A `Siblings` az eset ÖSSZES
// algoritmusa (a kiválasztottat is beleértve), hogy a nézetben lehessen
// előre/hátra lépni és váltani anélkül, hogy vissza kellene navigálni.
type AlgorithmDetail struct {
	Event     Event       `json:"event"`
	Case      Case        `json:"case"`
	Algorithm Algorithm   `json:"algorithm"`
	Siblings  []Algorithm `json:"siblings"`
}

// A teljes fa az admin felületnek: egy kérés, nem szintenként N+1.
type CaseNode struct {
	Case
	Algorithms []Algorithm `json:"algorithms"`
}

type EventNode struct {
	Event
	Cases []CaseNode `json:"cases"`
}
