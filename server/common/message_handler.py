HEADER_SIZE = 4

class MessageHandler():
    def __init__(self, socket, addr):
        self.socket = socket
        self.addr = addr

    def get_address(self):
        return self.addr[0]

    def send_message(self, msg):
        try:
            msg_bytes = msg.encode('utf-8')

            size_bytes = len(msg_bytes).to_bytes(HEADER_SIZE, byteorder='big')
            full_message = size_bytes + msg_bytes

            total_size = len(full_message)
            total_sent = 0

            while total_sent < total_size:
                sent = self.socket.send(full_message[total_sent:])
                if sent == 0:
                    raise RuntimeError("Socket connection broken")

                total_sent += sent

        except OSError as e:
            return e
        
        except RuntimeError as e:
            return e

    def receive_message(self) -> str:
        try:
            size_bytes = self.socket.recv(HEADER_SIZE)
            if len(size_bytes) < HEADER_SIZE:
                return None
                     
            size = int.from_bytes(size_bytes, byteorder='big')
            print('Header size: ', size)

            msg = self.socket.recv(size).rstrip().decode('utf-8')
            return msg

        except OSError as e:
            return e
        
    def close(self):
        self.socket.close()