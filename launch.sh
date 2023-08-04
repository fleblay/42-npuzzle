#!/bin/bash
make
GOGC=200 GOMEMLIMIT=7GiB API=true ./npuzzle -d
