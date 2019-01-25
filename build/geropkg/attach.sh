#!/bin/sh

DATADIR="./data"
if [ ! -z "$1" ]; then
    DATADIR=$1
fi

export DYLD_LIBRARY_PATH="./czero/lib/"
export LD_LIBRARY_PATH="./czero/lib/"

bin/gero --datadir="${DATADIR}" attach