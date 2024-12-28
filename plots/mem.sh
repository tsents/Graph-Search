#!/bin/bash

# Path to the executable
executable="./subgraph_isomorphism"

# Define the prior values
prior_values=(0 2 4 5)

# Define the Gnp parameters
n_values=(100 200 300)
avg_degree_values=(10 20 30)

# Define other flags
fmt_flag="-fmt=folder"
logs_folder="logs-Gnp"
# Declare counters for each prior
declare -A counters
declare -A memory_usage

for prior in "${prior_values[@]}"; do
    counters[$prior]=0
    memory_usage[$prior]=0
done

# Function to run the executable and monitor it
run_and_monitor() {
    mkdir -p $logs_folder/core
    local prior="$1"
    local n="$2"
    local avg_degree="$3"
    local p=$(echo "scale=4; $avg_degree / $n" | bc)
    local log_file="$logs_folder/core/output_prior_${prior}_${n}_${avg_degree}.csv"

    touch "$log_file"
    
    echo "Success,Memory,Time" > "$log_file"
    
    for i in {1..100}; do
        output=$(timeout -s SIGINT 2m time -v $executable $fmt_flag -parse="%v %v" -subset -prior=$prior -gnp -n=$n -p=$p 2>&1 > /dev/null)

        exit_code=$?

        mem_usage=$(echo "$output" | grep "Maximum resident set size" | awk '{print $6}')
        elapsed_time=$(echo "$output" | grep -oP 'Elapsed \(wall clock\) time \(h:mm:ss or m:ss\): \K[0-9:.]+')

        # Log the result based on the exit code
        if [ "$exit_code" -eq 124 ]; then
            echo -n "1," >> "$log_file"  # Timeout occurred
        elif [ "$exit_code" -eq 0 ]; then
            ((counters[$prior]++))
            echo -n "0," >> "$log_file"  # Successful execution
        else
            echo "Process with prior=$prior, n=$n, avg_degree=$avg_degree failed with exit code $exit_code (iteration $i)." >> "$log_file"
        fi

        # Log the peak memory usage for this iteration
        printf "%s,%s\n" "$mem_usage" "$elapsed_time" >> "$log_file"
    done
}

# Run each combination of prior, n, and avg_degree values in parallel
for prior in "${prior_values[@]}"; do
    for n in "${n_values[@]}"; do
        for avg_degree in "${avg_degree_values[@]}"; do
            run_and_monitor "$prior" "$n" "$avg_degree" &
        done
    done
done

wait
