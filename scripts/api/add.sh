#!/bin/bash
SIZE=${2:-4}
ITERATE=${1:-10}

while (( i < $ITERATE))
do
curl -sS localhost:8081/generate/$SIZE > grid.txt
curl -X POST --data "$(cat grid.txt)" localhost:8081/solve
((i++))
done
