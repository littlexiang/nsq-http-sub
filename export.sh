#!/bin/sh
export GOPATH=`pwd`/../..
export PATH="$PATH:$GOPATH/bin:/usr/local/go/bin"

echo "GOPATH:" 
echo $GOPATH
echo "PATH: " 
echo $PATH
