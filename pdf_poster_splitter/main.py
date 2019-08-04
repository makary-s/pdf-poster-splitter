# -*- coding: utf-8 -*-
from PyPDF2 import PdfFileReader
from copy import copy
import os

import config as cnf
from pdf_utils import get_pdf_paths, get_ploter_page, save_pdf, get_pdf_page
from watermarks import add_watermarks
from devide_page import get_small_pages, get_small_pages_bordered

import subprocess

class ShellLogger():
    def __init__(self, format_str):
        self._texts = {}
        self._format_str = format_str

    def set(self, key, val):
        self._texts[key] = val

    def assign(self, dict_):
        for key, val in dict_.items():
            self._texts[key] = val


    def log(self):
        
        subprocess.call('cls', shell=True)
        subprocess.call('echo ' + self.join_text(), shell=True)
    
    def join_text(self):
        return self._format_str.format(**self._texts)

def main(options, get_increaser, gui):
    options = options.copy()
    print options
    if not os.path.exists(options['base_path']):
        gui.show_popup('Директории %s не существует!' % options['base_path'])
        return 
    if not os.path.exists(options['target_path']):
        gui.show_popup('Директории %s не существует!' % options['target_path'])
        return
        
    if options['logo_path'].isspace():
            options['logo_path'] = None
    elif not os.path.exists(options['logo_path']):
        gui.show_popup('Директории %s не существует!' % options['logo_path'])
        return
    elif not os.path.isfile(options['logo_path']):
        gui.show_popup('%s не является файлом!' % options['logo_path'])
        return
    if options['title_path'].isspace():
        options['title_path'] = None
    elif not os.path.exists(options['title_path']):
        gui.show_popup('Директории %s не существует!' % options['title_path'])
        return
    elif not os.path.isfile(options['title_path']):
        gui.show_popup('%s не является файлом!' % options['title_path'])
        return

    shell_logger = ShellLogger("{i}/{len} : {pi}/{plen} : {path}")

    pdfs = []
    for cur_dir, folders, files in os.walk(options['base_path']):
        target_dir = os.path.normpath(cur_dir.replace(options['base_path'], options['target_path']))
        cur_pdfs = get_pdf_paths(cur_dir)
        for path in cur_pdfs:
            pdf = {
            'path': path,
            'target_path': target_dir,
            'name': os.path.splitext(os.path.basename(path))[0]
            }
            pdfs.append(pdf)

    increase_progress = get_increaser(len(pdfs))

    for pdf_i, pdf in enumerate(pdfs):
        shell_logger.assign({
            'i': pdf_i + 1,
            'len': len(pdfs),
            'path': pdf['path'] })

        base_page = get_pdf_page(pdf['path'])
        if options['logo_path']:
            add_watermarks(
                base_page, options['logo_path'], cnf.LOGO_WIDTH,
                cnf.LOGO_GAP, cnf.LOGO_DEG)
            save_pdf(base_page, pdf['name'] + '.pdf', pdf['target_path'])
            gui.update()

        ###
        def callback(p, i, len):
            shell_logger.assign({'pi': i, 'plen': len}),
            shell_logger.log(),
            gui.update()
            return gui.is_want_stop

        pages_a4 = get_small_pages_bordered(
                base_page, cnf.PAGE_SIZE, cnf.CUT_SPAN,
                cnf.CUT_DUSH, cnf.CUT_WEIGHT, callback)
        if pages_a4:
            if options['title_path']:
                pages_a4.insertPage(get_pdf_page(options['title_path']), 0)
            save_pdf(pages_a4, pdf['name'] + '_A4.pdf', pdf['target_path'])
        
        ###
        if gui.is_want_stop:
            gui.show_popup('Остановлено!')
            increase_progress(False)
            return
        increase_progress()

    gui.show_popup('Задача завершена!')

# if __name__ == '__main__':
    # main(options)
    

    
 

