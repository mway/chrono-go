#!/usr/bin/env just --justfile

coverprofile := "cover.out"

default:
    @just --list | grep -v default

tidy:
    go mod tidy

test PKG="./..." *ARGS="":
    go test -v -race -failfast -count 1 -coverprofile {{ coverprofile }} {{ PKG }} {{ ARGS }}

cover: test
    go tool cover -html {{ coverprofile }}

alias benchmark := bench

bench PKG="./..." *ARGS="":
    go test -v -count 1 -run x -bench . {{ PKG }} {{ ARGS }}

lint PKG="./...":
    golangci-lint run --new=false {{ PKG }}

mockgen COMMIT="5b455625bd2c8ffbcc0de6a0873f864ba3820904":
    command mockgen >/dev/null 2>&1 || go install github.com/golang/mock/mockgen@{{ COMMIT }}

generate PKG="./...": mockgen
    go generate {{ PKG }}
