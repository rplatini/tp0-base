import socket
import logging
import signal

from common.utils import deserialize, has_won, store_bets, load_bets
from common.message_handler import MessageHandler

DELIMITER = ','
AGENCIES = 5

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self._running = True
        self._client_connections = {}
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
                self._client_connections[messageHandler.get_address()] = messageHandler
                self.__handle_client_connection(messageHandler)

                if len(self._client_connections) == AGENCIES:
                    self.__start_lottery()
                    self.__close_client_connections() 

    def __graceful_shutdown(self, _signum, _frame):
        """
        Graceful shutdown of the server
        """
        logging.debug(f"action: shutdown | result: in_progress")
        self._running = False

        self._server_socket.close()
        logging.debug("action: exit | result: success")
 
        
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

        except OSError:
            logging.error(f'action: apuesta_recibida | result: fail | cantidad: ${len(bets)}')

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
        
    def __close_client_connections(self):
        """
        Close all client connections

        Function iterates over all client connections and closes them
        """
        for client in self._client_connections.values():
            client.close()
        
        self._client_connections.clear()

    def __start_lottery(self):
        winners = {}
        bets = load_bets()

        for bet in bets:
            if bet.agency not in winners:
                    winners[bet.agency] = []
            if has_won(bet):
                winners[bet.agency].append(bet.document)

        logging.info('action: sorteo | result: success')

        for client in self._client_connections.values():
                self.__send_winners(client, winners)
    

    def __send_winners(self, messageHandler: MessageHandler, winners: dict):
        try:
            _, winnersAsk = messageHandler.receive_message()
            if not winnersAsk:
                return
        
            agency = int(winnersAsk.split(DELIMITER)[1])
            
            winnersResponse = DELIMITER.join(winners[agency])
            messageHandler.send_message(winnersResponse, True)

            logging.debug(f'action: send_winners | result: success | Agency [{agency}] Dni winners: {winnersResponse}')

        except Exception as e:
            logging.error(f'action: send_winners | result: fail | error: {e}')
