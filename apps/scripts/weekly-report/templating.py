"""Közös Jinja2 environmentek a riport sablonokhoz.

Ugyanaz a riport kétféle formában jelenik meg:
  - HTML  -> e-mail (templates/competitions.html)
  - szöveg -> weboldalra kerülő Post (templates/post.txt)

A sablonútvonal és a szűrők ezért egy helyen vannak definiálva, hogy ne
duplikálódjanak a renderelők között.
"""

import os

from jinja2 import Environment, FileSystemLoader

from config import SCRIPT_DIR
from utils import FlagEmoji

TEMPLATES_PATH = os.path.join(SCRIPT_DIR, "templates")

_loader = FileSystemLoader(TEMPLATES_PATH)

# HTML (e-mail): a whitespace megtartva, ahogy eddig is.
html_environment = Environment(loader=_loader)
html_environment.filters["flag_emoji"] = FlagEmoji

# Sima szöveg (Post): trim_blocks + lstrip_blocks nélkül a blokk-tagek felesleges
# üres sorokat és behúzást hagynának a kimenetben.
text_environment = Environment(loader=_loader, trim_blocks=True, lstrip_blocks=True)
text_environment.filters["flag_emoji"] = FlagEmoji
