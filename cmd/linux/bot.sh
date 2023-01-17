#!/usr/bin/env bash
rm -rf ./build/*
cd src || exit
go env -w GOOS=linux
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go build -o "../build/alice-bot-go"
cd ../build || exit
./alice-bot-go
