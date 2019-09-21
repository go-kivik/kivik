#!/bin/bash

EXPECTED=$(cat ./script/license.txt)

CODE=0
GOFILES=$(find . -name "*.go")

for FILE in $GOFILES; do
  BLOCK=$(head -n $(wc -l <<< "$EXPECTED") $FILE)

  if [ "$BLOCK" != "$EXPECTED" ]; then
    echo "file missing license: $FILE"
    CODE=1
  fi
done

exit $CODE
