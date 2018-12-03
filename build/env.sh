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
#current system verion is  DARWIN_AMD64
rm -rf ../go-czero-import/czero/lib/*
cd "$root/../go-czero-import/czero"
cp -rf lib_DARWIN_AMD64/* lib/
DYLD_LIBRARY_PATH="../go-czero-import/czero/lib_DARWIN_AMD64"
export DYLD_LIBRARY_PATH
cd "$root"

#current system verion is LINUX_AMD64_V3
#rm -rf ../go-czero-import/czero/lib
#cd "$root/../go-czero-import/czero"
#cp -rf lib_LINUX_AMD64_V3/* lib/
#cd "$root"
#LD_LIBRARY_PATH="../go-czero-import/czero/lib_LINUX_AMD64_V3"
#export LD_LIBRARY_PATH
#cd "$root"

#current system verion is  LINUX_AMD64_V4
#rm -rf ../go-czero-import/czero/lib
#cd "$root/../go-czero-import/czero"
#p -rf lib_LINUX_AMD64_V4/* lib/
#cd "$root"
#LD_LIBRARY_PATH="../go-czero-import/czero/lib_LINUX_AMD64_V4"
#export LD_LIBRARY_PATH
#cd "$root"

#current system verion is WINDOWS_AMD64
#rm -rf ../go-czero-import/czero/lib
#cd "$root/../go-czero-import/czero"
#cp -rf lib_WINDOWS_AMD64/* lib/
#cd "$root"

if [ $1 == "linux-v3" ]; then
    cd "$root/../go-czero-import/czero"
    cp -rf lib_LINUX_AMD64_V3/* lib/
    cd "$root"
    unset args[0]
    LD_LIBRARY_PATH="../go-czero-import/czero/lib_LINUX_AMD64_V3"
    export LD_LIBRARY_PATH
elif [ $1 == "linux-v4" ];then
    cd "$root/../go-czero-import/czero"
    cp -rf lib_LINUX_AMD64_V4/* lib/
    cd "$root"
    unset args[0]
    LD_LIBRARY_PATH="../go-czero-import/czero/lib_DARWIN_AMD64"
    export LD_LIBRARY_PATH
elif [ $1 == "darwin-amd64" ];then
#    cd "$root/../go-czero-import/czero"
#    ln -s lib_DARWIN_AMD64/* lib/
#    cd "$root"
#    unset args[0]
#    DYLD_LIBRARY_PATH="../go-czero-import/czero/lib_DARWIN_AMD64"
#    export DYLD_LIBRARY_PATH
     unset args[0]
elif [ $1 == "windows-amd64" ];then
    unset args[0]
    cd "$root/../go-czero-import/czero"
    cp -rf lib_WINDOWS_AMD64/* lib/
    cd "$root"
else
     echo "defalut"
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