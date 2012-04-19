#!/usr/bin/env bash

if [ ! -x makeslide ]; then
	go build makeslide.go
fi

./makeslide index.slide > index.html

