import os
import pandas as pd
import re

def read_broken_csv(path):
    data = []

    with open(path) as f:
        for line in f:
            line = line.strip()
            # Match two numbers (including optional decimal)
            matches = re.findall(r"-?\d+(?:\.\d+)?", line)
            if len(matches) == 2:
                time, memory = map(float, matches)
                data.append((time, memory))
            else:
                print(f"Skipping malformed line in {path}: {line}")

    return pd.DataFrame(data, columns=["time", "memory"])


def summarize_results_from_files(datasets, algorithms, rows_per_dataset=None):
    summary = []

    for dataset in datasets:
        dataset_path = dataset
        print(dataset)
        n_rows = rows_per_dataset.get(dataset, 40) if rows_per_dataset else 40
        for algo in algorithms:
            file_path = os.path.join(dataset_path, f"{algo}_clique.csv")

            if not os.path.exists(file_path):
                summary.append({
                    'Dataset': dataset,
                    'Algorithm': algo,
                    'Time (s)': "File missing",
                    'Memory (MB)': "File missing",
                    'Success Rate': "—"
                })
                continue

            df = read_broken_csv(file_path)
            df = df.tail(n_rows)

            successful = df[(df['time'] != -1) & (df['memory'] != -1)]
            success_count = len(successful)
            total_count = len(df)

            if success_count > 0:
                time_str = f"{successful['time'].mean():.2f} ± {successful['time'].std():.2f}"
                mem_str = f"{successful['memory'].mean():.2f} ± {successful['memory'].std():.2f}"
            else:
                time_str = mem_str = "—"

            success_rate = f"{(success_count / total_count) * 100:.1f}%"

            summary.append({
                'Dataset': dataset,
                'Algorithm': algo,
                'Time (s)': time_str,
                'Memory (MB)': mem_str,
                'Success Rate': success_rate
            })

    return pd.DataFrame(summary)

# # Example usage:
datasets = ["IMDB", "wiki2009", "flybrain"]
algorithms = ["velcro", "VF2++", "SPHERA"]
results_table = summarize_results_from_files(datasets, algorithms)
print(results_table.to_string(index=False))
