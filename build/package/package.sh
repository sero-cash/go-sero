#!/bin/bash
LOCAL_PATH="$( cd "$( dirname "${BASH_SOURCE[0]}"  )" >/dev/null && pwd  )"
echo $LOCAL_PATH
GERO_BASE=${LOCAL_PATH}/../../
CZERO_PATH=${LOCAL_PATH}/../../../czero
if [ ! -d ${CZERO_PATH} ]; then
    echo "the path ${CZERO_PATH} should be available for build czero"
    echo "please cd ${CZERO_PATH}\n git clone https://gitee.com/hyperspace/czero.git"
    exit
fi
cd ${CZERO_PATH}/build
./depends.sh
cmake ..
make clean
make czero
CZERO_IMPT_PATH=${LOCAL_PATH}/../../../go-czero-import
if [ ! -d ${CZERO_IMPT_PATH} ]; then
    echo "the path ${CZERO_IMPT_PATH} should be available for build czero"
    echo "please cd ${CZERO_IMPT_PATH}\n git clone https://github.com/sero-cash/go-czero-import.git"
    exit
fi
cp ${CZERO_PATH}/lib/libczero.* ${CZERO_IMPT_PATH}/czero/lib/
cd ${GERO_BASE}
make clean all

cd $LOCAL_PATH
if [ -d ./output ]; then
	rm -rf ./output
fi
if [ -f ./output.tar ]; then
	rm ./output.tar
fi
mkdir output
cd ${GERO_BASE}
cp -rf ./build/bin ${LOCAL_PATH}/output
cp -rf ${CZERO_IMPT_PATH}/czero ${LOCAL_PATH}/output
cp ./build/package/*.sh ${LOCAL_PATH}/output 
cd ${LOCAL_PATH}
tar czvf ./output.tar.gz ./output/*
cd -

