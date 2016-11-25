#!/bin/bash

DIR="$(cd "$(dirname "$0")" && pwd)"

echo "Script Running From $DIR"

ROOT=$DIR

cd $ROOT
PWD="$(pwd)"

echo "PWD=$PWD"
EXECUTABLE="$(basename $PWD)"

echo "Executable = $EXECUTABLE"

justdoit -watch="./" -include="(.+\.go|.+\.c|.+\.yaml)$" -build="go install -v" -run="$GOPATH/bin/$EXECUTABLE"