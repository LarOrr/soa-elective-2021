import pika
from flask import Flask, jsonify, request, abort, Response, send_file
from flask_cors import CORS
from flask_jwt_extended import JWTManager, jwt_required, create_access_token, get_jwt_identity
from sqlalchemy.exc import IntegrityError
from users_database import db
from config import USERS_URI, PDF_LOCATION
from users_database import User, Stats

import config

app = Flask(__name__)
# Enable CORS
CORS(app)
# Beautiful JSON
app.config['JSONIFY_PRETTYPRINT_REGULAR'] = True
# Database configuration
app.config['SQLALCHEMY_DATABASE_URI'] = config.SQLALCHEMY_DATABASE_URI
app.config['SQLALCHEMY_TRACK_MODIFICATIONS'] = True
# Prefix for all resources
# app.config["APPLICATION_ROOT"] = "v1"
# TODO This should be hidden
app.config['JWT_SECRET_KEY'] = 'Super_Secret_JWT_KEY'
app.config['JWT_ACCESS_TOKEN_EXPIRES'] = False
# If config is off - everyone can access anything
if not config.require_auth:
    jwt_required = lambda fn: fn
# Database manager
# db = SQLAlchemy(app)
db.init_app(app)

jwt = JWTManager(app)


# TODO move it to the separate file?
# region  -------------------------------------------- Login and JWT ----------------------------------------------------------------
@app.route('/login', methods=['POST'])
def login():
    data = request.json
    username = data['username']
    password = data['password']

    if not username or not password:
        return http_error('Bad username or password', 400)

    user: User = User.query.filter_by(username=username, password=password).first()
    if user is None:
        return http_error('Bad username or password', 401)
    # Identity by id
    access_token = create_access_token(identity=user.id)
    return jsonify({'success': True, 'token': access_token, 'user': "/users/{}".format(user.id)}), 200


@app.route('/users/me', methods=['GET'])
@jwt_required
def get_auth_info():
    """
    :return: info about currently authorized user
    """
    return get_user(get_jwt_identity())


def check_access(user_id: int):
    """
    Checks if user_id is the same as the id of currently authorized user.
    If not - returns code 403.
    """
    # id of authorized user account
    if not config.require_auth:
        return

    current_id = get_jwt_identity()
    # If it's not account of this user then return 403 Forbidden
    if current_id != user_id:
        abort(403)


# endregion ------------------------------------------ // LOGIN ------------------------------------------------------------------

# region ---common---
def http_error(message: str, status_code: int):
    return jsonify({'success': False, 'message': message}), status_code


# endregion

#  region /users

@app.route(f'/{USERS_URI}', methods=['GET'])
def get_all_users():
    """
    :return: all users in the database in json
    """
    users = list(map(lambda user: user.to_dict(), User.query.all()))

    return jsonify(users)


@app.route(f'/{USERS_URI}/<id>', methods=['GET'])
def get_user(id: int):
    """
    :return: all users in the database in json
    """
    u = User.query.get(id)
    if u is None:
        return http_error('No such user', 404)
    # If HEAD returns nothing

    return jsonify(u.to_dict())


@app.route(f'/{USERS_URI}', methods=['POST'])
def register_user():
    """
    Creates new user
    :return: ID of the new user
    """
    data = request.json
    if not data['username'] or not data['password']:
        return http_error('Bad username or password', 400)
    new_user = User()
    new_user.Stats = Stats()
    try:
        # data['stats'] = new_user.Stats
        new_user.update_info(data)
    except IntegrityError:
        return http_error('User with such name already exists', 400)
    res = {'user': new_user.to_dict(), 'token': create_access_token(identity=new_user.id)}
    return jsonify(res)


@app.route(f'/{USERS_URI}/<id>', methods=['PUT', 'PATCH'])
@jwt_required
def patch_place(id: int):
    """
    Changes info about the user
    :param id: id of the user
    """
    id = int(id)
    check_access(id)
    user = User.query.get(id)
    data = request.json
    user.update_info(data)
    return get_user(id)


#
@app.route(f'/{USERS_URI}/<id>', methods=['DELETE'])
@jwt_required
def delete_user(id: int):
    """
    Deletes user with the id
    :param id: id of the user
    :return info about deleted user
    """
    id = int(id)
    check_access(id)
    # TODO use try except to catch 404 instead
    user: User = User.query.get(id)
    if not user:
        return http_error("No such user", 404)
    user.delete()
    return Response(status=204)


# endregion

# region --- stats ---
@app.route(f'/{USERS_URI}/<user_id>/stats', methods=['GET'])
@jwt_required
def get_stats(user_id: int):
    """
    :return: stats of user
    """
    user_id = int(user_id)
    check_access(user_id)
    u = User.query.get(user_id)
    if u is None:
        return http_error('No such user', 404)
    return jsonify(u.stats.to_dict())


@app.route(f'/{USERS_URI}/<user_id>/stats', methods=['PUT', 'PATCH'])
@jwt_required
def update_stats(user_id: int):
    """
    Update stats of user
    :return: updated stats
    """
    user_id = int(user_id)
    check_access(user_id)
    u: User = User.query.get(user_id)
    if u is None:
        return http_error('No such user', 404)
    data = {'stats': request.json}
    u.update_info(data)
    return jsonify(u.stats.to_dict())


@app.route(f'/{USERS_URI}/<user_id>/pdf', methods=['POST'])
@jwt_required
def request_stats_file(user_id: int):
    """
    Creates request for generating pdf file with user stats
    :return: url for GET, where the pdf will be available
    """
    user_id = int(user_id)
    check_access(user_id)
    user: User = User.query.get(user_id)
    if user is None:
        return http_error('No such user', 404)
    # SEND STATS TO CREATE PDF
    request_pdf_generation(user)
    return jsonify({"url": f'/{USERS_URI}/{user.id}/pdf'})


# endregion

# region --- pdf generation ---
@app.route(f'/{USERS_URI}/<user_id>/pdf', methods=['GET'])
@jwt_required
def get_stats_file(user_id: int):
    """
    :returns previously generated pdf file with stats
    """
    user_id = int(user_id)
    check_access(user_id)
    user: User = User.query.get(user_id)
    if user is None:
        return http_error('No such user', 404)
    # Sends file
    try:
        return send_file(PDF_LOCATION + f'/{user.id}.pdf', attachment_filename=f'{user.username}_stats.pdf')
    except FileNotFoundError:
        return http_error('No such file. Send generation request first.', 404)


def request_pdf_generation(user: User):
    # Creating string of profile
    stats = user.stats
    text = 'User profile:\n'
    for attr, value in user.to_dict().items():
        text += f'{attr}:{value}\n'

    text += '\n-------------------------------------------------\n'
    text += '\nUser\'s statistics:\n'
    for attr, value in stats.to_dict().items():
        text += f'{attr}:{value}\n'

    # Sending data to the pdf_generator
    connection = pika.BlockingConnection(
        pika.ConnectionParameters(host=config.mq_address))
    channel = connection.channel()

    channel.queue_declare(queue='pdf_gen')

    filename = PDF_LOCATION + "/" + str(user.id)
    pdf_request = '{"filename":"' + filename + '","text":"' + text + '"}'
    # This fails with KeyError
    # '{"filename":"{}","text":"{}"}'.format(filename, text)
    channel.basic_publish(exchange='', routing_key='pdf_gen',
                          body=pdf_request.encode('utf-8'))
    print(" [x] PDF request sent")
    connection.close()


#  endregion

if __name__ == '__main__':
    app.run(host=config.address, port=config.port, debug=config.debug)
