#!/usr/bin/env bash

protoc --go_out=.test --go-grpc_out=.test .test/proto/*.proto

