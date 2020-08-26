#!/bin/bash

if [ -f "pid" ];then
    kill -9 `cat pid`
    rm -rf pid
fi
