#!/bin/bash

# ensure a tag is passed in as an argument
[[ -z "$1" ]] && { echo "tag parameter not specified" ; exit 1; }

CURRENT_BRANCH=$(git symbolic-ref -q HEAD)

# current branch must be master
[[ "${CURRENT_BRANCH}" != "refs/heads/master" ]] && { echo "current working branch must be master" ; exit 1;}

set -x -e

go generate
git checkout -b release-$1
git add -f pkg/gdnative/*.gen.go
git add -f pkg/gdnative/*.classgen.go
git add -f pkg/gdnative/*.gen.c
git add -f pkg/gdnative/*.gen.h
git commit -m "release $1"
git tag $1 -f
git push origin --tag $1
git checkout master
