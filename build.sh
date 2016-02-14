#!/usr/bin/env bash
GO15VENDOREXPERIMENT=1

GOOS=windows GOARCH=amd64 go build -o bin/gmr.exe cli.go
GOOS=darwin GOARCH=amd64 go build -o bin/gmr.osx cli.go
GOOS=linux GOARCH=amd64 go build -o bin/gmr.linux cli.go
