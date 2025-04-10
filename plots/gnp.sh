#!/bin/bash

# Path to the executable
executable="./subgraph_isomorphism"

# Define the prior values
prior_values=(0 2 4 5)

# Define the Gnp parameters
n_values=(100 1000 100000 1000000)
avg_degree_values=(2 4 8 16 32 64 128 256 512 1024)

# Define the proportion values
proportion_values=(0.01 0.1 0.2 0.5)

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
    local experiment_folder="$1"
    mkdir -p $logs_folder/$experiment_folder
    local prior="$2"
    local n="$3"
    local avg_degree="$4"
    local proportion="$5"
    local p=$(echo "scale=4; $avg_degree / $n" | bc)
    local subset=$(echo "scale=0; $n * $proportion / 1" | bc)
    local log_file="$logs_folder/$experiment_folder/output_prior_${prior}_${n}_${avg_degree}_prop_${proportion}.csv"

    touch "$log_file"
    
    echo "Success,Memory,Time" > "$log_file"
    
    for i in {1..100}; do
        output=$(timeout -s SIGINT 2m time -v $executable -subset=100 -prior=$prior -gnp -n=$n -p=$p 2>&1 > /dev/null)

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
            echo "Process with prior=$prior, n=$n, p=$p failed with exit code $exit_code (iteration $i)." >> "$log_file"
        fi

        # Log the peak memory usage for this iteration
        printf "%s,%s\n" "$mem_usage" "$elapsed_time" >> "$log_file"
    done
}

# Run for fixed avg_degree and varying n
fixed_avg_degree=4
for prior in "${prior_values[@]}"; do
    for n in "${n_values[@]}"; do
        # p=$(echo "scale=4; $fixed_avg_degree / $n" | bc)
        run_and_monitor "only" "$prior" "$n" "$fixed_avg_degree" "0.1"
    done
done

# # Run for fixed n and varying avg_degree
# fixed_n=1000
# for prior in "${prior_values[@]}"; do
#     for avg_degree in "${avg_degree_values[@]}"; do
#         run_and_monitor "fixed_n" "$prior" "$fixed_n" "$avg_degree" "0.1" &
#     done
# done

# # Run for constant n and avg_degree but varying fraction of G in S
# constant_n=1000
# constant_avg_degree=4
# constant_p=$(echo "scale=4; $constant_avg_degree / $constant_n" | bc)
# for prior in "${prior_values[@]}"; do
#     for proportion in "${proportion_values[@]}"; do
#         run_and_monitor "constant_n_avg_degree" "$prior" "$constant_n" "$constant_avg_degree" "$proportion" &
#     done
# done

wait
