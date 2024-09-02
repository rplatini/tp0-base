#!/bin/bash
MESSAGE="testing my server"
NETWORK_NAME="tp0_testing_net"
IMAGE_NAME="example"

docker build -t "$IMAGE_NAME" .
RESPONSE=$(docker run --rm --network=$NETWORK_NAME $IMAGE_NAME)
echo "response: $RESPONSE"

if [ "$RESPONSE" == "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi