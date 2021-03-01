from app import app, db

with app.app_context():
    if db.exists:
        db.drop_all()