#!/bin/sh

ROOT=$(cd `dirname $0`; pwd)

DATADIR="${ROOT}/data"
if [ ! -z "$1" ]; then
    DATADIR=$1
fi

sh ${ROOT}/stop.sh

echo "rm -rf ${DATADIR}/gero/chaindata"
rm -rf ${DATADIR}/gero/chaindata
echo "rm -rf ${DATADIR}/gero.ipc"
rm -rf ${DATADIR}/gero.ipc
echo "rm -rf ${DATADIR}/balance"
rm -rf ${DATADIR}/balance
echo "rm -rf ${DATADIR}/exchange"
rm -rf ${DATADIR}/exchange
echo "rm -rf ${DATADIR}/stake"
rm -rf ${DATADIR}/stake
echo "rm -rf ${DATADIR}/light"
rm -rf ${DATADIR}/light
