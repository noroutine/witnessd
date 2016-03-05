#!/bin/bash
set -x
set -e

export GOROOT=${HOME}/go1.6
export PATH=${GOROOT}/bin:$PATH

if [ ! -e ${GOROOT} ]; then
	mkdir -p ${GOROOT}
	wget https://storage.googleapis.com/golang/go1.6.linux-amd64.tar.gz
	tar xzf go1.6.linux-amd64.tar.gz --strip-components=1 -C ${GOROOT}
fi

mkdir -p ${GOPATH}

go version