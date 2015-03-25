# avi

How to install

$ mkdir ~/gocode
$ export GOPATH=~/gocode
$ export PATH=$GOPATH/bin:$PATH
$ go get github.com/nvcook42/avi

How to run a simulation

$ cd $GOPATH/src/github.com/nvcook42/avi/
$ go run avi/avi.go  -logtostderr -ticks 50000 nathanielc/*.yaml

How to view the simulation

$ python head/main.py save.avi
