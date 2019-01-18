#!/bin/sh

if [ -f "pid" ];then
    echo "kill -9 `cat pid`"
    kill -9 `cat pid`
    rm -rf pid
else
    echo "no pid"
fi
