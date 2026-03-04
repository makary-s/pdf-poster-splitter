@echo off
setlocal

set PY_ENV=.venv

if not exist %PY_ENV% (
    python -m venv %PY_ENV%
    call %PY_ENV%\Scripts\activate.bat
    pip install -r requirements.txt
)

call %PY_ENV%\Scripts\activate.bat
python src\run.py
