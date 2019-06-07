# -*- coding: utf-8 -*-
from reportlab.lib.pagesizes import A4, A0
from reportlab.lib.units import cm

PAGE_SIZE = A4
BIG_PAGE_SIZE = [81, 50]

BORDER = 1 
CUT_SPAN = 1
CUT_WEIGHT = 2
CUT_DUSH = [20, 12]


LOGO_WIDTH = 30
LOGO_GAP = 3
LOGO_DEG = 30



#===============

BIG_PAGE_SIZE = [val * cm for val in BIG_PAGE_SIZE]
BORDER *= cm
CUT_SPAN *= cm
LOGO_WIDTH *= cm
LOGO_GAP *= cm