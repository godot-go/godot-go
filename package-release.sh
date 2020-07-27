#!/bin/bash

# ensure a tag is passed in as an argument
[[ -z "$1" ]] && { echo "tag parameter not specified" ; exit 1; }

CURRENT_BRANCH=$(git symbolic-ref -q HEAD)

# current branch must be master
[[ "${CURRENT_BRANCH}" != "refs/heads/master" ]] && { echo "current working branch must be master" ; exit 1;}

set -x -e

go generate
rm -fr dist
mkdir -p dist
git archive master | tar -x -C dist/
cd godot_headers
git archive 3.2 | tar -x -C ../dist/godot_headers
cd ../dist
git init .
git add .
git commit -m "release $1"
git tag $1 -f
git remote add origin git@github.com:pcting/godot-go.git
git push origin --tag -f $1
