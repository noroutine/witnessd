
[![Go](https://img.shields.io/badge/Go-1.11-blue.svg)](https://golang.org/) [![Build Status](https://travis-ci.org/noroutine/witnessd.svg?branch=master)](https://travis-ci.org/noroutine/witnessd) [![Gitter](https://badges.gitter.im/turbovillains/dominion.svg)](https://gitter.im/turbovillains/dominion?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge) [![Issue Count](https://codeclimate.com/github/noroutine/witnessd/badges/issue_count.svg)](https://codeclimate.com/github/noroutine/witnessd/issues)

### Dominion Client

Dominion client implements P2P protocol designed to allow 2-6 players enjoy the game without centralized server.

#### How to Use

Write your own AI that communicates with this client or play yourself via CLI

[More about the Dominion](https://en.wikipedia.org/wiki/Dominion_(card_game))

#### Get Involved

    apt-get -qq install git build-essential libreadline-dev
    go get github.com/noroutine/dominion
    cd $GOPATH/src/github.com/noroutine/dominion
    make
    ./dominion --name Jack --join Game --port 9999

#### Usage

Just type

    dominion --help

So far it's not much

    Usage of ./dominion:
      -join string
            name of the group of the node
      -name string
            name of the player
      -port int
            client API port (default 9999)


You will end up in CLI, while in background also HTTP interface starts at port 9999 (controlled by parameter).

Number of commands are available in CLI, type 'help' to check
