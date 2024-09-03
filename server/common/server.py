import socket
import logging
import signal

from common.utils import deserialize, store_bets, load_bets
from common.message_handler import MessageHandler

ACK_MESSAGE = "ACK"

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        signal.signal(signal.SIGTERM, self.__graceful_shutdown)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self._running:
            client_sock = self.__accept_new_connection()

            if client_sock:
                messageHandler = MessageHandler(client_sock, client_sock.getpeername())
                self.__handle_client_connection(messageHandler)  
        
    def __handle_client_connection(self, messageHandler: MessageHandler):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            end_flag = False

            while not end_flag:
                end_flag, msg = messageHandler.receive_message()

                bets = deserialize(msg)
                store_bets(bets)

                logging.info(f'action: apuesta_recibida | result: success | cantidad: ${len(bets)}')

            messageHandler.send_message(ACK_MESSAGE)

        except OSError:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: ${len(bets)}')

        except RuntimeError:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: ${len(bets)}')

        finally:
            messageHandler.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        
        except OSError as e:
            if self._running:
                logging.error(f"action: accept_connections | result: fail | error: {e}")
            
            return None

    
    def __graceful_shutdown(self, _signum, _frame):
        logging.debug(f"action: shutdown | result: in_progress")
        self._running = False

        logging.debug("action: closing server socket | result: in_progress")
        self._server_socket.close()
        logging.debug("action: exit | result: success")
