#!/bin/bash

# ensure a tag is passed in as an argument
[[ -z "$1" ]] && { echo "tag parameter not specified" ; exit 1; }

GIT_BRANCH=$(git symbolic-ref -q HEAD)
GIT_BRANCH=${GIT_BRANCH##refs/heads/}

if [[ $(git diff --stat) != '' ]]; then
  GIT_HASH=$(git stash create "dirty commit package")
  GIT_TAG_BRANCH=$(echo ${GIT_BRANCH} | sed 's/\//-/g')
  GIT_TAG="$1-${GIT_TAG_BRANCH}-${GIT_HASH}"
else
  GIT_HASH=${GIT_BRANCH}

  if [[ "${GIT_BRANCH}" == "master" ]]; then
      GIT_TAG="$1"
  else
      GIT_TAG_BRANCH=$(echo ${GIT_BRANCH} | sed 's/\//-/g')
      GIT_TAG="$1-${GIT_TAG_BRANCH}"
  fi
fi

echo "hash checkout ${GIT_HASH}, with tag ${GIT_TAG}"

# echo "Wainting for cancellation..."
# read -t 30 -n 1

set -x -e

go generate
rm -fr dist
mkdir -p dist
git archive ${GIT_HASH} | tar -x -C dist/
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

if [[ "$2" == "publish" ]]; then
  git remote add origin git@github.com:pcting/godot-go.git
  git push origin --tag -f
fi

cd ..
