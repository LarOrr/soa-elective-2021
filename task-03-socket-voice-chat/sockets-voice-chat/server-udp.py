#!/usr/bin/python3

import socket
import threading
from protocol import DataType, Protocol


class Server:
    def __init__(self):
        self.ip = socket.gethostbyname(socket.gethostname())
        self.id_counter = 1
        while 1:
            try:
                self.port = Protocol.PORT  ## 4444  # int(input('Enter port number to run on --> '))

                self.s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
                self.s.settimeout(5)
                self.s.bind((self.ip, self.port))

                break
            except:
                print("Couldn't bind to that port")

        self.clients = {}
        self.clientCharId = {}
        threading.Thread(target=self.receiveData).start()

    def receiveData(self):
        print('Running on IP: ' + self.ip)
        print('Running on port: ' + str(self.port))

        while True:
            try:
                data, addr = self.s.recvfrom(4096)
                message = Protocol(datapacket=data)
                self.handleMessage(message, addr)
            except socket.timeout:
                pass

    def handleMessage(self, message, addr):
        if self.clients.get(addr, None) is None:
            try:
                if message.DataType != DataType.Handshake:
                    return

                name = message.data.decode(encoding='UTF-8')

                self.clients[addr] = name
                self.clientCharId[addr] = self.id_counter  # len(self.clients)
                self.id_counter += 1

                print('{} has connected on {}!'.format(name, addr))
                ret = Protocol(dataType=DataType.Handshake, data='ok'.encode(encoding='UTF-8'))
                self.s.sendto(ret.out(), addr)
                self.broadcast(addr, Protocol(dataType=DataType.Notification,
                                              data="User {} has connected!".format(name).encode(encoding='UTF-8')),
                               change_header=False)
            except:
                print("Handshake error with {}!".format(addr))
            return

        if message.DataType == DataType.ClientData:
            self.broadcast(addr, message)
        elif message.DataType == DataType.Disconnect:
            name = self.clients[addr]
            print("Client {} has disconnected".format(name))
            message.data = name.encode(encoding='UTF-8')
            self.broadcast(addr, message, False)
            self.clients.pop(addr)
        elif message.DataType == DataType.Request:
            data = message.data.decode(encoding='UTF-8')
            if data == 'list':
                message = Protocol(
                    data="List of users:\n{}".format(str(list(self.clients.values()))).encode(encoding='UTF-8'),
                    dataType=DataType.Notification)
                self.s.sendto(message.out(), addr)
            else:
                print("Unsupported request")

    def broadcast(self, sentFrom, data, change_header=True):
        if change_header:
            data.head = self.clientCharId[sentFrom]
        for client in self.clients:
            if client != sentFrom:
                # try:
                self.s.sendto(data.out(), client)
                # except (ex):
                #     print("Problem with sending message to " + self.clients[client])


server = Server()
