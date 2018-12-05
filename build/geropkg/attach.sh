#!/bin/sh

ROOT=$(cd `dirname $0`; pwd)
DATADIR="${ROOT}/data"
if [ ! -z "$1" ]; then
    DATADIR=$1
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
    export DYLD_LIBRARY_PATH=${ROOT}/czero/lib/
    echo $DYLD_LIBRARY_PATH
else
    export LD_LIBRARY_PATH=${ROOT}/czero/lib/
    echo $LD_LIBRARY_PATH
fi

${ROOT}/bin/gero --datadir=${DATADIR} attach

