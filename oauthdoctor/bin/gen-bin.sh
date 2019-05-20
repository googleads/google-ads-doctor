#!/bin/bash

root=<Enter your Google Ads Doctor root path here>
bin=$root/bin

# Remove previously generated binaries
rm -f $root/windows
rm -f $root/osx
rm -f $root/linux

# Generate binaries from master branch
pushd $root
orig_branch=$(git rev-parse --abbrev-ref HEAD)
git checkout master
git pull origin master

# Generate binaries
CGO_ENABLED=0 GOOS=windows GOARCH=386 go build -o $bin/windows/oauthdoctor-386.exe $root/oauthdoctor.go
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $bin/windows/oauthdoctor-amd64.exe $root/oauthdoctor.go
CGO_ENABLED=0 GOOS=linux GOARCH=386 go build -o $bin/linux/oauthdoctor-386 $root/oauthdoctor.go
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $bin/linux/oauthdoctor-amd64 $root/oauthdoctor.go
CGO_ENABLED=0 GOOS=darwin GOARCH=386 go build -o $bin/osx/oauthdoctor-386 $root/oauthdoctor.go
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $bin/osx/oauthdoctor-amd64 $root/oauthdoctor.go

chmod 754 $bin/*/*
git checkout $orig_branch
popd
