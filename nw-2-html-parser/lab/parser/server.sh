#!/usr/bin/env bash

export LOGXI=*
export LOGXI_FORMAT=pretty,happy

go run server.go download.go
