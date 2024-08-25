def receive_message(client_sock):
    remaining = 2
    message_length = b''
    while remaining > 0:
        msg = client_sock.recv(remaining)
        remaining -= len(msg)
        message_length += msg
    remaining = int.from_bytes(message_length, byteorder='big')
    message = b''
    while remaining > 0:
        msg = client_sock.recv(remaining)
        remaining -= len(msg)
        message += msg
    return message.decode('utf-8')

def send_message(client_sock, message):
    message = message.encode('utf-8')
    message_length = len(message).to_bytes(2, byteorder='big')
    client_sock.sendall(message_length)
    client_sock.sendall(message)