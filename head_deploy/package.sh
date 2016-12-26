#!/bin/bash

for f in windows linux
do
    tar --transform 's,^\.,avi,' -czvf "avi-$f.tar.gz" --dereference -C "$f" .
done
