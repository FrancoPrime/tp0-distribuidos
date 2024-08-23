#!/bin/sh
source ./config-validador.ini

RESPONSE=$(echo "$TEST_MESSAGE" | nc $SERVER_IP $SERVER_PORT)

if [ "$RESPONSE" = "$TEST_MESSAGE" ]; then
  echo "action: test_echo_server | result: success"
  exit 0
else
  echo "action: test_echo_server | result: fail"
  exit 1
fi