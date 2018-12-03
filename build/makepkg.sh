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
make clean
#os_version=("linux-amd64-v3","linux-amd64-v4","darwin-amd64","windows-amd64")
os_version=("windows-amd64")
for os in ${os_version[@]}
    do
      echo $os
      make "gero-"${os}
      rm -rf $LOCAL_PATH/geropkg/bin
      rm -rf $LOCAL_PATH/geropkg/czero
      mkdir -p $LOCAL_PATH/geropkg/czero/data/
      mkdir -p $LOCAL_PATH/geropkg/czero/include/
      mkdir -p $LOCAL_PATH/geropkg/czero/lib/
      cp -rf $LOCAL_PATH/bin $LOCAL_PATH/geropkg
      cp -rf $CZERO_PATH/czero/data/* $SERO_PATH/build/geropkg/czero/data/
      cp -rf $CZERO_PATH/czero/include/* $SERO_PATH/build/geropkg/czero/include/
      if [ $os == "windows-amd64" ];then
        cp -rf  $CZERO_PATH/czero/lib_WINDOWS/* $SERO_PATH/build/geropkg/czero/lib/
      elif [ $os == "linux-amd64-v3" ];then
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $SERO_PATH/build/geropkg/czero/lib/
      elif [ $os == "linux-amd64-v4" ];then
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $SERO_PATH/build/geropkg/czero/lib/
      else
        cp -rf  $CZERO_PATH/czero/lib_DARWIN_AMD64/* $SERO_PATH/build/geropkg/czero/lib/
      if
      cd $LOCAL_PATH
      if [ -f ./geropkg_$os.tar.gz ]; then
        rm ./geropkg_$os.tar.gz
      fi
      tar czvf geropkg_$os.tar.gz geropkg/*

    done

