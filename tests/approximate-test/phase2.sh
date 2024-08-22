#!/bin/bash


for ((i=1; i<=16000; i+=5))
do
    python3 tests/approximate-test/noiser.py tests/fe-sphere/fe-sphere.edges tests/fe-sphere/numbers_with_random.node_labels dat/subgraphs/subs$i/sub.edges dat/subgraphs/modified$i.edges 5
done


source_dir="dat/subgraphs"
colors_file="tests/fe-sphere/numbers_with_random.node_labels"

for ((i=1; i<=16000; i+=5))
do
  dir="$source_dir/modified$i"
  file="modified$i.edges"
  
  # Create the new directory
  mkdir -p "$dir"
  
  # Move the file from the source directory to the new directory
  mv "$source_dir/$file" "$dir/"

  # Copy the colors
  cp $colors_file $dir/same.node_labels
done