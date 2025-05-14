#!/usr/bin/env just --justfile

coverprofile := "cover.out"

default:
    @just --list | grep -v default

deps:
    @mise install -q

test PKG="./..." *ARGS="": deps
    go test -race -failfast -count 1 -coverprofile {{ coverprofile }} {{ PKG }} {{ ARGS }}

vtest PKG="./..." *ARGS="": (test PKG ARGS "-v")

testsum PKG="./..." *ARGS="": deps
    gotestsum -f dots -- -v -race -failfast -count 1 -coverprofile {{ coverprofile }} {{ PKG }} {{ ARGS }}

cover PKG="./...": (test PKG)
    go tool cover -html {{ coverprofile }}

alias benchmark := bench

bench PKG="./..." *ARGS="": deps
    go test -v -count 1 -run x -bench . {{ PKG }} {{ ARGS }}

lint *PKGS="./...": deps
    golangci-lint run --new=false {{ PKGS }}

generate PKG="./...": deps
    go generate {{ PKG }}

tmpl DST:
    #!/usr/bin/env bash
    paths=(
        .github
        .gitignore
        .golangci.yml
        .mise
        justfile
        LICENSE
    )
    for p in "${paths[@]}"; do
        rm -rf "{{ DST }}/${p}"
        cp -R "$p" "{{ DST }}/${p}"
    done
