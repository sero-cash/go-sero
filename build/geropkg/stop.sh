#!/usr/bin/env bash

ROOT=$(cd `dirname $0`; pwd)

if [ -f "${ROOT}/pid" ];then
    echo "kill -9 `cat pid`"
    kill -9 `cat pid`
    rm -rf ${ROOT}/pid
else
    echo "no pid"
fi