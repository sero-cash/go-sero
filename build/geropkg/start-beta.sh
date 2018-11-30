#!/usr/bin/env bash

ROOT=$(cd `dirname $0`; pwd)
nohup ${ROOT}/startup.sh &> ${ROOT}/log/gero.log & echo $! > ${ROOT}/pid