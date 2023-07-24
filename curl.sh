curl -X POST --data "{\"size\": 4, \"board\":\"`cat maps/solvables/solvable_hard4_full_ram.map | tail -n 16 | tr '\n' ' '`\"}" localhost:8081
