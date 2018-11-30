#!/usr/bin/env bash

ROOT=$(cd `dirname $0`; pwd)
nohup ${ROOT}/startup.sh --alpha &> ${ROOT}/log/gero.log & echo $! > ${ROOT}/pid