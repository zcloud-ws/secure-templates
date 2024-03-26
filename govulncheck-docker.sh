#!/usr/bin/env bash

docker run --rm -it --name govulncheck -u $UID \
  -v $PWD:/source \
  -e GOCACHE=/tmp/.cache \
  -w /source \
  golang:1.21.8 \
  bash -c 'go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck -show verbose ./...' > govulncheck.log
