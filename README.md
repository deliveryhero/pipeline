# pipeline

[![GitHub Workflow Status](https://github.com/deliveryhero/pipeline/actions/workflows/ci.yml/badge.svg?branch=v2)](https://github.com/deliveryhero/pipeline/actions/workflows/ci.yml?query=branch:v2)
[![codecov](https://codecov.io/gh/deliveryhero/pipeline/branch/v2/graph/badge.svg)](https://codecov.io/gh/deliveryhero/pipeline)
[![GoDoc](https://img.shields.io/badge/pkg.go.dev-doc-blue)](http://pkg.go.dev/github.com/deliveryhero/pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/deliveryhero/pipeline)](https://goreportcard.com/report/github.com/deliveryhero/pipeline)

Pipeline is a go library that helps you build pipelines without worrying about channel management and concurrency.
It contains common fan-in and fan-out operations as well as useful utility funcs for batch processing and scaling.

If you have another common use case you would like to see covered by this package, please [open a feature request](https://github.com/deliveryhero/pipeline/issues).

## Functions

### func [Buffer](/buffer.go#L5)

`func Buffer(size int, in <-chan interface{}) <-chan interface{}`

Buffer creates a buffered channel that will close after the input
is closed and the buffer is fully drained

## Sub Packages

* [semaphore](./semaphore): package semaphore is like a sync.WaitGroup with an upper limit.

---
Readme created from Go doc with [goreadme](https://github.com/posener/goreadme)
