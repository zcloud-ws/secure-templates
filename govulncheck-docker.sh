#!/usr/bin/env bash
set -e

# Load Go version into STPL_GO_VERSION
. .go-version

docker run --rm -it --name govulncheck -u $UID \
  -v "$PWD":/source \
  -e GOCACHE=/tmp/.cache \
  -w /source \
  golang:"$STPL_GO_VERSION" \
  bash -c 'go install golang.org/x/vuln/cmd/govulncheck@latest && govulncheck -show verbose ./...' > govulncheck.log
