# -*- coding: utf-8 -*-
from copy import copy
from PyPDF2 import PdfFileWriter, PdfFileReader
from pdf_utils import get_page_size, crop_page
from reportlab.pdfgen import canvas
from reportlab.lib.units import cm
from reportlab.lib.colors import red
import io
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import cm

def get_small_pages(base_page, page_size=A4, border=1*cm,
                    span=cm, dush=[20, 12], line_weight=2, callback=None):
    base_page = copy(base_page)

    pdf_writer = PdfFileWriter()

    base_page_size = get_page_size(base_page)
    base_page_size_bordered = [val + border * 2 for val in base_page_size]
    subpage_num_x = get_parts_number(base_page_size_bordered[0], page_size[0], span)
    subpage_num_y = get_parts_number(base_page_size_bordered[1], page_size[1], span)


    total_subpage_num = subpage_num_y * subpage_num_x
    i = 0 
    for iy in range(subpage_num_y):
        for ix in range(subpage_num_x):
            mode = 0
            if (iy == subpage_num_y - 1):
                mode = 1
            if (ix == subpage_num_x - 1):
                mode = 2
            if (ix == subpage_num_x - 1 and iy == subpage_num_y - 1):
                mode = 3
            
            text = '{}.{}'.format(ix+1, iy+1)

            offset_rect = get_offset_rect(
                ix, iy, page_size, base_page_size, border, span)
            subpage = get_outline_page(
                base_page, text, offset_rect, page_size,
                span, dush, line_weight, mode)

            if callback:
                i += 1  
                is_want_stop = callback(subpage, i, total_subpage_num)
                if is_want_stop:
                    return

            pdf_writer.addPage(subpage)

    return pdf_writer


def get_parts_number(page_size, part_size, span):
    
    part_trimed_size = part_size - span
    n = page_size // part_trimed_size
    if (page_size % part_trimed_size) > span:
        n += 1
    return int(n)

def get_offset_rect(ix, iy, page_size, base_page_size, border, span):
    x1 = ix * (page_size[0] - span) - border
    y1 = base_page_size[1] - page_size[1] - iy * (page_size[1] - span) + border
    x2 = x1 + page_size[0]
    y2 = y1 + page_size[1]
    return (x1, y1, x2, y2)


def get_outline_page(base_page, text, offset_rect,
                     page_size, span, dush, line_weigth, mode=0):
    packet = io.BytesIO()
    c = canvas.Canvas(packet, page_size)

    y1 = span - line_weigth / 2
    x1 = (page_size[0] - span) + line_weigth
    y2 = span - line_weigth
    x2 = (page_size[0] - span) + line_weigth / 2
    text_gap = span * 0.2
    text_size = span * 0.8
    text_font = 'Helvetica-Bold'

    c.saveState()
    c.translate(offset_rect[0], offset_rect[1])
    c.setDash(dush[0], dush[1])
    c.setStrokeColor(red)
    c.setFillColor(red)
    c.setFont(text_font, text_size)
    c.setLineWidth(line_weigth)

    if mode == 0:
        c.line(x1, y1, 0, y1)
        c.line(x2, y2, x2, page_size[1])
    if mode == 1:
        c.line(x2, 0, x2, page_size[1])
    elif mode == 2:
        c.line(page_size[0], y1, 0, y1)

    c.drawRightString(page_size[0] - (span + text_gap), span + text_gap, text)

    c.restoreState() 
    c.save()
    
    packet.seek(0)
    new_pdf = PdfFileReader(packet)

    page = copy(base_page)
    crop_page(page, *offset_rect)
    page.mergePage(new_pdf.getPage(0))
    return page
