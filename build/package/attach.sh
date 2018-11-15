#!/bin/bash
LOCAL_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
export LD_LIBRARY_PATH=${LD_LIBRARY_PATH}:${LOCAL_PATH}/czero/lib/
DATADIR=${LOCAL_PATH}/../datadir
${LOCAL_PATH}/bin/gero --datadir=${DATADIR} attach
