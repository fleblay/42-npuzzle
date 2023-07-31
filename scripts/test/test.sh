#!/bin/bash
RES=0
SIZE=3
ITER=0

while true
do
    python3 npuzzle-gen.py -s $SIZE -i 20  > test.txt && ../npuzzle -no-i -w 8 -split 32 -f test.txt -no-ui
	RES=$?
    if [ "$RES" -ne 0 ]
    then
        echo "test failed"
        exit 1
    fi
    if [ $ITER -eq 5 ]
    then
        SIZE=$((SIZE+1))
		echo "Now testing with size ${SIZE}"
        ITER=0
    else
        ITER=$((ITER+1))
    fi
    if [ $SIZE -eq 15 ]
    then
        break
    fi
done
rm -rf text.txt
