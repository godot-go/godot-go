#!/bin/bash

# ensure a tag is passed in as an argument
[[ -z "$1" ]] && { echo "tag parameter not specified" ; exit 1; }

CURRENT_BRANCH=$(git symbolic-ref -q HEAD)

# current branch must be master
if [[ "${CURRENT_BRANCH}" != "refs/heads/master" ]]; then
    BRANCH_TAG=$(echo ${GIT_BRANCH} | tr -dc '[:alnum:]]')
    GIT_TAG="$1-${BRANCH_TAG}"
else
    GIT_TAG="$1"
fi

set -x -e

go generate
rm -fr dist
mkdir -p dist
git archive master | tar -x -C dist/
cd godot_headers
git archive 3.2 | tar -x -C ../dist/godot_headers
cd ../dist
git init .
rm .gitignore
cp ../pkg/gdnative/*.wrappergen.h pkg/gdnative/
cp ../pkg/gdnative/*.wrappergen.c pkg/gdnative/
cp ../pkg/gdnative/*.typegen.go pkg/gdnative/
cp ../pkg/gdnative/*.classgen.go pkg/gdnative/
git add -f pkg/gdnative/*.wrappergen.h
git add -f pkg/gdnative/*.wrappergen.c
git add -f pkg/gdnative/*.typegen.go
git add -f pkg/gdnative/*.classgen.go
git add .
git commit -m "release $1"
git tag ${GIT_TAG} -f
git remote add origin git@github.com:pcting/godot-go.git
git push origin --tag -f
