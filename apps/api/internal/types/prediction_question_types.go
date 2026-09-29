package types

type PredictionQuestionType string

const (
	QUESTION_PERSON PredictionQuestionType = "person"
	QUESTION_TIME   PredictionQuestionType = "time"
	QUESTION_NUMBER PredictionQuestionType = "number"
	// QUESTION_TEXT: szabadszavas válasz olyan kérdésekhez, amik nem
	// versenyzőt, időt vagy számot várnak (pl. "Milyen új trükköt fogunk látni?").
	QUESTION_TEXT PredictionQuestionType = "text"
)
