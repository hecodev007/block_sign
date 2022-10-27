#!/bin/bash

ROOTPATH=$(cd `dirname $0`; pwd)
dir=$(ls -l $ROOTPATH/bin |awk '/^d/ {print $NF}')

for i in $dir
do
rm $ROOTPATH/bin/$i/$i -f
done

rm $ROOTPATH/build/ -rf
