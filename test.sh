#!/bin/bash
RES=0
SIZE=4
ITER=0

while [ $RES -eq 0 ]
do
    python npuzzle-gen.py -s $SIZE -i 1  > test.txt && go run . -w 8 -ss 32 -f test.txt
    if [ $? -ne 0 ]
    then
        echo "test failed"
        exit 1
    fi
    if [ $ITER -eq 20 ]
    then
        SIZE=$((SIZE+1))
        ITER=0
    else
        ITER=$((ITER+1))
    fi
    RES=$?
    if [ $SIZE -eq 120 ]
    then
        break
    fi
done