#!/bin/bash

for ((i=1; i<=16000; i+=5))
do
    ./subgraph_isomorphism -fmt=folder -parse="%d %d" -prior=0 -err=10 -out=dat/subgraphs/modified_double$i.output dat/subgraphs/modified$i dat/subgraphs/subs$i > log.txt 
done