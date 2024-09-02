#!/bin/sh
. ./config-validador.ini

RESPONSE=$(sudo docker run --rm --network "$NETWORK_NAME" busybox sh -c "echo '$TEST_MESSAGE' | nc $SERVER_IP $SERVER_PORT")

if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
  echo "action: test_echo_server | result: success"
  exit 0
else
  echo "action: test_echo_server | result: fail"
  exit 1
fi