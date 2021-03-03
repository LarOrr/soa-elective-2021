FROM python:3.7

WORKDIR /usr/src/app

COPY requirements.txt ./

EXPOSE 8000

RUN pip install --no-cache-dir -r requirements.txt

COPY graphql_server.py ./

CMD [ "strawberry", "server", "graphql_server" ]