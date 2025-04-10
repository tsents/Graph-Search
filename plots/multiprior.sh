#!/bin/bash

# Prompt the user for input
echo "enter folder:"
read fname

priors="01234"

# Loop over each character in the string
for digit in $(echo $priors | fold -w1); do
    mkdir -p dat/$fname
    timeout -s SIGINT 2m ./subgraph_isomorphism -fmt=folder -branching=dat/$fname/branching$digit.csv -parse="%d,%d" -prior=$digit inputs/$fname inputs/$fname-sub/ > log.txt
done