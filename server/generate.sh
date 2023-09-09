#!/bin/bash

cd "$(dirname "$0")" 

OUT_DIR=./src/generated
PROTO_DIR=../proto

rm -rf ${OUT_DIR}
mkdir -p ${OUT_DIR}

GRPC_TOOLS_PATH=$(npm root)/.bin/grpc_tools_node_protoc

# generate js codes via grpc-tools
$GRPC_TOOLS_PATH \
--js_out=import_style=commonjs,binary:${OUT_DIR} \
--grpc_out=grpc_js:${OUT_DIR} \
--plugin=protoc-gen-grpc=$(npm root)/.bin/grpc_tools_node_protoc_plugin \
-I ${PROTO_DIR} \
${PROTO_DIR}/*.proto

# generate d.ts codes
protoc \
--plugin=protoc-gen-ts=$(npm root)/.bin/protoc-gen-ts \
--ts_out=grpc_js:${OUT_DIR} \
-I ${PROTO_DIR} \
${PROTO_DIR}/*.proto

cd - >/dev/null