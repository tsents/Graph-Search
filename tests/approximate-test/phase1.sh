#!/bin/bash

for ((i=1; i<=16000; i+=5))
do
  ./subgraph_isomorphism -subset=500 -fmt=folder -start=$i -parse="%d %d" -prior=0 -subout -err=0 -out=dat/subgraphs/sub$i.output tests/fe-sphere/ > log.txt
  # ./subgraph_isomorphism -subset=500 -fmt=folder -start=$i -parse="%d %d" -prior=0 -subout -err=5 -out=dat/subgraphs/sub$i.err tests/fe-sphere/ > log.txt
done
