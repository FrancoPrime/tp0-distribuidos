import socket
import logging
import os
import multiprocessing
from .communication import receive_message, send_message
from .utils import Bet, store_bets, load_bets, has_won, serialize_winners

SuccessMessage = "success"
ExitMessage = "exit"
ErrorMessage = "error"
WinnersMessage = "winners"

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)

        manager = multiprocessing.Manager()
        self.bets_lock = manager.Lock()

        self.running = True
        self.shared_data = manager.dict({
            'winners': [],
            'agencies': {f"{i}": False for i in range(1, int(os.getenv('AGENCIES', 0)) + 1)}
        })

    def run(self):
        """
        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        processes = []
        while self.running:
            client_sock = self.__accept_new_connection()
            if not self.running:
                logging.info('action: stop_server | result: success')
                break
            proc = multiprocessing.Process(target = self.__handle_client_connection, args = (client_sock,))
            proc.start()
            processes.append(proc)
        for proc in processes:
            proc.join()

    def __check_winners(self):
        if all(value == True for value in self.shared_data['agencies'].values()):
            logging.info('action: sorteo | result: success')
            bets = list(load_bets())
            self.shared_data['winners'] = [bet for bet in bets if has_won(bet)]

    def __handle_client_connection(self, client_sock):
        """
        Read message from a specific client socket and loops through their messages until receiving an exit message

        If a problem arises in the communication with the client, the
        client socket will also be closed.
        """
        try:
            while self.running:
                msg = receive_message(client_sock)

                if msg.startswith(ExitMessage):
                    logging.debug('Agency finished')
                    with self.bets_lock:
                        agencies = self.shared_data['agencies']
                        agencies[msg[len(ExitMessage):]] = True
                        self.shared_data['agencies'] = agencies
                        self.__check_winners()
                    break
                
                if msg.startswith(WinnersMessage):
                    with self.bets_lock:
                        self.__process_winners_message(client_sock, msg[len(WinnersMessage):])
                    return
                
                bets = Bet.fromStr(msg)
                with self.bets_lock:
                    store_bets(bets)
                logging.info(f'action: apuesta_recibida | result: success | cantidad: {len(bets)}')

                send_message(client_sock, SuccessMessage)
        except ConnectionResetError:
            logging.debug('Client disconnected')
        except Exception as e:
            logging.info(f'action: apuesta_recibida | result: fail | cantidad: 0')
            send_message(client_sock, ErrorMessage)
        finally:
            client_sock.close()

    def __process_winners_message(self, client_sock, agencyID):
        if not all(value == True for value in self.shared_data['agencies'].values()):
            logging.info('action: winners_request | result: fail | error: agencies not finished')
            send_message(client_sock, ErrorMessage)
            return
        agency_winners = [bet for bet in self.shared_data['winners'] if bet.agency == int(agencyID)]
        send_message(client_sock, serialize_winners(agency_winners))

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