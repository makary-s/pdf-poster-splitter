#! /bin/bash 

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PY_ENV=".env"

if ! type "python3" > /dev/null; then 
    if ! type "brew" > /dev/null; then 
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)";
    fi
    brew install python@3.8
fi

if [ ! -d "$DIR/$PY_ENV" ]; then 
  pip3 install virtualenv; 
  python3 -m virtualenv $PY_ENV; 
  source $DIR/$PY_ENV/bin/activate;
  pip3 install -r requirements.txt; 
fi

source $DIR/$PY_ENV/bin/activate;
python $DIR/src/run.py;
