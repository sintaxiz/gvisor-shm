#!/bin/sh
sudo make copy TARGETS=runsc DESTINATION=debug
cd debug
rm -rf logs/*
sudo ./runsc -debug -debug-log $(pwd)/logs/  -TESTONLY-unsafe-nonroot run forest
