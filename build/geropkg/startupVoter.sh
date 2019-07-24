#!/bin/sh
show_usage="args: [-d ,-k, -p, -n,-r,-h]\
                                  [--datadir=,--keystore=, --port=, --net=, --rpc=,--help]"
export DYLD_LIBRARY_PATH="./czero/lib/"
export LD_LIBRARY_PATH="./czero/lib/"
DEFAULT_DATD_DIR="./data"
LOGDIR="./log"
DEFAULT_PORT=53717
CONFIG_PATH="./geroConfig.toml"
DATADIR_OPTION=${DEFAULT_DATD_DIR}
KEYSTORE_OPTION=""
VOTER=""
VOTER_PASSWORD_PATH=""


GETOPT_ARGS=`getopt -o d:k:p:v:f:h -al datadir:,keystore:,port:,voter:,pf:,help -- "$@"`
eval set -- "$GETOPT_ARGS"
while [ -n "$1" ]
do
        case "$1" in
                -d|--datadir) DATADIR_OPTION=$2; shift 2;;
                -p|--port) DEFAULT_PORT=$2; shift 2;;
                -v|--voter) VOTER=$2; shift 2;;
                -k|--keystore) KEYSTORE_OPTION="--keystore $2"; shift 2;;
                -f|--pf) VOTER_PASSWORD_PATH=$2; shift 2;;
                -h|--help) echo $show_usage;exit 0;;
                --) break ;;
        esac
done

cmd="bin/gero --mineMode --config ${CONFIG_PATH} --unlock ${VOTER} --password ${VOTER_PASSWORD_PATH} --datadir ${DATADIR_OPTION} --port ${DEFAULT_PORT} ${KEYSTORE_OPTION}"
mkdir -p $LOGDIR

echo $cmd
current=`date "+%Y-%m-%d"`
logName="posGero_$current.log"
sh stop.sh
nohup ${cmd} >> "${LOGDIR}/${logName}" 2>&1 & echo $! > "./pid"