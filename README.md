# Tainted

A tool to compare which go packages will need be to rebuilt as a result of changes between two git diffs. 

Ideally used as part of a CI/CD pipeline to see which servies should be rebuilt and redeployed

N.B. Name inspired by terraforms taint terminology

# Project status

I do not have time to update or maintain this package but fortunatly Digital ocean have a much better and activly maintained tool here which does what atinted does and more,

https://github.com/digitalocean/gta

https://www.digitalocean.com/blog/gta-detecting-affected-dependent-go-packages/

# Requirments
- git MUST be installed and be on the path where tainted is run

# Install

### From Go
    go get -u github.com/kynrai/tainted

### From binaries
see releases for latest binaries

# Usage

### Basic usage
From the go project repo e.g. `$GOPATH/src/github.com/user/repo/`

    go list ./... | tainted

It using the standard go repo layout (recommended)

    go list ./cmd/... | tainted

You can manually set the git commit ranges, by default the previous commit is checked with HEAD. i.e. `HEAD~1..HEAD`

You can change any or all of the params

    go list ./... | tainted -from=HEAD~2

    go list ./... | tainted -from=HEAD~1 -to=HEAD~1

