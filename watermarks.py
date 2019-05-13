# -*- coding: utf-8 -*-
from reportlab.pdfgen import canvas
import io, PIL, math
from pdf_utils import get_page_size
from PyPDF2 import PdfFileReader


def create_watermarked_page(page, logo, width, gap, rotation):
    page = copy(page)
    add_watermarks(page, logo, width, gap, rotation)
    return page

def add_watermarks(page, logo, width, gap, rotation):
    logo_w, logo_h = get_logo_size(logo, width)
    page_w, page_h = get_page_size(page)

    packet = io.BytesIO()
    c = canvas.Canvas(packet, (page_w, page_h))
    
    logo_box = get_logo_box(page_w, page_h, rotation)

    c.saveState()
    c.translate(logo_box['x'], logo_box['y'])
    c.rotate(rotation)
 
    logo_span = logo_w / 2 + gap
    col_num = int((logo_box['w'] + logo_span*2) / logo_w)
    row_num = int(logo_box['h'] / logo_h)
    for col in range(col_num):
        for row in range(row_num):
            y = row * (logo_h + gap)
            x = col * (logo_w + gap) - (logo_span)
            if row % 2:
                x += logo_w / 2
            c.drawImage(logo, x , y, logo_w, logo_h, mask='auto')

    c.restoreState()

    # c.showPage()
    c.save()
    packet.seek(0)
    new_pdf = PdfFileReader(packet)
    logos_page = new_pdf.getPage(0)
    
    page.mergePage(logos_page)


def get_logo_size(logo, logo_width):
    logo_w, logo_h = PIL.Image.open(logo).size
    logo_ratio = logo_w / logo_h
    logo_w = logo_width
    logo_h = logo_width / logo_ratio
    return (logo_w, logo_h)


def get_logo_box(page_w, page_h, rotation):
    r1 = math.radians(rotation)
    r2 = math.radians(90 - rotation)

    h1 = page_w * math.cos(r2)
    h2 = page_h * math.sin(r2)
    h = h1 + h2

    w1 = page_w * math.sin(r2)
    w2 = page_h * math.cos(r2)
    w = w1 + w2

    x = h1 * math.sin(r1)
    y = -h1 * math.cos(r1)

    return {
        'h': h,
        'w': w,
        'x': x,
        'y': y
    }