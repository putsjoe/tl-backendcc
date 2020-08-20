#!/usr/bin/bash

SF1="sample-folders/folder1";
SF2="sample-folders/folder2";
NMS1="a1 a2 a3 a4 a5 a6 a7 a8 a9"
NMS2="b1 b2 b3 b4 b5 b6 b7 b8 b9"

function toucher1 () {
    for p in $NMS1; do 
        touch $SF1/$p.doc;
    done
}

function toucher2 () {
    for p in $NMS2; do 
        touch $SF2/$p.doc;
    done
}

if [ "$1" == "clean" ];
then
    rm $SF1/*doc;
    rm $SF2/*doc;
else
    toucher1 &
    toucher2 &
    wait
fi

