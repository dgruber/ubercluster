#!/bin/sh

mkdir -p ca
mkdir -p server
mkdir -p client

# create CA
openssl genrsa -aes256 -out ca/ca.key 4096 
chmod 400 ca/ca.key
openssl req -new -x509 -sha256 -days 3650 -key ca/ca.key -out ca/ca.crt
chmod 400 ca/ca.crt

# server
openssl genrsa -out server/server.key 2048
chmod 400 server/server.key
openssl req -new -key server/server.key -sha256 -out server/client.csr

# sign
openssl x509 -req -days 3650 -sha256 -in server/client.csr -CA ca/ca.crt -CAkey ca/ca.key -set_serial 1 -out server/client.crt
chmod 400 server/client.crt

openssl verify -CAfile ca/ca.crt server/client.crt

# client 1
openssl genrsa -out client/client1.key 2048
openssl req -new -key client/client1.key -out client/client1.csr
openssl x509 -req -days 3650 -sha256 -in client/client1.csr -CA ca/ca.crt -CAkey ca/ca.key -set_serial 2 -out client/client1.crt

