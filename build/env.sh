#!/bin/sh
set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
ethdir="$workspace/src/github.com/sero-cash"
rm -rf "$ethdir"
if [ ! -L "$ethdir/go-sero" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../. go-sero
    cd "$root"
fi

args=()
index=0
for i in "$@"; do
   args[$index]=$i
   index=$[$index+1]
done

mkdir -p "../go-czero-import/czero/lib"


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
    rm -rf ../go-czero-import/czero/lib/*
    cd "$root/../go-czero-import/czero"
    cp -rf lib_DARWIN_AMD64/* lib/
    DYLD_LIBRARY_PATH="../go-czero-import/czero/lib_DARWIN_AMD64"
    export DYLD_LIBRARY_PATH
elif [ "Linux-V3" == "$SNAME" ]
then
   rm -rf ../go-czero-import/czero/lib/*
   cd "$root/../go-czero-import/czero"
   cp -rf lib_LINUX_AMD64_V3/* lib/
   LD_LIBRARY_PATH="../go-czero-import/czero/lib_LINUX_AMD64_V3"
   export LD_LIBRARY_PATH
elif [ "Linux-V4" == "$SNAME" ]
then
    rm -rf ../go-czero-import/czero/lib/*
    cd "$root/../go-czero-import/czero"
    cp -rf lib_LINUX_AMD64_V4/* lib/
    LD_LIBRARY_PATH="../go-czero-import/czero/lib_LINUX_AMD64_V4"
    export LD_LIBRARY_PATH
elif [ "$SNAME" == "Linux-*" ]
then
     echo "only support linux kernal v3 or v4"
     exit
elif [ "$SNAME" == "MINGW32" ]
then
    rm -rf ../go-czero-import/czero/lib/*
    cd "$root/../go-czero-import/czero"
    cp -rf lib_WINDOWS_AMD64/* lib/
else
   echo "only support Mingw"
   exit
fi

cd "$root"

if [ $1 == "linux-v3" ]; then
    cd "$root/../go-czero-import/czero"
    cp -rf lib_LINUX_AMD64_V3/* lib/
    cd "$root"
    unset args[0]
elif [ $1 == "linux-v4" ];then
    cd "$root/../go-czero-import/czero"
    cp -rf lib_LINUX_AMD64_V4/* lib/
    cd "$root"
    unset args[0]
elif [ $1 == "darwin-amd64" ];then
     cd "$root/../go-czero-import/czero"
     cp -rf lib_DARWIN_AMD64/* lib/
     cd "$root"
     unset args[0]
elif [ $1 == "windows-amd64" ];then
    unset args[0]
    cd "$root/../go-czero-import/czero"
    cp -rf lib_WINDOWS_AMD64/* lib/
    cd "$root"
else
     echo "local"
fi


if [ ! -L "$ethdir/go-czero-import" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../../go-czero-import/. go-czero-import
    cd "$root"
fi

# Set up the environment to use the workspace.
GOPATH="$workspace"
export GOPATH


# Run the command inside the workspace.
cd "$ethdir/go-sero"
PWD="$ethdir/go-sero"

#Launch the arguments with the configured environment.
exec "${args[@]}"