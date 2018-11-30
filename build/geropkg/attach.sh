#!/bin/bash

ROOT=$(cd `dirname $0`; pwd)
DATADIR="${ROOT}/data"
if [ ! -z "$1" ]; then
    DATADIR=$1
fi
${ROOT}/bin/gero --datadir=${DATADIR} attach

