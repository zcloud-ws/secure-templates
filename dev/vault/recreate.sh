#!/usr/bin/env bash
set -e

./destroy.sh || true
./start-and-init.sh