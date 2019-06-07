# -*- coding: utf-8 -*-

import Tkinter as tk
import tkFileDialog as tkfd
import ttk
import pickle, os

from main import main

import sys
reload(sys)
sys.setdefaultencoding('utf8')

PICKLE_PATH = 'data.pickle'
MAIN_SCRIPT = main
options = {
    'base_path': '',
    'target_path': '',
    'logo_path': '',
    'title_path': ''}

if os.path.isfile(PICKLE_PATH):
    with open(PICKLE_PATH, 'rb') as f:
        options = pickle.load(f)
else:
    with open(PICKLE_PATH, 'wb') as f:
        pickle.dump(options, f)

class TkApp(tk.Tk, object):
    def __init__(self, *args, **kwargs):
        super(TkApp, self).__init__(*args, **kwargs)

        self.title('Лекала')
        self.protocol('WM_DELETE_WINDOW', self.on_exit)

        self.frames = {}
        for frame in [MainFrame]:
            self.frames[frame.__name__] = frame = frame(self)
            frame.grid(row=2, column=0, sticky="nsew")

    def show_tab(self, name):
        self.frames[name].tkraise()

    def on_exit(self):
        with open(PICKLE_PATH, 'wb') as f:
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
    



def create_callback(option, var=None):
    def wrapper(_var=var):
        options[option] = os.path.normpath(_var.get())
        return True
    return wrapper

class MainFrame(tk.Frame, object):
    def __init__(self, parent):
        super(MainFrame, self).__init__(parent)
        self.parent = parent
        self.is_run = False
        self.is_want_stop = False

        paths_frame = tk.Frame(self)
        paths_frame.grid_columnconfigure(1, weight=1)

        place_entery_row(paths_frame, 1, 'Из папки', 'folder', create_callback('base_path'), options['base_path'])
        place_entery_row(paths_frame, 2, 'В папку', 'folder', create_callback('target_path'), options['target_path'])
        place_entery_row(paths_frame, 3, 'Логотип*', 'png', create_callback('logo_path'), options['logo_path'])
        place_entery_row(paths_frame, 4, 'Титульный лист*', 'pdf', create_callback('title_path'), options['title_path'])
        paths_frame.pack(fill='x', expand=True)


        run_frame = tk.Frame(self)
        progressbar = myProgressbar(run_frame)
        progressbar.pack(side='left', fill='both', expand=True)

        button_run_text = tk.StringVar()
        button_run_text.set('Начать')
        button_run = tk.Button(
            run_frame,
            textvariable=button_run_text,
            command=lambda: self.handle_run(button_run_text, progressbar))
        button_run.pack(side='left')
        run_frame.pack(fill='both')

    def handle_run(self, button_text, progressbar):
        # loose entery focuse to save path
        self.focus()
        self.update()
        
        if not self.is_run:
            button_text.set('Отменить')
            self.is_run = True
            MAIN_SCRIPT(
                    options,
                    progressbar.get_increaser,
                    self)
            button_text.set('Начать')
            self.is_run = False
            self.is_want_stop = False
        else:
            self.is_want_stop = True

    def show_popup(self, about_message, width=200):
        top  = tk.Toplevel()
        top.title("Внимание")
        top.resizable(0, 0)
        
        msg = tk.Message(top, text=about_message, width=width)
        msg.pack(fill='both', expand=True)


        button = tk.Button(top, text="Ok", command=top.destroy)
        button.pack()
        
        x = self.parent.winfo_x()
        y = self.parent.winfo_y()
        dx = dy = 20
        top.geometry("+%d+%d" % (x + dx, y + dy))

class myProgressbar(ttk.Progressbar, object):
    def __init__(self, parent, *args, **kwargs):
        super(myProgressbar, self).__init__(parent)
        
        self.parent = parent
        self.variable = tk.DoubleVar(self)
        self.max_val = 0

        self.style = ttk.Style(self)
        self.style.layout('text.Horizontal.TProgressbar', 
                    [('Horizontal.Progressbar.trough',
                        {'children': [('Horizontal.Progressbar.pbar',
                                      {'side': 'left', 'sticky': 'ns'})],
                         'sticky': 'nswe'}), 
                     ('Horizontal.Progressbar.label', {'sticky': ''})])
        self.style.configure('text.Horizontal.TProgressbar', text='')

    
        self.configure(
            style='text.Horizontal.TProgressbar',
            mode='determinate',
            variable=self.variable)

    def increase(self):
        self.step() 
        self.style.configure('text.Horizontal.TProgressbar', 
                        text='{:g}/{:g}'.format(self.variable.get(), self.max_val))

    def get_increaser(self, max_val):
        self.max_val = max_val
        self.config(maximum=max_val)
        def increase(mode=True):
            if mode:
                self.increase()
            else:
                self.reset()
        return increase

    def reset(self):
        self.max_val = 0
        self.config(maximum=0)
        self.variable.set(0)
        self.style.configure('text.Horizontal.TProgressbar', text='')


def create_path_getter(atype, parent, entery, var):
    def wrapper():
        if atype == 'folder':
            path = tkfd.askdirectory()
        else:
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

def place_entery_row(parent, row, lable, atype, callback, deftext=''):

    lable_logo_path = tk.Label(parent, text=lable)
    lable_logo_path.grid(row=row, column=0, sticky='w')

    var_path = tk.StringVar()
    var_path.set(deftext)
    entery = tk.Entry(parent, textvariable=var_path, validate="all", validatecommand=lambda: callback(var_path))
    entery.grid(row=row, column=1, sticky='ew')
    entery.xview_moveto(1)
    
    command =  create_path_getter(atype, parent, entery, var_path)
    button_logo_path = tk.Button(parent,
        text='Найти',
        command=command)
    button_logo_path.grid(row=row, column=3, sticky='e')

tkApp = TkApp()
tkApp.mainloop()
