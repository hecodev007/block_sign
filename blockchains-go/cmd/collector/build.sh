#!/bin/bash

uNames=`uname -s`
osName=${uNames: 0: 4}
if [ "$osName" == "Darw" ] # Darwin
then
	echo "Mac OS X"
	set -x
  CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w"
elif [ "$osName" == "Linu" ] # Linux
then
	echo "GNU/Linux"
	set -x
  CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w"
else
	echo "unknown os"
fi
chmod +x collector