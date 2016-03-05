#!/bin/bash
set -x
set -e

rm -rf /usr/local/go/*
curl https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz | tar zx --strip-components=1 -C /usr/local/go

go version