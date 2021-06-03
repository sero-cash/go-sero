#!/bin/sh

if [ -f "pid" ];then
    kill -INT `cat pid`
    rm -rf pid
fi
