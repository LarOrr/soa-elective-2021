# Voice chat based on UDP Sockets in Python

Based on: 
https://github.com/domage/soa-curriculum-2021/tree/main/examples/sockets-voice-chat

and

https://github.com/TomPrograms/Python-Voice-Chat

Run client with:
1) pip install requirements.txt
2) python client-udp.py

Then enter address of server

----------------------------------------

Server listens at port 4444
Container: https://hub.docker.com/repository/docker/larorr/soa-voice-chat

To run container:
docker run -p 4444:4444/udp larorr/soa-voice-chat

Enter 127.0.0.1 in client as address
