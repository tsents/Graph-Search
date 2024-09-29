#!/bin/bash

# Prompt the user for input
echo "enter folder:"
read fname

priors="0243"

# Loop over each character in the string
for digit in $(echo $priors | fold -w1); do
    mkdir -p dat/$fname
    timeout -s SIGINT 720s ./subgraph_isomorphism -fmt=folder -parse="%d %d" -prior=$digit -branching=dat/$fname/branching$digit.csv -hair=dat/$fname/hair$digit.csv inputs/$fname inputs/$fname-sub/
done