#!/bin/sh

LOCAL_PATH=$(cd `dirname $0`; pwd)
echo "LOCAL_PATH=$LOCAL_PATH"
SERO_PATH="${LOCAL_PATH%}"
echo "SERO_PATH=$SERO_PATH"
CZERO_PATH="${SERO_PATH%/*}/go-czero-import"
echo "CZERO_PATH=$CZERO_PATH"

echo "update go-czero-import"
cd $CZERO_PATH
git fetch&&git rebase

echo "update go-sero"
cd $SERO_PATH
git fetch&&git rebase
make clean
BUILD_PATH="${SERO_PATH%}/build"
os_version=("linux-amd64-v3" "linux-amd64-v4" "darwin-amd64" "windows-amd64")
for os in ${os_version[@]}
    do
      echo "make gero-${os}"
      make "gero-"${os}
      rm -rf $BUILD_PATH/geropkg/bin
      rm -rf $BUILD_PATH/geropkg/czero
      mkdir -p $BUILD_PATH/geropkg/bin
      mkdir -p $BUILD_PATH/geropkg/czero/data/
      mkdir -p $BUILD_PATH/geropkg/czero/include/
      mkdir -p $BUILD_PATH/geropkg/czero/lib/
      cp -rf $CZERO_PATH/czero/data/* $SERO_PATH/build/geropkg/czero/data/
      cp -rf $CZERO_PATH/czero/include/* $SERO_PATH/build/geropkg/czero/include/
      if [ $os == "windows-amd64" ];then
        mv $BUILD_PATH/bin/gero*.exe $BUILD_PATH/geropkg/bin/gero.exe
        cp -rf  $CZERO_PATH/czero/lib_WINDOWS_AMD64/* $SERO_PATH/build/geropkg/czero/lib/
      elif [ $os == "linux-amd64-v3" ];then
        mv $BUILD_PATH/bin/gero-v3* $BUILD_PATH/geropkg/bin/gero
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $SERO_PATH/build/geropkg/czero/lib/
      elif [ $os == "linux-amd64-v4" ];then
        mv $BUILD_PATH/bin/gero-v4* $BUILD_PATH/geropkg/bin/gero
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $SERO_PATH/build/geropkg/czero/lib/
      else
        mv $BUILD_PATH/bin/gero-darwin* $BUILD_PATH/geropkg/bin/gero
        cp -rf  $CZERO_PATH/czero/lib_DARWIN_AMD64/* $SERO_PATH/build/geropkg/czero/lib/
      fi
      cd $BUILD_PATH

      if [ $os == "windows-amd64" ];then
        if [ -f ./geropkg-$os.zip ]; then
              rm ./geropkg-$os.zip
        fi
        zip -r geropkg-$os.zip geropkg/*
      else
         if [ -f ./geropkg-$os.tar.gz ]; then
              rm ./geropkg-$os.tar.gz
         fi
         tar czvf geropkg-$os.tar.gz geropkg/*
      fi

      cd $LOCAL_PATH

    done
rm -rf $BUILD_PATH/geropkg/bin
rm -rf $BUILD_PATH/geropkg/czero

