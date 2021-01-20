#!/bin/bash

docker build -t sftp-server $(dirname $0)
docker run -p 9000:22 -d sftp-server

sleep 2
