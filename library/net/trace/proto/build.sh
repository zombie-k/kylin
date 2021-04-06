#!/usr/bin/env bash

protoc -I=. -I=${GOPATH}/src --go_out=plugins=grpc:. span.proto