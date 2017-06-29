#!/bin/bash

set -e

echo building gruffles server
go install github.com/russross/gruffles

echo installing gruffles server
sudo mv $GOPATH/bin/gruffles /usr/local/bin/
sudo setcap cap_net_bind_service=+ep /usr/local/bin/gruffles
