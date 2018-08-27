
[![Go](https://img.shields.io/badge/Go-1.11-blue.svg)](https://golang.org/) [![Build Status](https://travis-ci.org/noroutine/witnessd.svg?branch=master)](https://travis-ci.org/noroutine/witnessd) [![Gitter](https://badges.gitter.im/turbovillains/dominion.svg)](https://gitter.im/turbovillains/dominion?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

### Witness Daemon

Rule-based datacenter switchover daemon

### Get Involved

    apt-get -qq install git build-essential libreadline-dev
    go get github.com/noroutine/witnessd
    cd $GOPATH/src/github.com/noroutine/witnessd
    make

### Usage

In one terminal:

    ./witnessd --name Jack --join Group --port 9999

In another terminal:

    ./witnessd --name Jill --join Group --port 9998

For parameters just type:

    witnessd --help

So far it's not much

    Usage of ./witnessd:
      -join string
            name of the group of the node
      -name string
            name of the client
      -port int
            client API port (default 9999)


You will end up in CLI, while in background also HTTP interface starts at port 9999 (controlled by parameter).

Number of commands are available in CLI, type 'help' to check
