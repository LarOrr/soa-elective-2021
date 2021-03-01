import enum


class DataType(enum.Enum):
    ClientData = 1
    Handshake = 2
    Disconnect = 3
    Request = 4
    Notification = 5


class Protocol:
    PORT = 4444
    CLIENT_DATA_MIN = 0
    CLIENT_DATA_MAX = 50
    HANDSHAKE = CLIENT_DATA_MAX + 1
    DISCONNECT = CLIENT_DATA_MAX + 2
    REQUEST = CLIENT_DATA_MAX + 3
    NOTIFICATION = CLIENT_DATA_MAX + 4

    typeToOrd = {DataType.ClientData: CLIENT_DATA_MIN, DataType.Handshake: HANDSHAKE, DataType.Disconnect: DISCONNECT,
                 DataType.Request: REQUEST, DataType.Notification: NOTIFICATION}
    ordToType = {v: k for k, v in typeToOrd.items()}

    def __init__(self, dataType=None, head=None, data=None, datapacket=None):
        if dataType is not None:
            self.head = Protocol.typeToOrd[dataType]
        else:
            self.head = datapacket[0] if head is None else head
        if data is None and datapacket is not None:
            self.data = datapacket[1:]
        else:
            self.data = data
        self.DataType = Protocol.getDataType(self.head)

    @staticmethod
    def getDataType(head):
        if head <= Protocol.CLIENT_DATA_MAX and head >= Protocol.CLIENT_DATA_MIN:
            return DataType.ClientData
        try:
            return Protocol.ordToType[head]
        except:
            return None

    def out(self):
        bytearr = bytearray(b'')
        bytearr.append(self.head)
        if self.data is None:
            return bytes(bytearr)
        else:
            return bytes(bytearr + self.data)
