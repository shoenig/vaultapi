# vaultapi
A Go vault client for the rest of us.

[![Go Report Card](https://goreportcard.com/badge/github.com/shoenig/vaultapi)](https://goreportcard.com/report/github.com/shoenig/vaultapi) [![Build Status](https://travis-ci.org/shoenig/vaultapi.svg?branch=master)](https://travis-ci.org/shoenig/vaultapi) [![GoDoc](https://godoc.org/github.com/shoenig/vaultapi?status.svg)](https://godoc.org/github.com/shoenig/vaultapi) [![License](https://img.shields.io/github/license/shoenig/vaultapi.svg?style=flat-square)](LICENSE)

### About
[vaultapi](https://github.com/shoenig/vaultapi) is a vault client library for Go programs, targeted at
the "99% use case". What this means is that while an official Go client provided by Hashicorp exists
and exposes the complete functionality of vault, it is often difficult to use and is extremely inconvenient
to work with in test cases. This vault library for Go aims to be easily mockable while providing interfaces
that are easy to work with.

### Install
Like any Go library, just use `go get` to install. If the Go team ever officially blesses a package
manager, this library will incorporate that.

`go get github.com/shoenig/vaultapi`

### Usage
Creating a vault Client is very simple, just call `New` with the desired `ClientOptions`.

```go
tracer := log.New(os.Stdout, "vaultapi-", log.LstdFlags)
options := vaultapi.ClientOptions{
    Servers: []string{"https://localhost:8200"},
    HTTPTimeout: 10 * time.Seconds, // default
    SkipTLSVerification: false, // default
    Logger: tracer,
}

tokener := vaultapi.NewStaticToken("abcdefgh-abcd-abcd-abcdefgh")
client := vaultapi.New(options, tokener)
// client implements the vaultapi.Client interface

leader, err := client.Leader()
// etc ...
```