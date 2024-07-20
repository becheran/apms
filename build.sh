#!/bin/sh

env GOOS=linux GOARCH=arm GOARM=5 go build
scp ./apms pi@192.168.0.199:~/
scp ./apms.service pi@192.168.0.199:~/