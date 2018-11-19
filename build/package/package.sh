#!/bin/bash
LOCAL_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
echo $LOCAL_PATH
GERO_BASE=${LOCAL_PATH}/../../
#CZERO_PATH=${LOCAL_PATH}/../../../czero
#compileCZero() {
#    if [ ! -d ${CZERO_PATH} ]; then
#        echo "the path ${CZERO_PATH} should be available for build czero"
#        echo "please cd ${CZERO_PATH}\n git clone https://gitee.com/hyperspace/czero.git"
#        exit
#    fi
#    cd ${CZERO_PATH}/build
#    ./depends.sh
#    cmake ..
#    make clean
#    make czero
#}
#if [ -z "$1" ]; then
#    echo "No argument supplied"
#elif [ "x$1" == "xcompile" ]; then
#    compileCZero
#fi
CZERO_IMPT_PATH=${LOCAL_PATH}/../../../go-czero-import
if [ ! -d ${CZERO_IMPT_PATH} ]; then
    echo "the path ${CZERO_IMPT_PATH} should be available for build czero"
    echo "please cd ${GERO_BASE}/..\n git clone https://github.com/sero-cash/go-czero-import.git"
    exit
fi
cp ${CZERO_PATH}/lib/libczero.* ${CZERO_IMPT_PATH}/czero/lib/
cd ${GERO_BASE}
make clean all

cd $LOCAL_PATH
if [ -d ./geropkg ]; then
	rm -rf ./geropkg
fi
if [ -f ./geropkg.tar.gz ]; then
	rm ./geropkg.tar.gz
fi
mkdir geropkg
cd ${GERO_BASE}
cp -rf ./build/bin ${LOCAL_PATH}/geropkg
cp -rf ${CZERO_IMPT_PATH}/czero ${LOCAL_PATH}/geropkg
cp ./build/package/*.sh ${LOCAL_PATH}/geropkg 
cd ${LOCAL_PATH}
tar czvf ./geropkg.tar.gz ./geropkg/*
cd -
echo "${LOCAL_PATH}/geropkg.tar.gz is available now"
