#!/usr/bin/env bash

export LOGXI=*
export LOGXI_FORMAT=pretty,happy

go run proxy.go -addr=127.0.0.1:5000 \
                -server=127.0.0.1:6000 \
                -loss=30 \
                -dup=0
