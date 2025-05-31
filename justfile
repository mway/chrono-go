#!/usr/bin/env just --justfile

coverprofile := "cover.out"

default:
    @just --list | grep -v default

check:
    mise run check

cover: test
    mise run cover

fix:
    mise run fix

test:
    mise run test

vtest PKG="./..." *ARGS="":
    go test -race -failfast -count 1 -coverprofile {{ coverprofile }} -v {{ PKG }} {{ ARGS }}

tests PKG="./..." *ARGS="":
    gotestsum -f dots -- -v -race -failfast -count 1 -coverprofile {{ coverprofile }} {{ PKG }} {{ ARGS }}

alias bench := benchmark

benchmark PKG="./..." *ARGS="":
    go test -v -count 1 -run x -bench . {{ PKG }} {{ ARGS }}

lint *PKGS="./...":
    golangci-lint run --new=false {{ PKGS }}

alias gen := generate

generate PKG="./...":
    go generate {{ PKG }}
