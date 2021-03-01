from app import db
if db.exists:
    db.drop_all()