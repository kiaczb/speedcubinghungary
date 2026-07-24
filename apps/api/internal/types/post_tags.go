package types

type PostTag string

const (
	TAG_EVENT        PostTag = "event"
	TAG_COMPETITION  PostTag = "competition"
	TAG_NEWS         PostTag = "news"
	TAG_ANNOUNCEMENT PostTag = "announcement"
	// TAG_REPORT: a weekly-report job által generált heti riport. A posztot a
	// Python job hozza létre közvetlenül (lásd apps/scripts/weekly-report),
	// de a tag itt is szerepel, hogy a validátor és az API elfogadja.
	TAG_REPORT PostTag = "report"
)