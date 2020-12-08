if not exist src (cd ..)
pip install virtualenv && python -m virtualenv .env && .env\Scripts\activate && pip install -r requirements.txt
