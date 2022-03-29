#!/bin/sh

if [[ $# -eq 0 ]] ; then
  echo "Provide a output path for the binary"
  exit 1
fi

mkdir release

BASE=${PWD##*/}

#env GOOS=linux GOARCH=amd64 go build -o $1/$BASE-linux-amd64
#env GOOS=linux GOARCH=arm GOARM=5 go build -o $1/$BASE-linux-arm5-raspi
#env GOOS=linux GOARCH=arm GOARM=6 go build -o $1/$BASE-linux-arm6-raspi
#env GOOS=linux GOARCH=arm GOARM=7 go build -o $1/$BASE-linux-arm7-raspi
#env GOOS=linux GOARCH=arm64 go build -o $1/$BASE-linux-arm8-raspi4
env GOOS=windows GOARCH=amd64 go build -o $1/$BASE-windows-amd64.exe
env GOOS=darwin GOARCH=amd64 go build -o $1/$BASE-darwin-amd64

gh release upload v1.0 ./release/*
