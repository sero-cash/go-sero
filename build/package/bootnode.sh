#!/bin/bash
PATTERN_MAIN_PROCESS="bootnode.*nodekey.*60609.*"
LOGDIR='../.log'
killProcess() {
    if [[ -z $1 ]]; then
        echo "please input the process pattern to kill"
    fi
    if [ $(ps -ef | grep $1 | grep -v grep | awk '{print $2}'|wc -l) -gt 0 ]; then
        echo "to kill process with pattern:$1"
        ps -ef | grep $1 | grep -v grep | awk '{print $2}' | xargs kill -9
    fi

}
killProcess ${PATTERN_MAIN_PROCESS}
if [ -f ~/.bootnode/bootnode.key.* ]; then
	if [ -f ./bootnode.key ]; then
        rm ./bootnode.key
    fi
    ln -s ~/.bootnode/bootnode.key.* bootnode.key
    if [ ! -d ${LOGDIR}} ]; then
    	echo "now create ${LOGDIR}"
    	mkdir -p log
    	mv log ${LOGDIR}
    fi
fi
export LD_LIBRARY_PATH=`pwd`/czero/lib:${LD_LIBRARY_PATH}
nohup ./bin/bootnode -nodekey=bootnode.key -verbosity=9 -addr :60609 -nat=any &>${LOGDIR}/bootnode.log &
