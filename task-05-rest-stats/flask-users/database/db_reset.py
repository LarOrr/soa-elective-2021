# Problems with import - to run from console use "python3 db_reset.py"
from flask_sqlalchemy import SQLAlchemy

from users_database import User, Stats
from app import app, db

with app.app_context():
    if db.exists:
        db.drop_all()
    db.create_all()
    # Testing data
    users = [
        {'username': 'user1', "password":"pass1", "email": "user1@ggmail.com",
        "gender": "male",
        "profile_pic": None,
         'stats':{'session_count': 2, 'victory_count':0,'loss_count':2, 'total_time':1000}},
        {'username': 'user2', "password":"pass2", "email": "user2@yahooo.org",
        "gender": None,
        "profile_pic": None,
    'stats':{'session_count': 12, 'victory_count':8,'loss_count':4, 'total_time':5000}}
        ]

    for user_dict in users:
        user = User()
        if not ('username' in user_dict and 'password' in user_dict):
            user_dict['username'] = user_dict['username'] + '_username'
            user_dict['password'] = user_dict['password'] + '_password'
        user.stats = Stats()
        user.update_info(user_dict)

    # if __name__ == '__main__':
