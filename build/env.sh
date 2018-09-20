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
if [ ! -L "$ethdir/go-sero" ]; then
    mkdir -p "$ethdir"
    cd "$ethdir"
    ln -s ../../../../../. go-sero
    cd "$root"
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

DYLD_LIBRARY_PATH="../go-czero-import/czero/lib"
export DYLD_LIBRARY_PATH
LD_LIBRARY_PATH="../go-czero-import/czero/lib"
export LD_LIBRARY_PATH

# Run the command inside the workspace.
cd "$ethdir/go-sero"
PWD="$ethdir/go-sero"

# Launch the arguments with the configured environment.
exec "$@"
