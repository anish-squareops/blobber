#!/bin/sh
  
for i in $(seq 1 2)
do
  rm docker.local/blobber$i/log/*
  rm -rf docker.local/blobber$i/data/postgresql/*
  rm -rf docker.local/blobber$i/files/*
done

