#!/bin/sh
cd debug
rm -rf logs/*
sudo ./runsc -debug -debug-log $(pwd)/logs/  -TESTONLY-unsafe-nonroot run forest
