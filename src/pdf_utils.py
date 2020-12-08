# -*- coding: utf-8 -*-
import  PIL, os
from copy import copy
from reportlab.lib.units import cm
from PyPDF2.pdf import PageObject
from PyPDF2 import PdfFileWriter, PdfFileReader
import datetime, time

def get_pdf_page(path, page_index=0):
    return PdfFileReader(open(path, "rb")).getPage(page_index)

def get_page_size(page):
    page = copy(page)  #QUEST! почему ломется без этого???
    page_w = float(page.cropBox.upperRight[0]) 
    page_h = float(page.cropBox.upperRight[1])

    # print(page.get('/Rotate')) #QUEST! почему поворот может сохраняться???

    return (page_w, page_h)

def crop_page(page, x1, y1, x2, y2):
    page.mediaBox.setLowerLeft((x1, y1))
    page.mediaBox.setUpperRight((x2, y2))

    page.cropBox.setLowerLeft((x1, y1))
    page.cropBox.setUpperRight((x2, y2))

    page.trimBox.setLowerLeft((x1, y1))
    page.trimBox.setUpperRight((x2, y2))

    page.bleedBox.setLowerLeft((x1, y1))
    page.bleedBox.setUpperRight((x2, y2))

    page.artBox.setLowerLeft((x1, y1))
    page.artBox.setUpperRight((x2, y2))


def get_pdf_paths(path):
    pdfs = []
    for f in os.listdir(path):
        filepath = os.path.normpath(os.path.join(path, f))
        if os.path.isfile(filepath):
            ext = os.path.splitext(filepath)[1].lower()
            if  ext == '.pdf':
                pdfs.append(filepath)
    return pdfs


def get_ploter_page(page, width, min_height, debug_path=None, debug_targetpath=None,  debug_filename=None):
    page = copy(page)
    page_w, page_h = get_page_size(page)

    dw = width - page_w
    dh = width - page_h

    if (abs(dw) <= abs(dh) and dw >= -0.1 * cm) or \
       (dw >= -0.1 * cm and dh <= -0.1 * cm):
        rotation = 0
        offset_x = (width - page_w) / 2
        offset_y = 0
        print_page_w = width
        print_page_h = min_height if min_height >= page_h else page_h

    elif dh >= -0.1*cm:
        rotation = 90
        offset_x = 0
        offset_y = (width - page_h) / 2
        print_page_w = min_height if min_height >= page_w else page_w
        print_page_h = width
    else:
        if (debug_path):
            st = datetime.datetime.fromtimestamp(time.time()).strftime('%Y-%m-%d %H:%M:%S')
            error_message = st + ' > Страница слишком большая! > ' + debug_path
            print(error_message)
            logpath = os.path.join(debug_targetpath, debug_filename) + '.txt'
            with open(logpath, 'a') as file:
                file.write(error_message + '\n')
            return PageObject.createBlankPage(width=1, height=1)
        else:
            raise Exception('PAGE IS TOO BIG!')

    big_page = PageObject.createBlankPage(width=print_page_w, height=print_page_h)
    big_page.mergeTranslatedPage(page, offset_x, offset_y)
    big_page.rotateClockwise(rotation)

    return big_page


def save_pdf(pages, name, dir):
    if isinstance(pages, PdfFileWriter):
        pdf_writer = pages
    elif isinstance(pages, PageObject):
        pdf_writer = PdfFileWriter()
        pdf_writer.addPage(pages)

    if not os.path.exists(dir):
        os.makedirs(dir)
    
    path = os.path.join(dir, name)
    with open(path, "wb") as out:
        pdf_writer.write(out)
    return 