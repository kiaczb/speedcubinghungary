package types

type PredictionQuestionType string

const (
	QUESTION_PERSON PredictionQuestionType = "person"
	QUESTION_TIME   PredictionQuestionType = "time"
	QUESTION_NUMBER PredictionQuestionType = "number"
	QUESTION_TEXT   PredictionQuestionType = "text"
)
