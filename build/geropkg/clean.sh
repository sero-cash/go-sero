#!/bin/bash

ROOT=$(cd `dirname $0`; pwd)

DATADIR="${ROOT}/data"
if [ ! -z "$1" ]; then
    DATADIR=$1
fi

sh ${ROOT}/stop.sh

echo "rm -rf ${DATADIR}/gero"
rm -rf ${DATADIR}/gero
echo "rm -rf ${DATADIR}/gero.ipc"
rm -rf ${DATADIR}/gero.ipc
echo "rm -rf ${DATADIR}/state1"
rm -rf ${DATADIR}/state1