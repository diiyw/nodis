#!/bin/bash

protoc -I=. --proto_path=./ --go_out=. ./*.proto
pbjs -t static-module -w commonjs -o ../pb/nodis.js ./*.proto
pbts -o ../pb/nodis.d.ts ../pb/nodis.js