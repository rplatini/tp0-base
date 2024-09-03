FROM ubuntu:latest

RUN apt-get update && apt-get install -y netcat
CMD echo "testing my server" | nc server 12345