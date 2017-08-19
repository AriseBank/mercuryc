# Go Bindings for MERCURY (Linux Containers)

This package implements [Go](http://golang.org) bindings for the [MERCURY](http://linuxcontainers.org/) C API (libmercury).

## Requirements

This package requires [MERCURY 1.x](https://github.com/mercury/mercury/releases) and its development package to be installed. Works with [Go 1.x](http://golang.org/dl). Following command should install required dependencies on Ubuntu:

	apt-get install -y pkg-config mercury-dev

## Installing

To install it, run:

    go get gopkg.in/mercury/go-mercury.v2

## Documentation

Documentation can be found at [GoDoc](http://godoc.org/gopkg.in/mercury/go-mercury.v2).

## Stability

The package API will remain stable as described in [gopkg.in](https://gopkg.in).

## Examples

See the [examples](https://github.com/mercury/go-mercury/tree/v2/examples) directory for some.

## Contributing

We'd love to see go-mercury improve. To contribute to go-mercury;

* **Fork** the repository
* **Modify** your fork
* Ensure your fork **passes all tests**
* **Send** a pull request
	* Bonus points if the pull request includes *what* you changed, *why* you changed it, and *tests* attached.
	* For the love of all that is holy, please use `go fmt` *before* you send the pull request.

We'll review it and merge it in if it's appropriate.
