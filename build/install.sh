#! /bin/sh

_GOPATH=`cd ../../../../../;pwd`

export GOPATH=$_GOPATH
echo $GOPATH

go install -v ../cmd/gero
go install -v ../cmd/serokey
