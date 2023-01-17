#!/usr/bin/env bash
rm ./build/*
cd src || exit
go build -o "../build/alice-bot-go"
cd ../build || exit
./alice-bot-go
