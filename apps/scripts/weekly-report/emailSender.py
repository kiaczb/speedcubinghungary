import copy
import logging
import smtplib
from email.mime.text import MIMEText
from utils import ConvertResult, ConvertDate
from templating import html_environment
from config import emailFrom, emailTo, emailSubject, emailCode

logger = logging.getLogger(__name__)

def WriteEmail(report_data):
    SendEmail(RenderHtmlEmail(report_data))

def EmailCompetitions(competitions):
    selected = []
    for comp in competitions:
        if not comp.isHungarian:
            selected.append(comp)
            continue

        recorders = [person for person in comp.CompetitorWithRecords if person.Records]
        if recorders:
            filtered = copy.copy(comp)
            filtered.CompetitorWithRecords = recorders
            selected.append(filtered)
    return selected

def RenderHtmlEmail(report_data):
    template = html_environment.get_template('competitions.html')

    return template.render(
        convert_date=ConvertDate,
        convert_result=ConvertResult,
        **{**report_data, "competitions": EmailCompetitions(report_data["competitions"])}
    )

def SendEmail(html_body):
    msg = MIMEText(html_body, 'html', 'utf-8')
    msg['From'] = f"Bonsz <{emailFrom}>"
    msg['To'] = emailTo
    msg['Subject'] = emailSubject

    server = smtplib.SMTP("smtp.gmail.com", 587)
    try:
        server.starttls()
        server.login(emailFrom, emailCode)
        server.sendmail(emailFrom, emailTo, msg.as_string())
        logger.info("Email elküldve (%s -> %s).", emailFrom, emailTo)
    except Exception:
        logger.exception("Email küldése sikertelen.")
        raise
    finally:
        server.quit()
