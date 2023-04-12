#!/bin/bash

cd ../bin/unm
go build
cp unm ../../debug/rootfs/bin/
cd ../../debug
