package weeklyreport

import (
	"log"
	"sort"
	"sync"
)

type DryRunRepository struct {
	Repository

	mu             sync.Mutex
	counts         map[string]PersonCounts
	delegated      map[string]DelegatedCount
	storedRecords  RecordTable
	changedRecords int
}

func NewDryRunRepository(repo Repository) *DryRunRepository {
	return &DryRunRepository{
		Repository: repo,
		counts:     map[string]PersonCounts{},
		delegated:  map[string]DelegatedCount{},
	}
}

func (d *DryRunRepository) SaveCounts(counts []PersonCounts) error {
	stored, err := d.storedPeople()
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	changed := 0
	for _, c := range counts {
		d.counts[c.WCAID] = c
		old, ok := stored[c.WCAID]
		if ok && old.CompetitionCount == c.CompetitionCount && old.SolveCount == c.SolveCount {
			continue
		}
		changed++
		if !ok {
			log.Printf("heti riport [dry-run]: új versenyző lenne: %s %s (%d verseny, %d kirakás)", c.WCAID, c.Name, c.CompetitionCount, c.SolveCount)
			continue
		}
		log.Printf("heti riport [dry-run]: %s %s: verseny %d → %d, kirakás %d → %d", c.WCAID, c.Name, old.CompetitionCount, c.CompetitionCount, old.SolveCount, c.SolveCount)
	}
	log.Printf("heti riport [dry-run]: %d versenyző számlálója változna (nem íródik ki)", changed)
	return nil
}

func (d *DryRunRepository) SaveDelegatedCounts(counts []DelegatedCount) error {
	stored, err := d.storedPeople()
	if err != nil {
		return err
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	for _, c := range counts {
		d.delegated[c.WCAID] = c
		if old := stored[c.WCAID]; old.DelegatedCount != c.Count {
			log.Printf("heti riport [dry-run]: %s %s: delegált %d → %d", c.WCAID, c.Name, old.DelegatedCount, c.Count)
		}
	}
	return nil
}

func (d *DryRunRepository) UpsertRecord(region, eventID string, values RecordValues) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.storedRecords == nil {
		stored, err := d.Repository.GetRecords()
		if err != nil {
			return err
		}
		d.storedRecords = stored
	}
	if old, ok := d.storedRecords[region][eventID]; !ok || !sameValue(old.Single, values.Single) || !sameValue(old.Average, values.Average) {
		d.changedRecords++
		log.Printf("heti riport [dry-run]: rekord változna: %s %s", region, eventID)
	}
	return nil
}

func sameValue(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func (d *DryRunRepository) PublishReport(post ReportPost, replace bool) (bool, error) {
	log.Printf("heti riport [dry-run]: a poszt nem íródik ki: %q (%d bájt JSON, felülírás: %v)", post.Title, len(post.Content), replace)
	return true, nil
}

func (d *DryRunRepository) TrackedPeople() ([]Person, error) {
	people, err := d.people()
	if err != nil {
		return nil, err
	}
	out := make([]Person, 0, len(people))
	for _, p := range people {
		if p.CompetitionCount > 0 || p.DelegatedCount > 0 {
			out = append(out, p)
		}
	}
	return out, nil
}

func (d *DryRunRepository) CompetitorsNearMilestones(milestones []int, difference int) ([]Person, error) {
	return d.near(func(p Person) int { return p.CompetitionCount }, milestones, difference)
}

func (d *DryRunRepository) DelegatesNearMilestones(milestones []int, difference int) ([]Person, error) {
	return d.near(func(p Person) int { return p.DelegatedCount }, milestones, difference)
}

func (d *DryRunRepository) SolversNearMilestones(step, difference int) ([]Person, error) {
	people, err := d.people()
	if err != nil {
		return nil, err
	}
	out := make([]Person, 0)
	for _, p := range people {
		rem := p.SolveCount % step
		if p.SolveCount > 0 && ((rem <= difference && p.SolveCount >= step) || rem >= step-difference) {
			out = append(out, p)
		}
	}
	return out, nil
}

func (d *DryRunRepository) near(value func(Person) int, milestones []int, difference int) ([]Person, error) {
	people, err := d.people()
	if err != nil {
		return nil, err
	}
	out := make([]Person, 0)
	for _, p := range people {
		for _, m := range milestones {
			if v := value(p); v >= m-difference && v <= m {
				out = append(out, p)
				break
			}
		}
	}
	return out, nil
}

func (d *DryRunRepository) storedPeople() (map[string]Person, error) {
	tracked, err := d.Repository.TrackedPeople()
	if err != nil {
		return nil, err
	}
	stored := make(map[string]Person, len(tracked))
	for _, p := range tracked {
		stored[p.WCAID] = p
	}
	return stored, nil
}

func (d *DryRunRepository) people() ([]Person, error) {
	stored, err := d.storedPeople()
	if err != nil {
		return nil, err
	}
	d.mu.Lock()
	defer d.mu.Unlock()

	for id, c := range d.counts {
		p := stored[id]
		p.WCAID, p.Name = id, c.Name
		p.CompetitionCount, p.SolveCount = c.CompetitionCount, c.SolveCount
		stored[id] = p
	}
	for id, c := range d.delegated {
		p := stored[id]
		p.WCAID, p.Name = id, c.Name
		p.DelegatedCount = c.Count
		stored[id] = p
	}

	people := make([]Person, 0, len(stored))
	for _, p := range stored {
		people = append(people, p)
	}
	sort.Slice(people, func(i, j int) bool { return people[i].Name < people[j].Name })
	return people, nil
}

type subjectPrefixMailer struct {
	Mailer
	prefix string
}

func WithSubjectPrefix(m Mailer, prefix string) Mailer {
	return subjectPrefixMailer{Mailer: m, prefix: prefix}
}

func (m subjectPrefixMailer) Send(subject, htmlBody string) error {
	return m.Mailer.Send(m.prefix+subject, htmlBody)
}
