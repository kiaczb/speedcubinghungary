package competition

// countryNames maps ISO 3166-1 alpha-2 codes to human-readable country names.
var countryNames = map[string]string{
	"HU": "Hungary",
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// GetUpcomingHungarian returns the simplified list of upcoming Hungarian
// competitions, already sorted by start date by the upstream API.
func (s *Service) GetUpcomingHungarian() ([]Competition, error) {
	raw, err := s.repo.GetUpcomingByCountry("HU")
	if err != nil {
		return nil, err
	}

	competitions := make([]Competition, 0, len(raw))
	for _, c := range raw {
		competitions = append(competitions, Competition{
			ID:        c.ID,
			Name:      c.Name,
			StartDate: c.StartDate,
			EndDate:   c.EndDate,
			City:      c.City,
			Country:   countryName(c.CountryISO2),
			URL:       c.URL,
		})
	}

	return competitions, nil
}

// countryName resolves an ISO2 code to a readable name, falling back to the
// code itself when it is not known.
func countryName(iso2 string) string {
	if name, ok := countryNames[iso2]; ok {
		return name
	}
	return iso2
}
