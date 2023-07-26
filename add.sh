#!/bin/bash
curl -sS localhost:8081/generate/4 > grid.txt
curl -X POST --data "$(cat grid.txt)" localhost:8081/solve
curl -X POST --data "$(cat grid.txt)" localhost:8081/solve
