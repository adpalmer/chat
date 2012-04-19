#!/usr/bin/env bash

BINS="
	tcp
	http
	markov
"

SUPPORT="
	chan.go
	chat-simple.go
	echo-no-concurrency.go
	echo.go
	embed.go
	goroutines.go
	hello-net.go
	hello.go
	select.go
"

cd bin
for fn in $BINS; do go build ../$fn; done
for fn in $SUPPORT; do go build ../support/$fn; done

