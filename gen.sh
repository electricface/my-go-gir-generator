#!/bin/sh
pkg=github.com/linuxdeepin/go-gir
./girgen $GOPATH/src/$pkg/$1 gir-gen.toml &&\
	time go build -i -v $pkg/$1

