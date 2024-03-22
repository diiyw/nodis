#!/bin/bash

protoc -I=. --proto_path=./ --go_out=. ./*.proto