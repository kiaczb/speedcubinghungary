package learn

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) ListEvents() ([]Event, error) {
	args := m.Called()
	return args.Get(0).([]Event), args.Error(1)
}

func (m *mockRepo) GetEvent(id string) (Event, error) {
	args := m.Called(id)
	return args.Get(0).(Event), args.Error(1)
}

func (m *mockRepo) CreateEvent(e Event) error { return m.Called(e).Error(0) }
func (m *mockRepo) UpdateEvent(e Event) error { return m.Called(e).Error(0) }
func (m *mockRepo) DeleteEvent(id string) error {
	return m.Called(id).Error(0)
}

func (m *mockRepo) ListCases(eventID string) ([]Case, error) {
	args := m.Called(eventID)
	return args.Get(0).([]Case), args.Error(1)
}

func (m *mockRepo) ListAllCases() ([]Case, error) {
	args := m.Called()
	return args.Get(0).([]Case), args.Error(1)
}

func (m *mockRepo) GetCase(id string) (Case, error) {
	args := m.Called(id)
	return args.Get(0).(Case), args.Error(1)
}

func (m *mockRepo) CreateCase(c Case) error { return m.Called(c).Error(0) }
func (m *mockRepo) UpdateCase(c Case) error { return m.Called(c).Error(0) }
func (m *mockRepo) DeleteCase(id string) error {
	return m.Called(id).Error(0)
}

func (m *mockRepo) ListAlgorithms(caseID string) ([]Algorithm, error) {
	args := m.Called(caseID)
	return args.Get(0).([]Algorithm), args.Error(1)
}

func (m *mockRepo) ListAllAlgorithms() ([]Algorithm, error) {
	args := m.Called()
	return args.Get(0).([]Algorithm), args.Error(1)
}

func (m *mockRepo) GetAlgorithm(id string) (Algorithm, error) {
	args := m.Called(id)
	return args.Get(0).(Algorithm), args.Error(1)
}

func (m *mockRepo) CreateAlgorithm(a Algorithm) error { return m.Called(a).Error(0) }
func (m *mockRepo) UpdateAlgorithm(a Algorithm) error { return m.Called(a).Error(0) }
func (m *mockRepo) DeleteAlgorithm(id string) error {
	return m.Called(id).Error(0)
}

func TestValidateAlgorithm(t *testing.T) {
	valid := Algorithm{
		ID:       uuid.NewString(),
		CaseID:   uuid.NewString(),
		Name:     "Sune",
		Notation: "R U R' U R U2 R'",
	}

	tests := []struct {
		name    string
		mutate  func(a *Algorithm)
		wantErr error
	}{
		{"érvényes", func(a *Algorithm) {}, nil},
		{"név nélkül", func(a *Algorithm) { a.Name = "   " }, ErrEmptyName},
		{"notáció nélkül", func(a *Algorithm) { a.Notation = "" }, ErrEmptyNotation},
		{"eset nélkül", func(a *Algorithm) { a.CaseID = "" }, ErrInvalidID},
		{
			"videó nélkül is jó",
			func(a *Algorithm) { a.VideoURL = "" },
			nil,
		},
		{
			"https videó",
			func(a *Algorithm) { a.VideoURL = "https://www.youtube.com/watch?v=abc" },
			nil,
		},
		// A videó-URL a kliensen iframe/anchor forrásába kerül — csak http(s)
		// séma engedhető, különben `javascript:` is bekerülhetne.
		{
			"javascript séma tiltott",
			func(a *Algorithm) { a.VideoURL = "javascript:alert(1)" },
			ErrInvalidVideoURL,
		},
		{
			"relatív URL tiltott",
			func(a *Algorithm) { a.VideoURL = "/videos/sune.mp4" },
			ErrInvalidVideoURL,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			a := valid
			tc.mutate(&a)

			err := ValidateAlgorithm(a)
			if tc.wantErr == nil {
				assert.NoError(t, err)
				return
			}
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

// A szülőt a service előre ellenőrzi, hogy hiányzó esetre 404 legyen a válasz,
// ne egy FK-hibából lett 500.
func TestService_CreateAlgorithm_MissingCase(t *testing.T) {
	repo := new(mockRepo)
	caseID := uuid.NewString()
	repo.On("GetCase", caseID).Return(Case{}, ErrCaseNotFound)

	_, err := NewService(repo).CreateAlgorithm(caseID, "Sune", "R U R' U R U2 R'", "", "", 0)

	assert.ErrorIs(t, err, ErrCaseNotFound)
	repo.AssertNotCalled(t, "CreateAlgorithm", mock.Anything)
}

// Formailag hibás id-re nem szabad lekérdezni: a Postgres UUID-oszlopon
// típushibát dobna, abból 500 lenne a 404 helyett.
func TestService_GetAlgorithmDetail_InvalidID(t *testing.T) {
	repo := new(mockRepo)

	_, err := NewService(repo).GetAlgorithmDetail("not-a-uuid")

	assert.ErrorIs(t, err, ErrAlgorithmNotFound)
	repo.AssertNotCalled(t, "GetAlgorithm", mock.Anything)
}

func TestService_GetAlgorithmDetail_IncludesParentsAndSiblings(t *testing.T) {
	repo := new(mockRepo)

	eventID, caseID, algorithmID := uuid.NewString(), uuid.NewString(), uuid.NewString()
	algorithm := Algorithm{ID: algorithmID, CaseID: caseID, Name: "Sune", Notation: "R U R' U R U2 R'"}
	siblings := []Algorithm{algorithm, {ID: uuid.NewString(), CaseID: caseID, Name: "Antisune"}}

	repo.On("GetAlgorithm", algorithmID).Return(algorithm, nil)
	repo.On("GetCase", caseID).Return(Case{ID: caseID, EventID: eventID, Name: "OLL 21"}, nil)
	repo.On("GetEvent", eventID).Return(Event{ID: eventID, Name: "3x3x3"}, nil)
	repo.On("ListAlgorithms", caseID).Return(siblings, nil)

	detail, err := NewService(repo).GetAlgorithmDetail(algorithmID)

	require.NoError(t, err)
	assert.Equal(t, "3x3x3", detail.Event.Name)
	assert.Equal(t, "OLL 21", detail.Case.Name)
	assert.Equal(t, algorithmID, detail.Algorithm.ID)
	assert.Len(t, detail.Siblings, 2)
}

// A fa három lekérdezésből épül fel, nem szintenként N+1-ből.
func TestService_GetTree_NestsChildren(t *testing.T) {
	repo := new(mockRepo)

	eventID, caseID := uuid.NewString(), uuid.NewString()
	emptyEventID := uuid.NewString()

	repo.On("ListEvents").Return([]Event{
		{ID: eventID, Name: "3x3x3"},
		{ID: emptyEventID, Name: "2x2x2"},
	}, nil)
	repo.On("ListAllCases").Return([]Case{{ID: caseID, EventID: eventID, Name: "OLL 21"}}, nil)
	repo.On("ListAllAlgorithms").Return([]Algorithm{{ID: uuid.NewString(), CaseID: caseID, Name: "Sune"}}, nil)

	tree, err := NewService(repo).GetTree()

	require.NoError(t, err)
	require.Len(t, tree, 2)
	require.Len(t, tree[0].Cases, 1)
	assert.Len(t, tree[0].Cases[0].Algorithms, 1)
	// Üres versenyszám sem tűnhet el, és a gyerekei nem lehetnek nil-ek: a
	// JSON-ban üres tömbként kell megjelenniük, hogy a kliensnek ne kelljen
	// null-t kezelnie.
	assert.NotNil(t, tree[1].Cases)
	assert.Empty(t, tree[1].Cases)
}
