#!/bin/bash

# Prompt the user for input
echo "enter folder:"
read fname

priors="01234"

# Loop over each character in the string
for digit in $(echo $priors | fold -w1); do
    mkdir -p dat/$fname
    timeout -s SIGINT 12m ./subgraph_isomorphism -fmt=folder -factor=dat/$fname/factor$digit.csv -parse="%d %d" -prior=$digit inputs/$fname inputs/$fname-sub/ > log.txt
done