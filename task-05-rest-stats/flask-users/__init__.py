import app
from pdf_generator import start_generator
import threading

# from app import *
# from app import PDF_LOCATION
# import database.db_reset
from pathlib import Path
from config import PDF_LOCATION

# For storing the pdfs
Path(PDF_LOCATION).mkdir(parents=True, exist_ok=True)

# Start generator


threading.Thread(target=start_generator, daemon=True).start()

print("__init___ done")