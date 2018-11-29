#!/usr/bin/env bash

ROOT=$(cd `dirname $0`; pwd)
LOGDIR="${ROOT%/*}/log"
nohup ./startup-beta.sh &> ${LOGDIR}/gero.log 2>&1& echo $! ${ROOT}/pid