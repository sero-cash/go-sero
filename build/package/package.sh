#!/bin/bash
LOCAL_PATH=`pwd`
echo $LOCAL_PATH
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
cp ${CZERO_PATH}/lib/libczero.* $LOCAL_PATH/github.com/sero-cash/go-czero-import/czero/lib/
cd $LOCAL_PATH/github.com/sero-cash/go-szero
make clean all

cd $LOCAL_PATH
if [ -d ./output ]; then
	rm -rf ./output
fi
if [ -f ./output.tar ]; then
	rm ./output.tar
fi
mkdir output
cp -rf $LOCAL_PATH/github.com/sero-cash/go-szero/build/bin ${LOCAL_PATH}/output
cp -rf $LOCAL_PATH/github.com/sero-cash/go-czero-import/czero ${LOCAL_PATH}/output
cp $LOCAL_PATH/github.com/sero-cash/go-szero/build/mine.sh ~/geroscripts/
cp -rf ~/gerolibs/* ${LOCAL_PATH}/output/czero/lib/
cp -rf ~/sshlics/alpha/* ${LOCAL_PATH}/output/czero/data/
cp -rf ~/geroscripts/* ${LOCAL_PATH}/output
cd ${LOCAL_PATH}
tar cvf ${LOCAL_PATH}/output.tar ./output/*
cd -

