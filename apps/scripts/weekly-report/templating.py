import os

from jinja2 import Environment, FileSystemLoader, select_autoescape

from config import SCRIPT_DIR
from utils import FlagEmoji

TEMPLATES_PATH = os.path.join(SCRIPT_DIR, "templates")

html_environment = Environment(
    loader=FileSystemLoader(TEMPLATES_PATH),
    autoescape=select_autoescape(["html"]),
)
html_environment.filters["flag_emoji"] = FlagEmoji
