if not exist .env (
    pip install virtualenv && python -m virtualenv .env && .env\Scripts\activate && pip install -r requirements.txt && .env\Scripts\activate && python src\run.py
) else (
    .env\Scripts\activate && python src\run.py
)
