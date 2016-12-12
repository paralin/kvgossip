#!/bin/bash
set +x
set -e

source ./jenkins_scripts/jenkins_env.bash

mkdir -p ${GOPATH}/bin
mkdir -p ${GOPATH}/src/github.com/fuserobotics
ln -fs $(pwd) ${GOPATH}/src/github.com/fuserobotics/kvgossip

pushd ${GOPATH}/src/github.com/fuserobotics/kvgossip
go get -d -v ./...
popd

set -x
