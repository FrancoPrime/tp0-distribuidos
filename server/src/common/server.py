import socket
import logging
from .communication import receive_message, send_message
from .utils import Bet, store_bets

SuccessMessage = "success"
ExitMessage = "exit"
ErrorMessage = "error"

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.running = True

    def run(self):
        """
        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        while self.running:
            client_sock = self.__accept_new_connection()
            if not self.running:
                logging.info('action: stop_server | result: success')
                break
            self.__handle_client_connection(client_sock)

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and loops through their messages until receiving an exit message

        If a problem arises in the communication with the client, the
        client socket will also be closed.
        """
        try:
            while self.running:
                msg = receive_message(client_sock)

                if msg == ExitMessage:
                    logging.info('Agency finished')
                    break
                
                bets = Bet.fromStr(msg)
                store_bets(bets)

                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

                send_message(client_sock, SuccessMessage)
        except ConnectionResetError:
            logging.info('Client disconnected')
        except Exception as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: 0')
            send_message(client_sock, ErrorMessage)
        finally:
            client_sock.close()

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
        except Exception as e:
            logging.error(f'action: accept_connections | result: fail | error: {e}')
            return None
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c

    def stop(self):
        """
        Stop the server
        """

        logging.info('action: stop_server | result: in_progress')
        self.running = False
        logging.info('action: closing socket | result: in_progress')
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()
        logging.info('action: closing socket | result: success')