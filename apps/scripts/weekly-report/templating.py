"""Közös Jinja2 environment a riport e-mail sablonjához.

Az e-mail HTML-t renderel (templates/competitions.html). A weboldalra kerülő
Post NEM sablonból készül, hanem strukturált JSON adatból (lásd reportPayload.py),
amit a frontend WeeklyReportView komponense jelenít meg.
"""

import os

from jinja2 import Environment, FileSystemLoader

from config import SCRIPT_DIR
from utils import FlagEmoji

TEMPLATES_PATH = os.path.join(SCRIPT_DIR, "templates")

# HTML (e-mail): a whitespace megtartva, ahogy eddig is.
html_environment = Environment(loader=FileSystemLoader(TEMPLATES_PATH))
html_environment.filters["flag_emoji"] = FlagEmoji
