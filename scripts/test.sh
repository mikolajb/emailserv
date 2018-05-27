#!/usr/bin/env bash

rm coverage.out
set -e
echo "mode: atomic" > coverage.out

for d in $(go list ./... | grep -v /vendor); do
    go test -race -coverprofile=profile.out -coverpkg=$d -covermode=atomic $d | tee result.out
    if [ -f profile.out ]; then
		tail -n +2 profile.out >> coverage.out
		rm profile.out
	fi
	if [ -f result.out ]; then
		tail -n +1 result.out >> test-result.out
		rm result.out
	fi
done
