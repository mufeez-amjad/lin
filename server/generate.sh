#!/bin/bash

OUT_DIR=./src/generated

rm -rf ${OUT_DIR}
mkdir -p ${OUT_DIR}

# generate js codes via grpc-tools
grpc_tools_node_protoc \
--js_out=import_style=commonjs,binary:${OUT_DIR} \
--grpc_out=grpc_js:${OUT_DIR} \
--plugin=protoc-gen-grpc=$(npm root)/.bin/grpc_tools_node_protoc_plugin \
-I ./proto \
./proto/*.proto

# generate d.ts codes
protoc \
--plugin=protoc-gen-ts=$(npm root)/.bin/protoc-gen-ts \
--ts_out=grpc_js:${OUT_DIR} \
-I ./proto \
./proto/*.proto