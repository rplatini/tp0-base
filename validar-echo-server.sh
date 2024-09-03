#!/bin/bash
MESSAGE="testing my server"
NETWORK_NAME="tp0_testing_net"
IMAGE_NAME="busybox"

# docker build -t "$IMAGE_NAME" .
RESPONSE=$(docker run --rm --network=$NETWORK_NAME $IMAGE_NAME sh -c "echo testing my server | nc -w 3 server 12345")
python3 -c "print('action: test_echo_server | result: success' if '$RESPONSE' == '$MESSAGE' else 'action: test_echo_server | result: fail')"

# if [ "$RESPONSE" == "$MESSAGE" ]; then
#     echo "action: test_echo_server | result: success"
# else
#     echo "action: test_echo_server | result: fail"
# fi