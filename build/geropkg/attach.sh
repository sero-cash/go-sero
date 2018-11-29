#!/bin/bash
PACKAGEDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
export LD_LIBRARY_PATH=${PACKAGEDIR}/czero/lib/:${LD_LIBRARY_PATH}
DATADIR="${PACKAGEDIR}/../datadir"
${PACKAGEDIR}/bin/gero --datadir=${DATADIR} attach

