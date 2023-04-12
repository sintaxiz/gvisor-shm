#!/bin/bash

cd ../bin/file
go build
cp file ../../debug/rootfs/bin/
cd ../../debug
