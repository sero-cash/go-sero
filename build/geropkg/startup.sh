#!/bin/sh
show_usage="args: [-d , -p, -n,-r,-h]\
                                  [--datadir=, --port=, --net=, --rpc=,--help]"
export DYLD_LIBRARY_PATH="./czero/lib/"
export LD_LIBRARY_PATH="./czero/lib/"
DEFAULT_DATD_DIR="./data"
LOGDIR="./log"
DEFAULT_PORT=53717

DATADIR_OPTION=${DEFAULT_DATD_DIR}
NET_OPTION=""
RPC_OPTION=""
PORT_OPTION=${DEFAULT_PORT}


GETOPT_ARGS=`getopt -o d:p:n:r:h -al datadir:,port:,net:,rpc:,help -- "$@"`
eval set -- "$GETOPT_ARGS"
while [ -n "$1" ]
do
        case "$1" in
                -d|--datadir) DATADIR_OPTION=$2; shift 2;;
                -p|--port) PORT_OPTION=$2; shift 2;;
                -n|--net) NET_OPTION=--$2; shift 2;;
                -r|--rpc)
                        localhost=$(hostname -I|awk -F ' ' '{print $1}')
                        RPC_OPTION="$cmd --rpc --rpcport $2 --rpcaddr $localhost --rpcapi personal,sero,web3 --rpccorsdomain '*'"; shift 2;;
                -h|--help) echo $show_usage exit 0;;
                --) break ;;
        esac
done

cmd="bin/gero --datadir ${DATADIR_OPTION} --port ${PORT_OPTION} ${NET_OPTION} ${RPC_OPTION}"
mkdir -p $LOGDIR

echo $cmd
current=`date "+%Y-%m-%d"`
logName="gero_$current.log"
sh stop.sh
nohup ${cmd} >> "${LOGDIR}/${logName}" & echo $! > "./pid"