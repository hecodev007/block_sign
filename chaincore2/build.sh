#!/bin/bash

if [ $# == 0 ]; then
echo "./build.sh hcsync"
exit
fi;

BUILDSERVICE=$1
ROOTPATH=$(cd `dirname $0`; pwd)

cd $ROOTPATH

#all
if [ $BUILDSERVICE == 'all' ]; then
dir=$(ls -l $ROOTPATH/bin |awk '/^d/ {print $NF}')
for i in $dir
do
mkdir $ROOTPATH/build/$i -p
cd $ROOTPATH/bin/$i
echo "build" $i
go build -i -ldflags "-extldflags=-Wl,--allow-multiple-definition"
cp ./$i $ROOTPATH/build/$i/ -rf
done
exit
fi;

#one
mkdir ./build/$BUILDSERVICE -p
cd ./bin/$BUILDSERVICE
go build -i -ldflags "-extldflags=-Wl,--allow-multiple-definition"
cp ./$BUILDSERVICE $ROOTPATH/build/$BUILDSERVICE/
