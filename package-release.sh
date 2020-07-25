#!/bin/bash

git checkout -b $1
git add -f pkg/gdnative/*.gen.go
git add -f pkg/gdnative/*.classgen.go
git add -f pkg/gdnative/*.gen.c
git add -f pkg/gdnative/*.gen.h
git commit -m "release $1"
git tag $1 -f