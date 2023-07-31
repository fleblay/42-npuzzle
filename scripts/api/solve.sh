#!/bin/bash
MAP=${1:-"maps/solvables/solvable_hard4_full_ram.map"}
curl -X POST --data "{\"size\": 4, \"board\":\"$(cat $MAP | tail -n 16 | tr '\n' ' ')\"}" localhost:8081/solve
