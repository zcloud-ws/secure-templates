#!/usr/bin/env bash
set -e

COVERAGE_CMD="
go test -covermode=count -coverpkg=./... -coverprofile coverage/cover.out ./...
go tool cover -html coverage/cover.out -o coverage/cover.html
go tool cover -func coverage/cover.out | grep total | awk '{print \$3}' > coverage/total.txt
"

if [ "$RUN_IN_DOCKER" = "true" ]; then
docker run --rm -it --name stpl-coverage --restart no -u $UID \
  -v "$PWD":/source \
  -e GOCACHE=/tmp/.cache \
  -w /source \
  golang:"$STPL_GO_VERSION" \
  bash -c "$COVERAGE_CMD"
else
  bash -l -c "$COVERAGE_CMD"
fi

#go test -covermode=count -coverpkg=./... -coverprofile coverage/cover.out ./...
#go tool cover -html coverage/cover.out -o coverage/cover.html
#go tool cover -func coverage/cover.out | grep total | awk '{print $3}' > coverage/total.txt

COVERAGE_TOTAL="$(cut -d. -f1 coverage/total.txt)"
COLOR=lightgreen

if ((COVERAGE_TOTAL < 70)); then
  COLOR=orange
fi
if ((COVERAGE_TOTAL < 60)) ; then
  COLOR=yellow
fi
if ((COVERAGE_TOTAL < 50)) ; then
  COLOR=red
fi

echo "Update coverage to $COVERAGE_TOTAL% / $COLOR"

sed -i "s/Coverage-[[:digit:].]*%25-[a-zA-Z]*/Coverage-$COVERAGE_TOTAL%25-$COLOR/g" README.md
git add coverage
git add README.md

if [ "$COMMIT_COVERAGE_UPDATES" = "true" ]; then
  git commit -m "Update coverage to ${COVERAGE_TOTAL}%" README.md coverage/
fi
