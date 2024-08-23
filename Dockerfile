FROM alpine:latest

RUN apk add --no-cache netcat-openbsd
COPY validar-echo-server.sh /validar-echo-server.sh
COPY config-validador.ini /config-validador.ini
RUN chmod +x /validar-echo-server.sh

CMD ["/validar-echo-server.sh"]