#!/bin/bash
PACKAGEDIR=$(cd `dirname $0`; pwd)
RPCPORT=8545
SERVERPORT=60602
LOCALHOST=`/sbin/ifconfig -a|grep inet|grep -v 127.0.0.1|grep -v inet6|awk '{print $2}'|tr -d "addr:"`
RPCAPI='sero,web3'
DATADIR="${PACKAGEDIR%/*}/datadir"


if [ -f "pid" ];then
    kill -9 `cat pid`
fi

function sysname() {

    SYSTEM=`uname -s`
    if [[ "Darwin" == "$SYSTEM" ]]
    then
        echo "Darwin"
    fi

    if [[ "Linux" == "$SYSTEM" ]]
    then
        name=`cat /etc/system-release|awk '{print $1}'`
        echo "Linux $name"
    fi
}

SNAME=`sysname`
if [[ "Darwin" = "$SNAME" ]];then
    export DYLD_LIBRARY_PATH=${PACKAGEDIR}/czero/lib/
    echo $DYLD_LIBRARY_PATH
else
    export LD_LIBRARY_PATH=${PACKAGEDIR}/czero/lib/
    echo $LD_LIBRARY_PATH
fi


sleep 10
${PACKAGEDIR}/bin/gero --datadir=${DATADIR} --rpc --rpcport ${RPCPORT} --rpcaddr ${LOCALHOST} --rpcapi ${RPCAPI} --port ${SERVERPORT} --rpccorsdomain "*"
