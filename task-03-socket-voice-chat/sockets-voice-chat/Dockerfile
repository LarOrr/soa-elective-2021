FROM python:3

WORKDIR /usr/src/app

COPY requirements.txt ./

EXPOSE 4444/udp

# additional dependencies
# We need it for PyAudio
RUN apt-get update \
        && apt-get install libportaudio2 libportaudiocpp0 portaudio19-dev libsndfile1-dev -y

RUN pip install --no-cache-dir -r requirements.txt

COPY protocol.py ./
COPY server-udp.py ./

CMD [ "python", "-u", "./server-udp.py" ]