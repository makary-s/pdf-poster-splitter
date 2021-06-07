# -*- coding: utf-8 -*-
from reportlab.lib import pagesizes
from reportlab.lib.units import cm
from collections import namedtuple

def getSizeNames():
    sizes = [(name, getattr(pagesizes, name)) for name in dir(pagesizes) if name[0].isupper()]

    pretty_sizes: 'dict[str, tuple[float]]' = {};
    for (name, values) in sizes:
        [width, height] = [int(round(val / cm, 2) * 10) for val in values]
        pretty_name = "{name} ({width}x{height})".format(name=name, width=width, height=height)
        pretty_sizes[pretty_name] = values

    return pretty_sizes

PAGE_SIZES = getSizeNames()

BORDER = 1 
CUT_SPAN = 1
CUT_WEIGHT = 2
CUT_DUSH = [20, 12]

LOGO_WIDTH = 30
LOGO_GAP = 3
LOGO_DEG = 30

BORDER *= cm
CUT_SPAN *= cm
LOGO_WIDTH *= cm
LOGO_GAP *= cm
