#!/bin/sh

echo "Compile binary"
env GOOS=linux GOARCH=arm GOARM=5 go build
echo "Copy binary"
scp ./apms pi@192.168.0.199:~/
echo "Copy service files"
scp ./apms.service pi@192.168.0.199:~/