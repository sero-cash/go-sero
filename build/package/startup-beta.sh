#!/bin/bash
PACKAGEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
export LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:~/working/czero/lib/
rm -rf ~/working/*
cp -rf ${PACKAGEDIR}/* ~/working/
cd ~/working
RPCPORT=8545
SERVERPORT=60602
RPCADDR=$(hostname -I|awk -F ' ' '{print $1}')
RPCAPI='sero,web3'
DATADIR='~/.datadir'
LOGDIR='../.log'
PATTERN_MAIN_PROCESS="gero.*datadir="
killProcess() {
    if [[ -z $1 ]]; then
        echo "please input the process pattern to kill"
    fi
    if [ $(ps -ef | grep $1 | grep -v grep | awk '{print $2}'|wc -l) -gt 0 ]; then
        echo "to kill process with pattern:$1"
        ps -ef | grep $1 | grep -v grep | awk '{print $2}' | xargs kill -9
    fi

}
if [ ! -d ${LOGDIR} ]; then
	mkdir ${LOGDIR}
fi
killProcess ${PATTERN_MAIN_PROCESS}
sleep 10
nohup ~/working/bin/gero --datadir=${DATADIR} --rpc --rpcport ${RPCPORT} --rpcaddr ${RPCADDR} --rpcapi ${RPCAPI} --port ${SERVERPORT} --rpccorsdomain "*" &> ${LOGDIR}/gero.log &
