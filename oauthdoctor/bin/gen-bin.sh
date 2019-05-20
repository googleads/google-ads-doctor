#!/bin/bash

# Before you run this script, please make sure your current directory is inside
# Google Ads Doctor git repo locally.

# Show the absolute path of the top-level directory
root=$(git rev-parse --show-toplevel)
src=$root/oauthdoctor
bin=$src/bin
pwd=$PWD

# Remove previously generated binaries
rm -f $root/windows
rm -f $root/osx
rm -f $root/linux

# Generate binaries from master branch
cd $src
cur_branch=$(git rev-parse --abbrev-ref HEAD)
if [[ "$cur_branch" != "master" ]] ; then
  git checkout master
fi
git pull origin master

# Generate binaries
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o $bin/windows/oauthdoctor-386.exe $src/oauthdoctor.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $bin/windows/oauthdoctor-amd64.exe $src/oauthdoctor.go
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o $bin/linux/oauthdoctor-386 $src/oauthdoctor.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $bin/linux/oauthdoctor-amd64 $src/oauthdoctor.go
CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -o $bin/osx/oauthdoctor-386 $src/oauthdoctor.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $bin/osx/oauthdoctor-amd64 $src/oauthdoctor.go

chmod 754 $bin/*/*
if [[ "$cur_branch" != "master" ]] ; then
  git checkout $cur_branch
fi
cd $pwd
