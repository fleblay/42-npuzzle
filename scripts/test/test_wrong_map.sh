#!/bin/bash

search_dir=../maps/wrongMap
for entry in "$search_dir"/*
do
    echo "test : $entry"
    ../npuzzle -w 8 -split 32 -f "$entry"
done
