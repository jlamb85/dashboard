#!/bin/bash
# Quick wrapper script for the hashpass utility

cd "$(dirname "$0")"

if [ ! -f "hashpass" ]; then
    echo "Building hashpass utility..."
    go build -o hashpass
fi

./hashpass "$@"
