#!/bin/bash
SIZE=${1:-4}
curl -sS localhost:8081/generate/$SIZE > grid.txt
curl -X POST --data "$(cat grid.txt)" localhost:8081/solve
curl -X POST --data "$(cat grid.txt)" localhost:8081/solve
