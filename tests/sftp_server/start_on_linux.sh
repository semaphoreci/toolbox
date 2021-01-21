#!/bin/bash

docker build -t sftp_server tests/sftp_server
docker run -p 9000:22 -d sftp_server

sleep 2

cp tests/sftp_server/id_rsa ~/.ssh/semaphore_cache_key
chmod 0600 ~/.ssh/semaphore_cache_key

ssh-keyscan -p 9000 -H localhost >> ~/.ssh/known_hosts

export SEMAPHORE_CACHE_URL=localhost:9000
export SEMAPHORE_CACHE_USERNAME=tester
