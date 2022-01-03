#!/bin/bash

BRANCH=`git rev-parse --abbrev-ref HEAD | sed -r 's/\/+/-/g'`

# stop
sudo docker stop ttrack_api
sudo docker rm ttrack_api

# build
sudo docker build -t rebel1l/ttrack_api:$BRANCH .

# start
sudo docker run --name ttrack_api -d -it -p 3000:3000 rebel1l/ttrack_api:$BRANCH
sudo docker ps
