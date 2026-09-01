import logging
import smtplib
from email.mime.text import MIMEText
from utils import ConvertResult, ConvertDate
from templating import html_environment
from config import emailFrom, emailTo, emailSubject, emailCode

logger = logging.getLogger(__name__)

def WriteEmail(report_data):
    """Elküldi a riportot e-mailben.

    A riport adatait NEM itt állítjuk elő: a hívó adja át (lásd reportData.py),
    mert ugyanazt az adathalmazt használja a weboldalra kerülő Post is.
    """
    SendEmail(RenderHtmlEmail(report_data))

def RenderHtmlEmail(report_data):
    template = html_environment.get_template('competitions.html')

    return template.render(
        convert_date=ConvertDate,
        convert_result=ConvertResult,
        **report_data
    )
def SendEmail(html_body):
    """Elküldi a riportot e-mailben.

    Sikertelen küldésnél a hibát logoljuk ÉS továbbdobjuk — a job álljon le
    hibával (lásd main.py + a hívási sorrend ott), ne jelentsen csendben
    sikert, ha az email valójában nem ment el. (Korábban a kivétel csak
    print()-elve lett elnyelve, a job pedig "sikeresként" futott tovább.)
    """
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
