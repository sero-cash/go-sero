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

os="all"
version="v0.3.1-beta.rc.5"
while getopts ":o:v:" opt
do
    case $opt in
        o)
        os=$OPTARG
        ;;
        v)
        version=$OPTARG
        ;;
        ?)
        echo "unkonw param"
        exit 1;;
    esac
done

if [ "$os" = "all" ]; then
    os_version=("linux-amd64-v3" "linux-amd64-v4" "darwin-amd64" "windows-amd64")
else
    os_version[0]="$os"
fi

for os in ${os_version[@]}
    do
      echo "make gero-${os}"
      make "gero-"${os}
      rm -rf $BUILD_PATH/gerotxpkg/bin
      rm -rf $BUILD_PATH/gerotxpkg/czero
      mkdir -p $BUILD_PATH/gerotxpkg/bin
      mkdir -p $BUILD_PATH/gerotxpkg/czero/data/
      mkdir -p $BUILD_PATH/gerotxpkg/czero/include/
      mkdir -p $BUILD_PATH/gerotxpkg/czero/lib/
      cp -rf $CZERO_PATH/czero/data/* $SERO_PATH/build/gerotxpkg/czero/data/
      cp -rf $CZERO_PATH/czero/include/* $SERO_PATH/build/gerotxpkg/czero/include/
      if [ $os == "windows-amd64" ];then
        mv $BUILD_PATH/bin/gero*.exe $BUILD_PATH/gerotxpkg/bin/tx.exe
        cp -rf  $CZERO_PATH/czero/lib_WINDOWS_AMD64/* $SERO_PATH/build/gerotxpkg/czero/lib/
      elif [ $os == "linux-amd64-v3" ];then
        mv $BUILD_PATH/bin/gero-v3* $BUILD_PATH/gerotxpkg/bin/tx
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $SERO_PATH/build/gerotxpkg/czero/lib/
      elif [ $os == "linux-amd64-v4" ];then
        mv $BUILD_PATH/bin/gero-v4* $BUILD_PATH/gerotxpkg/bin/tx
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $SERO_PATH/build/gerotxpkg/czero/lib/
      else
        mv $BUILD_PATH/bin/gero-darwin* $BUILD_PATH/gerotxpkg/bin/tx
        cp -rf  $CZERO_PATH/czero/lib_DARWIN_AMD64/* $SERO_PATH/build/gerotxpkg/czero/lib/
      fi
      cd $BUILD_PATH

      if [ $os == "windows-amd64" ];then
        rm -rf ./gerotx-*-$os.zip
        zip -r gero-$version-$os.zip gerotxpkg/*
      else
         rm -rf ./gerotx-*-$os.tar.gz
         tar czvf gerotx-$version-$os.tar.gz gerotxpkg/*
      fi

      cd $LOCAL_PATH

    done
rm -rf $BUILD_PATH/gerotxpkg/bin
rm -rf $BUILD_PATH/gerotxpkg/czero

