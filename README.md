# Avi

##How to install
```sh
$ mkdir ~/gocode
$ export GOPATH=~/gocode
$ go get github.com/nvcook42/avi
```

## How to run a simulation
```sh
$ cd $GOPATH/src/github.com/nvcook42/avi/
$ make build
$ go run avi/avi.go  -logtostderr -ticks 50000 nathanielc/*.yaml
```
## How to view the simulation
The head program needs two packages. 

* protobuf-python
* panda3d runtime

```sh
$ python head/main.py save.avi
```
