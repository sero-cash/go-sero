#!/bin/sh

LOCAL_PATH=$(cd `dirname $0`; pwd)
SERO_PATH="${LOCAL_PATH%/*}"
CZERO_PATH="${SERO_PATH%/*}/go-czero-import"

echo "update go-czero-import"
cd $CZERO_PATH
git fetch&&git rebase

echo "update go-sero"
cd $SERO_PATH
git fetch&&git rebase
make clean all

rm -rf $LOCAL_PATH/geropkg/bin
rm -rf $LOCAL_PATH/geropkg/czero
mkdir -p $LOCAL_PATH/geropkg/czero/data/
mkdir -p $LOCAL_PATH/geropkg/czero/include/
mkdir -p $LOCAL_PATH/geropkg/czero/lib/
cp -rf $LOCAL_PATH/bin $LOCAL_PATH/geropkg
cp -rf $CZERO_PATH/czero/data/* $SERO_PATH/build/geropkg/czero/data/
cp -rf $CZERO_PATH/czero/include/* $SERO_PATH/build/geropkg/czero/include/

function sysname() {

    SYSTEM=`uname -s |cut -f1 -d_`

    if [ "Darwin" == "$SYSTEM" ]
    then
        echo "Darwin"

    elif [ "Linux" == "$SYSTEM" ]
    then
        name=`uname  -r |cut -f1 -d.`
        echo Linux-V"$name"
    else
        echo "$SYSTEM"
    fi
}

SNAME=`sysname`

if [ "Darwin" == "$SNAME" ]
then
    echo $SNAME
    cp $CZERO_PATH/czero/lib_DARWIN_AMD64/* $SERO_PATH/build/geropkg/czero/lib/
elif [ "Linux-v3" == "$SNAME" ]
then
    echo $SNAME
    cp $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $SERO_PATH/build/geropkg/czero/lib/
elif [ "Linux-v4" == "$SNAME" ]
then
    echo $SNAME
    cp $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $SERO_PATH/build/geropkg/czero/lib/
fi

cd $LOCAL_PATH
if [ -f ./geropkg_*.tar.gz ]; then
	rm ./geropkg_*.tar.gz
fi
tar czvf geropkg_$SNAME.tar.gz geropkg/*
