#!/bin/sh

ROOT=$(cd `dirname $0`; pwd)
DATADIR="${ROOT}/data"
if [ ! -z "$1" ]; then
    DATADIR=$1
fi

export DYLD_LIBRARY_PATH=${ROOT}/czero/lib/
export LD_LIBRARY_PATH=${ROOT}/czero/lib/

${ROOT}/bin/gero --datadir="${DATADIR}" attach

