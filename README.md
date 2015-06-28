knowledge-system
================

A project attempting to create a knowledge system by replicating human brain processes.

##Build instructions
###Prerequisites
Assuming a Linux system, install [Go](http://golang.org/). Run the following commands:

```bash
$ wget https://storage.googleapis.com/golang/go1.4.2.linux-amd64.tar.gz
$ tar -C /usr/local -xzf go1.4.2.linux-amd64.tar.gz
$ rm go1.4.2.linux-amd64.tar.gz
# Create Go workspace
$ mkdir ~/go
```

Add the following environment variables to finish installation and setup a working environment:

```bash
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH
```

Install build tool [gb](http://getgb.io/)

```bash
$ go get github.com/constabulary/gb/...
```

###To build

In the directory containing this project, run:

```bash
$ gb build
```

## Running

Use the help command to get instructions:

```bash
$ ./bin/system -h
$ ./bin/system test -h
$ ./bin/system concurrent -h
```