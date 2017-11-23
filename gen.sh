#!/bin/sh
pkg=github.com/linuxdeepin/go-gir
./girgen $GOPATH/src/$pkg/$1 &&\
	time go build -i -v $pkg/$1

