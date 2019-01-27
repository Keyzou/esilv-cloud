# Back-end server for Countries

_For ESILV - Cloud Application Development (2018 - 2019)_

Made by Dean Chérif, Lucile Jeanneret & Matthias Exbrayat.

## Installation

This project was made using Go 1.10, and should be working with the latest version of Go.

**Make sure your Go environment is correctly set up ! ($GOROOT, $GOPATH, etc...)**

**[You can use g to setup golang easily.](https://github.com/stefanmaric/g)**

In order to successfully setup the server, you have to

- Clone this repository in the following directory: `$GOPATH/src/github.com/keyzou/esilv-cloud`
- Install [dep](https://github.com/golang/dep) for dependency management
- Run `dep ensure` in your working directory
- After everything is done, make sure MongoDB is up and running (on port 27017), and run `go run main.go`
