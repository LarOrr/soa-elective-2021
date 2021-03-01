# Docker
# address = '0.0.0.0' # 'localhost'
# mq_address = 'host.docker.internal'  # 'localhost'

#  Local
address = 'localhost'
mq_address = 'localhost'
# Set to False to allow anyone to access everything (for debug purposes)
require_auth = True

port = 5000
debug = True

USERS_URI = 'users'
PDF_LOCATION = './pdfs'

SQLALCHEMY_DATABASE_URI = 'sqlite:///database/users.db'
