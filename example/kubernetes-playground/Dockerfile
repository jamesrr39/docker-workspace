FROM ubuntu:16.04

RUN apt-get update && apt-get install -y git sudo wget unzip

RUN mkdir -p /opt/resources

# atom setup
#RUN apt-get install -y libxss1 libasound2 libxkbfile1
#RUN wget -O /opt/resources/atom.deb https://atom.io/download/deb && apt-get install -y -f /opt/resources/atom.deb

RUN adduser --disabled-password --gecos "" user && usermod -aG sudo user && echo "user\nuser\n" | passwd user

WORKDIR /home/user

USER user
