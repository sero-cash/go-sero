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
    os_version=("linux-amd64-v3" "darwin-amd64" "windows-amd64")
else
    os_version[0]="$os"
fi

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
#      cp -rf $CZERO_PATH/czero/include/* $SERO_PATH/build/geropkg/czero/include/
      if [ $os == "windows-amd64" ];then
        mv $BUILD_PATH/bin/gero*.exe $BUILD_PATH/geropkg/bin/gero.exe
        cp -rf  $CZERO_PATH/czero/lib_WINDOWS_AMD64/* $SERO_PATH/build/geropkg/czero/lib/
      elif [ $os == "linux-amd64-v3" ];then
#        mv $BUILD_PATH/bin/bootnode-v3*  $BUILD_PATH/geropkg/bin/bootnode
        mv $BUILD_PATH/bin/gero-v3* $BUILD_PATH/geropkg/bin/gero
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V3/* $SERO_PATH/build/geropkg/czero/lib/
      elif [ $os == "linux-amd64-v4" ];then
#        mv $BUILD_PATH/bin/bootnode-v4*  $BUILD_PATH/geropkg/bin/bootnode
        mv $BUILD_PATH/bin/gero-v4* $BUILD_PATH/geropkg/bin/gero
        cp -rf  $CZERO_PATH/czero/lib_LINUX_AMD64_V4/* $SERO_PATH/build/geropkg/czero/lib/
      else
        mv $BUILD_PATH/bin/gero-darwin* $BUILD_PATH/geropkg/bin/gero
        cp -rf  $CZERO_PATH/czero/lib_DARWIN_AMD64/* $SERO_PATH/build/geropkg/czero/lib/
      fi
      cd $BUILD_PATH

      if [ $os == "windows-amd64" ];then
        rm -rf ./gero-*-$os.zip
        zip -r gero-$version-$os.zip geropkg/*
      else
         rm -rf ./gero-*-$os.tar.gz
         tar czvf gero-$version-$os.tar.gz geropkg/*
      fi

      cd $LOCAL_PATH

    done
rm -rf $BUILD_PATH/geropkg/bin
rm -rf $BUILD_PATH/geropkg/czero

