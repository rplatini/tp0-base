PACKET_SIZE = 4
FLAG_SIZE = 1

class MessageHandler():
    def __init__(self, socket, addr):
        self.socket = socket
        self.addr = addr

    def get_address(self):
        return self.addr[0]

    def send_message(self, msg):
        try:
            msg_bytes = msg.encode('utf-8')

            size_bytes = len(msg_bytes).to_bytes(PACKET_SIZE, byteorder='big')
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

    def receive_message(self) -> (bool, str):
        try:
            size_bytes = self.socket.recv(PACKET_SIZE)
            if len(size_bytes) < PACKET_SIZE:
                raise RuntimeError("Socket connection broken")
                     
            size = int.from_bytes(size_bytes, byteorder='big')
            # print('Total bytes received: ', size)

            end_flag_bytes = self.socket.recv(FLAG_SIZE)
            if len(end_flag_bytes) < FLAG_SIZE:
                raise RuntimeError("Socket connection broken")
            
            end_flag = int.from_bytes(end_flag_bytes, byteorder='big')

            msg = self.read_message(size)
            if msg is None:
                return False, "Error: Message read failed."

            return end_flag, msg

        except OSError as e:
            return e
        
    def read_message(self, size):
        full_message = self.socket.recv(size)
        totalRead = len(full_message)

        while totalRead < size:
            msg = self.socket.recv(size - totalRead)
            full_message += msg
            totalRead += len(msg)

        return full_message.rstrip().decode('utf-8')
        
    def close(self):
        self.socket.close()