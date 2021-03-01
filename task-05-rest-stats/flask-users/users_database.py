# from app import db, USERS_URI
from flask_sqlalchemy import SQLAlchemy
from config import USERS_URI

db = SQLAlchemy()

class Stats(db.Model):
    def __init__(self):
        self.session_count = self.loss_count = self.total_time = self.victory_count = 0

    id = db.Column(db.Integer, primary_key=True)
    session_count = db.Column(db.Integer)
    victory_count = db.Column(db.Integer)
    loss_count = db.Column(db.Integer)
    total_time = db.Column(db.Integer)
    attrs = ['session_count', 'victory_count', 'loss_count', 'total_time']

    def to_dict(self) -> dict:
        """
         :return: Dict that contain only attributes with needed information
        """
        res = {}
        for attr in self.attrs:
            res[attr] = getattr(self, attr)
        # res = self.__dict__
        return res


class User(db.Model):
    id = db.Column(db.Integer, primary_key=True)
    username = db.Column(db.String(32), unique=True)
    # TODO add password hashing
    password = db.Column(db.String(128))
    email = db.Column(db.String())
    gender = db.Column(db.String())
    profile_pic = db.Column(db.BLOB)
    stats_id = db.Column(db.Integer, db.ForeignKey(Stats.id))
    stats = db.relationship(Stats, backref='stats', uselist=False)
    attrs = ['username', 'email', 'gender', 'profile_pic']
    additional_update_attrs = ['password']

    def delete(self):
        """
        Deletes place from database
        :return: id
        """
        Stats.query.filter_by(id=self.stats.id).delete()
        User.query.filter_by(id=self.id).delete()
        db.session.commit()
        return self.id

    def update_info(self, data: dict):
        """
            Updates place with information given in json
            :param data: json with data
            :return: Place id
            :exception possible KeyError or TypeError if something is wrong with data
            """
        if not  self.stats:
            self.stats = Stats()
        stats = self.stats
        # TODO check for data types - for now we suppose that all coming data is correct
        for attr in data.keys():
            # Setting all attributes
            if attr in (User.attrs + User.additional_update_attrs):
                setattr(self, attr, data[attr])
                # setattr(new_place, attr, None)
        if 'stats' in data:
            for attr in data['stats'].keys():
                # Setting all attributes
                if attr in Stats.attrs:
                    setattr(stats, attr, data['stats'][attr])
        db.session.add(self)
        # db.session.add(stats)
        db.session.commit()
        db.session.refresh(self)
        return self.id

    def to_dict(self) -> dict:
        """
        :return: Dict that contain only attributes with needed information
        """
        # + id and stats
        res = {'self': f'/{USERS_URI}/{self.id}', 'stats': f'/{USERS_URI}/{self.id}/stats'} # 'self.stats.to_dict()'
        for attr in self.attrs:
            res[attr] = getattr(self, attr)
        return res