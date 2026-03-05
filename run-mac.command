#! /bin/bash
set -e

export SYSTEM_VERSION_COMPAT=0

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PY_ENV="venv"

# Ensure Homebrew is installed
if ! type "brew" > /dev/null 2>&1; then
    /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
fi

# Use Homebrew Python (3.13+) instead of outdated system Python 3.9
BREW_PYTHON=$(brew --prefix python 2>/dev/null)/bin/python3
if [ ! -x "$BREW_PYTHON" ]; then
    echo "Installing Python via Homebrew..."
    brew install python python-tk
    BREW_PYTHON=$(brew --prefix python)/bin/python3
fi

PY_VER=$("$BREW_PYTHON" -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')")

# Install matching python-tk if missing
if ! "$BREW_PYTHON" -c "import _tkinter" 2>/dev/null; then
    echo "Installing python-tk@${PY_VER}..."
    brew install "python-tk@${PY_VER}"
fi

rebuild_venv() {
    rm -rf "$DIR/$PY_ENV"
    "$BREW_PYTHON" -m venv "$DIR/$PY_ENV"
    source "$DIR/$PY_ENV/bin/activate"
    pip install --upgrade pip
    pip install -r "$DIR/requirements.txt"
}

if [ ! -d "$DIR/$PY_ENV" ]; then
    rebuild_venv
fi

source "$DIR/$PY_ENV/bin/activate"

echo "Python: $(python3 --version)"
echo "BREW_PYTHON: $BREW_PYTHON"

if ! python3 -c "import PIL; import tkinter" 2>/dev/null; then
    echo "Venv broken, rebuilding..."
    rebuild_venv
fi

python3 "$DIR/src/run.py"
