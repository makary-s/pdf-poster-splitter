# -*- coding: utf-8 -*-

import tkinter as tk
from tkinter import filedialog as tkfd
from tkinter import ttk
import pickle, os, datetime, traceback
from importlib import reload
from config import PAGE_SIZES
from main import main

import sys
reload(sys)

TEMP_PATH = '.temp'
PICKLE_PATH = 'data.pickle'
MAIN_SCRIPT = main
global EXIT_FLAG
EXIT_FLAG = False

options = {
    'base_path': '',
    'target_path': '',
    'logo_path': '',
    'title_path': '',
    'page_size': ''}

if not os.path.exists(TEMP_PATH):
    os.makedirs(TEMP_PATH)

pickle_path = os.path.join(TEMP_PATH, PICKLE_PATH)
if os.path.isfile(pickle_path):
    with open(pickle_path, 'rb') as f:
        options = pickle.load(f)
else:
    with open(pickle_path, 'wb') as f:
        pickle.dump(options, f)

class TkApp(tk.Tk, object):
    def __init__(self, *args, **kwargs):
        super(TkApp, self).__init__(*args, **kwargs)

        self.title('Лекала')
        self.protocol('WM_DELETE_WINDOW', self.on_exit)

        self.frames = {}
        for frame in [MainFrame]:
            self.frames[
                frame.__name__ # type: ignore
            ] = frame = frame(self) 
            frame.grid(row=2, column=0, sticky="nsew")

    def show_tab(self, name):
        self.frames[name].tkraise()

    def on_exit(self):
        global EXIT_FLAG
        EXIT_FLAG = True
        with open(pickle_path, 'wb') as f:
            pickle.dump(options, f)
        self.destroy()

   

# class TabFrame(tk.Frame):
#     def __init__(self, parent):
#         super().__init__(parent)
        
#         button_tab_slice = tk.Button(self,
#             text='Нарезка',
#             command=lambda: parent.show_tab("SliceFrame") )
#         button_tab_slice.pack(side='left')

#         button_tab_slice = tk.Button(self,
#             text='Лист для плоттера',
#             command=lambda: parent.show_tab("PlotterFrame") )
#         button_tab_slice.pack(side='left')

#         button_tab_slice = tk.Button(self, text='Вотермарки')
#         button_tab_slice.pack(side='left')

def saveOption(option_name: str, var: tk.Variable, prepare = lambda x: x):
    res_var = var.get()
    options[option_name] = prepare(res_var)

    return True

class MainFrame(tk.Frame, object):
    def __init__(self, parent):
        super(MainFrame, self).__init__(parent)
        self.parent = parent
        self.is_run = False
        self.is_want_stop = False

        paths_frame = tk.Frame(self)
        paths_frame.grid_columnconfigure(1, weight=1)

        place_entery_row(paths_frame, 1, 'Из папки', 'folder', 'base_path', options)
        place_entery_row(paths_frame, 2, 'В папку', 'folder', 'target_path', options)
        place_select_row(paths_frame, 3, 'Формат бумаги', PAGE_SIZES, 'page_size', options)
        place_entery_row(paths_frame, 4, 'Логотип*', 'png', 'logo_path', options)
        place_entery_row(paths_frame, 5, 'Титульный лист*', 'pdf', 'title_path', options)
        paths_frame.pack(fill='x', expand=True)


        run_frame = tk.Frame(self)
        progressbar_f = myProgressbar(run_frame, 'fp')
        progressbar_f.pack(side='left', fill='both', expand=True)
        progressbar_p = myProgressbar(run_frame, 'pp')
        progressbar_p.pack(side='left', fill='both', expand=True)

        button_run_text = tk.StringVar()
        button_run_text.set('Начать')
        button_run = tk.Button(
            run_frame,
            textvariable=button_run_text,
            command=lambda: self.handle_run(button_run_text, progressbar_f, progressbar_p))
        button_run.pack(side='left')
        run_frame.pack(fill='both')
        
        
    def create_api(self, progressbar_f, progressbar_p):
        api = lambda: None
        api.update = self.update
        api.check_stop = lambda: self.is_want_stop
        api.show_popup = self.show_popup
        api.set_file_progress = progressbar_f.set_progress
        api.set_page_progress = progressbar_p.set_progress
        return api
        
    def handle_run(self, button_text, progressbar_f, progressbar_p):
        # loose entery focuse to save path
        self.focus()
        self.update()
        
        if not self.is_run:
            button_text.set('Отменить')
            self.is_run = True
            try:
                MAIN_SCRIPT(
                        options,
                        self.create_api(progressbar_f, progressbar_p))
            except Exception as e:
                global EXIT_FLAG
                if not EXIT_FLAG:
                    print(e)
                    with open('error_log.txt', 'a') as f:
                        log = '\n' + str(datetime.datetime.now()) + '\n' +  str(traceback.format_exc()) + '\n' + str(e) + '\n'
                        f.write(log)
                    self.show_popup('Ошибка!')
            button_text.set('Начать')
            self.is_run = False
            self.is_want_stop = False
        else:
            self.is_want_stop = True

    def show_popup(self, about_message, width=200):
        top  = tk.Toplevel()
        top.title("Внимание")
        top.resizable(width=None, height=None)
        
        msg = tk.Message(top, text=about_message, width=width)
        msg.pack(fill='both', expand=True)


        button = tk.Button(top, text="Ok", command=top.destroy)
        button.pack()
        
        x = self.parent.winfo_x()
        y = self.parent.winfo_y()
        dx = dy = 20
        top.geometry("+%d+%d" % (x + dx, y + dy))

class myProgressbar(ttk.Progressbar, object):
    def __init__(self, parent, name="LabeledProgressbar"):
        super(myProgressbar, self).__init__(parent,
            orient="horizontal",
            mode="determinate")
        
        self.name = name
        
        self.style = ttk.Style(parent)
        self.style.layout(self.name,
         [(self.name + '.trough',
           {'children': [(self.name + '.pbar',
                          {'side': 'left', 'sticky': 'ns'}),
                         (self.name + ".label",
                          {"sticky": ""})],
           'sticky': 'nswe'})])
        self['style'] = self.name
        self["value"] = 0
        self["maximum"] = 0

        self.parent = parent
        self.set_progress(0, 0)

    def set_progress(self, i, imax):
        self["value"] = i
        self["maximum"] = imax
        self.style.configure(self.name, text="{}/{}        ".format(i, imax))

def create_path_getter(atype, parent, entery, var):
    def wrapper():
        if atype == 'folder':
            path = tkfd.askdirectory()
        else:
            filetypes = []
            if atype == 'png':
                filetypes = [("png files", "*.png")]
            elif atype == 'pdf':
                filetypes = [("pdf files", "*.pdf")]
                
            path = tkfd.askopenfilename(
                parent=parent,
                filetypes = filetypes,
                multiple=False,
                title='Choose a file')
        var.set(path)
        entery.xview_moveto(1)
    return wrapper

def prepare_path(path: str):
     return os.path.normpath(path) if path else ''

def place_entery_row(parent, row, lable, atype, option_name, options):
    defval = options[option_name] if option_name in options else ''

    tk.Label(parent, text=lable).grid(row=row, column=0, sticky='w')

    tkvar = tk.StringVar()
    tkvar.set(defval)
    entery = tk.Entry(parent, textvariable=tkvar, validate="all", validatecommand=lambda: saveOption(option_name, tkvar, prepare_path))
    entery.grid(row=row, column=1, sticky='ew')
    
    command =  create_path_getter(atype, parent, entery, tkvar)
    button_logo_path = tk.Button(parent,
        text='Найти',
        command=command)
    button_logo_path.grid(row=row, column=3, sticky='e')

def place_select_row(parent: tk.Frame, row: 'int', lable: 'str', items: 'dict[str, tuple[float]]', option_name, options: 'dict[str, str]'):
    defval = options[option_name] if option_name in options else None

    tk.Label(parent, text=lable).grid(row=row, column=0, sticky='w')

    tkvar = tk.StringVar()
    tkvar.set(defval if defval != None else list(items.keys())[0])

    entery = tk.OptionMenu(parent, tkvar, *items)

    tkvar.trace('w', lambda *x: saveOption(option_name, tkvar, prepare_path))
    
    entery.grid(row=row, column=1, sticky='ew')

tkApp = TkApp()
tkApp.mainloop()
