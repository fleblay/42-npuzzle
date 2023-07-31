#!/bin/bash
SIZE=${1:-3}
curl localhost:8081/generate/${SIZE}
