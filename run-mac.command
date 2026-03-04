#! /bin/bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PY_ENV=".venv"

if ! python3 -c "import _tkinter" 2>/dev/null; then
    if ! type "brew" > /dev/null 2>&1; then
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    PY_VER=$(python3 -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")
    brew install "python-tk@${PY_VER}"
fi

if [ ! -d "$DIR/$PY_ENV" ]; then
    python3 -m venv "$DIR/$PY_ENV"
    source "$DIR/$PY_ENV/bin/activate"
    pip install -r "$DIR/requirements.txt"
fi

source "$DIR/$PY_ENV/bin/activate"
python3 "$DIR/src/run.py"
