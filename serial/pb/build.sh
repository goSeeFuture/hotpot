#!/bin/bash

OUT_PATH="serial"

# 适用于多个gopath路径
tmp=(${GOPATH//:/ })
FIRST_GOPATH=${tmp[0]}


echo "FIRST_GOPATH" $FIRST_GOPATH

protoc  -I serial/pb -I=$FIRST_GOPATH/src -I=$FIRST_GOPATH/src/github.com/gogo/protobuf/protobuf --gofast_out="$OUT_PATH" serial/pb/proto_warpper.proto

if [ $? != 0 ];then
  echo "build protobuf fail!"
  read -p "press Enter to exit."
  exit
fi

echo "build protobuf success!"