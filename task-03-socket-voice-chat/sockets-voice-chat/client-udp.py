#!/usr/bin/python3
import copy
import socket
import threading
import pyaudio
from protocol import DataType, Protocol


class Client:
    def __init__(self):
        self.socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        self.bufferSize = 4096
        self.connected = False
        # По умолчанию выключено
        self.send_on = False
        self.name = "111" # input('Enter your name --> ')

        while 1:
            try:
                self.target_ip = input('Enter IP address of server --> ')  # "172.22.224.1"
                self.target_port = Protocol.PORT # int(input('Enter target port of server --> '))
                self.server = (self.target_ip, self.target_port)
                self.connect_to_server()
                break
            except ():
                print("Couldn't connect to server...")

        chunk_size = 1024
        audio_format = pyaudio.paInt16
        channels = 1
        rate = 20000

        # initialise microphone recording
        self.p = pyaudio.PyAudio()
        self.playing_stream = self.p.open(format=audio_format, channels=channels, rate=rate, output=True,
                                          frames_per_buffer=chunk_size)
        self.recording_stream = self.p.open(format=audio_format, channels=channels, rate=rate, input=True,
                                            frames_per_buffer=chunk_size)

        # start threads
        # receive_thread
        threading.Thread(target=self.receive_server_data, daemon=True).start()
        # send_thread
        # threading.Thread(target=self.send_data_to_server, daemon=True).start() # .start()
        # send_thread.start()
        print("---------------------------------")
        print("Enter 'exit' to exit the chat")
        print("Enter 'on'/'off' to turn on/off the micro")
        print("Enter 'list' to see list of users")
        print("---------------------------------")
        while 1:
            command = input()
            if command == 'exit':
                message = Protocol(dataType=DataType.Disconnect)
                self.socket.sendto(message.out(), self.server)
                break
            elif command == 'off':
                self.send_on = False
            elif command == 'on':
                self.send_on = True
                threading.Thread(target=self.send_data_to_server, daemon=True).start()
            elif command == 'list':
                message = Protocol(dataType=DataType.Request, data='list'.encode(encoding='UTF-8'))
                self.socket.sendto(message.out(), self.server)
            else:
                print("Command '{}' is incorrect!".format(command))


    def receive_server_data(self):
        while self.connected:
            try:
                # 2049 - buffer size
                data, addr = self.socket.recvfrom(2049)
                message = Protocol(datapacket=data)
                if message.DataType == DataType.ClientData:
                    self.playing_stream.write(message.data)
                #     CAN be changed to notification
                elif message.DataType == DataType.Disconnect:
                    print("User {} has disconnected".format(message.data.decode(encoding='UTF-8')))
                elif message.DataType == DataType.Notification:
                    print(message.data.decode(encoding='UTF-8'))
                else:
                    print("WARNING: Received unsupported data type")
            except:
                pass

    def connect_to_server(self):
        if self.connected:
            return True

        message = Protocol(dataType=DataType.Handshake, data=self.name.encode(encoding='UTF-8'))
        self.socket.sendto(message.out(), self.server)

        try:
            data, addr = self.socket.recvfrom(2049)
        except:
            print("Can't connect to the server, try again!")
            exit()
        datapack = Protocol(datapacket=data)

        if (addr == self.server and datapack.DataType == DataType.Handshake and
                datapack.data.decode('UTF-8') == 'ok'):
            print('Connected to server successfully!')
            self.connected = True
        return self.connected

    def send_data_to_server(self):
        while self.connected and self.send_on:
            try:
                data = self.recording_stream.read(1024)
                message = Protocol(dataType=DataType.ClientData, data=data)
                self.socket.sendto(message.out(), self.server)
            except:
                pass


client = Client()
