#!/bin/bash

LICENSE=$(cat ./script/license.txt)

if [ -z "$1" ]; then
  echo "filename argument required, but none provided"
  echo "usage: $0 [filename]"
  exit 1
fi

if [ ! -f "$1"]; then
  echo "file not found"
  echo "usage: $0 [filename]"
  exit 2
fi

echo -e "$LICENSE\n" | cat - $1 > .bkp && mv .bkp $1
