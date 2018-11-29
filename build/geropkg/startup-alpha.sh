#!/bin/bash
PACKAGEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
export LD_LIBRARY_PATH=${PACKAGEDIR}/czero/lib/:${LD_LIBRARY_PATH}
RPCPORT=8545
SERVERPORT=60602
RPCADDR=$(hostname -I|awk -F ' ' '{print $1}')
RPCAPI='sero,web3'
DATADIR="${PACKAGEDIR}/../datadir"
LOGDIR="${PACKAGEDIR}/../log"
PATTERN_MAIN_PROCESS="gero.*datadir="

kill -9 `cat pid`

sleep 10
nohup ${PACKAGEDIR}/bin/gero --alpha --datadir=${DATADIR} --rpc --rpcport ${RPCPORT} --rpcaddr ${RPCADDR} --rpcapi ${RPCAPI} --port ${SERVERPORT} --rpccorsdomain "*" &> ${LOGDIR}/gero.log 2>&1& echo $! pid
