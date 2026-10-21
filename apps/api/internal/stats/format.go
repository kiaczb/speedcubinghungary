package stats

import "fmt"

const (
	resultDNF = -1
	resultDNS = -2
)

func FormatResult(eventID string, value int, average bool) string {
	switch {
	case value == resultDNF:
		return "DNF"
	case value == resultDNS:
		return "DNS"
	case value <= 0:
		return ""
	}

	switch eventID {
	case "333fm":
		if average {
			return fmt.Sprintf("%.2f", float64(value)/100)
		}
		return fmt.Sprintf("%d", value)
	case "333mbf":
		return formatMultiBlind(value)
	case "333mbo":
		if value >= 1000000000 {
			return formatOldMultiBlind(value)
		}
		return formatMultiBlind(value)
	}
	return formatCentiseconds(value)
}

func formatCentiseconds(value int) string {
	hours := value / 360000
	minutes := value % 360000 / 6000
	seconds := value % 6000 / 100
	hundredths := value % 100

	switch {
	case hours > 0:
		return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, hundredths)
	case minutes > 0:
		return fmt.Sprintf("%d:%02d.%02d", minutes, seconds, hundredths)
	}
	return fmt.Sprintf("%d.%02d", seconds, hundredths)
}

func formatMultiBlind(value int) string {
	missed := value % 100
	seconds := value / 100 % 100000
	points := 99 - value/10000000
	solved := points + missed
	attempted := solved + missed

	if seconds == 99999 {
		return fmt.Sprintf("%d/%d", solved, attempted)
	}
	return fmt.Sprintf("%d/%d %d:%02d", solved, attempted, seconds/60, seconds%60)
}

func formatOldMultiBlind(value int) string {
	seconds := value % 100000
	attempted := value / 100000 % 100
	solved := 99 - value/10000000%100

	if seconds == 99999 {
		return fmt.Sprintf("%d/%d", solved, attempted)
	}
	return fmt.Sprintf("%d/%d %d:%02d", solved, attempted, seconds/60, seconds%60)
}

func isTimedEvent(eventID string) bool {
	return eventID != "333fm" && eventID != "333mbf" && eventID != "333mbo"
}
