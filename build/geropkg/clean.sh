#!/usr/bin/env bash

ROOT=$(cd `dirname $0`; pwd)
echo "rm -rf ${ROOT}/gero"
rm -rf ${ROOT}/gero
echo "rm -rf ${ROOT}/gero.ipc"
rm -rf ${ROOT}/gero.ipc
echo "rm -rf ${ROOT}/state1"
rm -rf ${ROOT}/state1