#!/bin/sh
set -e

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

# Create fake Go workspace if it doesn't exist yet.
workspace="$PWD/build/_workspace"
root="$PWD"
SERO_PATH="$PWD"
CZERO_PATH="$PWD/../go-czero-import"
echo $CZERO_PATH
_GOPATH=`cd ../../../../;pwd`
echo $_GOPATH

cd "$root"
args=()
index=0
for i in "$@"; do
   args[$index]=$i
   index=$[$index+1]
done

mkdir -p "$root/../go-czero-import/czero/lib"

cd "$root/../go-czero-import/czero"

cp -rf lib_DARWIN_AMD64/* lib/

cp -rf lib_LINUX_AMD64_V3/* lib/

cp -rf lib_WINDOWS_AMD64/* lib/

export LD_LIBRARY_PATH="../go-czero-import/czero/lib"

export DYLD_LIBRARY_PATH="../go-czero-import/czero/lib"


# Set up the environment to use the workspace.
GOPATH="$_GOPATH"
export GOPATH


# Run the command inside the workspace.
cd "$SERO_PATH"
PWD="$SERO_PATH"

#Launch the arguments with the configured environment.
exec "${args[@]}"