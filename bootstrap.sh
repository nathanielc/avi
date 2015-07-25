#!/bin/bash

yum install -y golang
if [ ! -d /home/vagrant/gocode/src/github.com/nathanielc/avi ]
then
    mkdir -p /home/vagrant/gocode/src/github.com/nathanielc/avi
    chown -R vagrant:vagrant /home/vagrant/gocode
fi
