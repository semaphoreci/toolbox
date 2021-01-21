#!/bin/bash

docker build -t sftp-server tests/sftp_server
docker run --net tmp_default --name sftp_server -d sftp-server

sleep 2

cp tests/sftp_server/id_rsa ~/.ssh/semaphore_cache_key
chmod 0600 ~/.ssh/semaphore_cache_key

ssh-keyscan -H sftp_server >> ~/.ssh/known_hosts

export SEMAPHORE_CACHE_URL=sftp_server:22
export SEMAPHORE_CACHE_USERNAME=tester
