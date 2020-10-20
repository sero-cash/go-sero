#!/bin/sh

export DYLD_LIBRARY_PATH="./czero/lib/"

export LD_LIBRARY_PATH="./czero/lib/"

DEFAULT_DATD_DIR="./data"

LOGDIR="./log"

DEFAULT_PORT=53717

CONFIG_PATH="./geroConfig.toml"

DATADIR_OPTION=${DEFAULT_DATD_DIR}

RPC_OPTION="--rpc --rpcport 8545 --rpcapi exchange,sero,net --rpcaddr 127.0.0.1  --rpccorsdomain=*"

cmd="bin/gero --mineMode --config ${CONFIG_PATH} --datadir ${DATADIR_OPTION} --port ${DEFAULT_PORT}  ${RPC_OPTION} --confirmedBlock 32 --rpcwritetimeout 1800 --exchangeValueSt"
mkdir -p $LOGDIR

echo $cmd
current=`date "+%Y-%m-%d"`
logName="posGero_$current.log"
sh stop.sh
nohup ${cmd} >> "${LOGDIR}/${logName}" 2>&1 & echo $! > "./pid"