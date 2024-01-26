#!/usr/bin/env just --justfile

coverprofile := "cover.out"

default:
    @just --list | grep -v default

test PKG="./..." *ARGS="":
    go test -race -failfast -count 1 -coverprofile {{ coverprofile }} {{ PKG }} {{ ARGS }}

vtest PKG="./..." *ARGS="": (test PKG ARGS "-v")

tests PKG="./..." *ARGS="":
    gotestsum -f dots -- -v -race -failfast -count 1 -coverprofile {{ coverprofile }} {{ PKG }} {{ ARGS }}

cover: test
    go tool cover -html {{ coverprofile }}

alias benchmark := bench

bench PKG="./..." *ARGS="":
    go test -v -count 1 -run x -bench . {{ PKG }} {{ ARGS }}

lint PKG="./...":
    golangci-lint run --new=false {{ PKG }}

mockgen:
    command mockgen >/dev/null 2>&1 || go install github.com/golang/mock/mockgen@latest

generate PKG="./...": mockgen
    go generate {{ PKG }}
