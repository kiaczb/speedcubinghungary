package weeklyreport

import (
	"bytes"
	"embed"
	"encoding/base64"
	"fmt"
	"html/template"
	"mime"
	"net/smtp"
	"strings"
	"time"
)

//go:embed templates/email.html
var templateFS embed.FS

var emailTemplate = template.Must(template.ParseFS(templateFS, "templates/email.html"))

const (
	smtpHost   = "smtp.gmail.com"
	smtpPort   = "587"
	senderName = "Bonsz"
)

type emailView struct {
	Competitions         []emailCompetition
	CompetitorMilestones []CompetitorMilestone
	DelegateMilestones   []DelegateMilestone
	SolveMilestones      []SolveMilestone
	Summary              string
}

type emailCompetition struct {
	Name        string
	Flag        string
	Date        string
	Competitors []emailCompetitor
}

type emailCompetitor struct {
	Name   string
	Events []emailEvent
}

type emailEvent struct {
	ID      string
	Records []emailRecord
}

type emailRecord struct {
	Type       string
	Result     string
	Badge      string
	BadgeClass string
}

func EmailCompetitions(competitions []Competition) []Competition {
	selected := make([]Competition, 0, len(competitions))
	for _, comp := range competitions {
		if !comp.Hungarian {
			selected = append(selected, comp)
			continue
		}

		recorders := make([]Competitor, 0)
		for _, c := range comp.Competitors {
			if c.HasRecords() {
				recorders = append(recorders, c)
			}
		}
		if len(recorders) > 0 {
			comp.Competitors = recorders
			selected = append(selected, comp)
		}
	}
	return selected
}

func RenderEmail(data ReportData) (string, error) {
	competitions := EmailCompetitions(data.Competitions)

	view := emailView{
		Competitions:         make([]emailCompetition, 0, len(competitions)),
		CompetitorMilestones: data.CompetitorMilestones,
		DelegateMilestones:   data.DelegateMilestones,
		SolveMilestones:      data.SolveMilestones,
		Summary:              EmailSummary(competitions),
	}
	for _, comp := range competitions {
		ec := emailCompetition{
			Name: comp.Name,
			Flag: FlagEmoji(comp.CountryISO2),
			Date: FormatDateRange(comp.From, comp.To),
		}
		for _, c := range comp.Competitors {
			competitor := emailCompetitor{Name: c.Name}
			for _, event := range c.Events {
				ee := emailEvent{ID: event.EventID}
				for _, r := range event.Records {
					ee.Records = append(ee.Records, emailRecord{
						Type:       r.Type,
						Result:     formatRecord(event.EventID, r),
						Badge:      r.Badge,
						BadgeClass: "badge-" + strings.ToLower(r.Badge),
					})
				}
				competitor.Events = append(competitor.Events, ee)
			}
			ec.Competitors = append(ec.Competitors, competitor)
		}
		view.Competitions = append(view.Competitions, ec)
	}

	var buf bytes.Buffer
	if err := emailTemplate.Execute(&buf, view); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func EmailSummary(competitions []Competition) string {
	if len(competitions) == 0 {
		return "Az elmúlt hétvégén nem volt magyar érdekeltségű versenyeredmény."
	}

	sentences := make([]string, 0, len(competitions))
	for _, comp := range competitions {
		names := make([]string, 0, len(comp.Competitors))
		for _, c := range comp.Competitors {
			names = append(names, c.Name)
		}
		if len(names) > 1 {
			sentences = append(sentences, fmt.Sprintf("%s és %s a %s versenyen vettek részt",
				strings.Join(names[:len(names)-1], ", "), names[len(names)-1], comp.Name))
		} else {
			sentences = append(sentences, fmt.Sprintf("%s a %s versenyen vett részt", names[0], comp.Name))
		}
	}
	return "Az elmúlt hétvégén " + strings.Join(sentences, ",\n") + ".\n\nMindenkinek gratulálunk!"
}

func EmailSubject(now time.Time) string {
	return now.Format(time.DateOnly) + " Weekly Hungarian competitors dump"
}

type SMTPMailer struct {
	from     string
	password string
	to       []string
}

func NewSMTPMailer(from, password, to string) *SMTPMailer {
	recipients := make([]string, 0)
	for _, address := range strings.Split(to, ",") {
		if address = strings.TrimSpace(address); address != "" {
			recipients = append(recipients, address)
		}
	}
	return &SMTPMailer{from: from, password: password, to: recipients}
}

func (m *SMTPMailer) Send(subject, htmlBody string) error {
	auth := smtp.PlainAuth("", m.from, m.password, smtpHost)
	return smtp.SendMail(smtpHost+":"+smtpPort, auth, m.from, m.to, buildMessage(m.from, m.to, subject, htmlBody))
}

func buildMessage(from string, to []string, subject, htmlBody string) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "From: %s <%s>\r\n", mime.QEncoding.Encode("utf-8", senderName), from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(to, ", "))
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("utf-8", subject))
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
	b.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

	encoded := base64.StdEncoding.EncodeToString([]byte(htmlBody))
	for len(encoded) > 76 {
		b.WriteString(encoded[:76] + "\r\n")
		encoded = encoded[76:]
	}
	b.WriteString(encoded + "\r\n")
	return b.Bytes()
}
