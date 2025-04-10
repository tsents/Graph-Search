import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
from matplotlib.gridspec import GridSpec
# Plot distribution as lines
from scipy.stats import gaussian_kde


plt.rcParams["font.family"] = "DejaVu Serif"
plt.rcParams["font.serif"] = ["Times New Roman"]
plt.rcParams["font.size"] = 16

data_list = []


files = [f'dat/flybrain/branching{i}.csv' for i in range(5)]
permuted_indices = [3, 2, 1, 4, 0, 5]

colors = ['c', 'r', 'pink', 'y', 'b', 'mediumpurple'] 
names = [
    "Degree S",
    "Degree G",
    "Greedy G",
    "Random",
    "Greedy S",
    "Combined S"
]
# print(names.permute())
# names = [
#     "Random",
#     "Gready G",
#     "Degree G",
#     "Gready S",
#     "Degree S",
#     "Combined S"
# ]


# Create figure and grid layout
fig = plt.figure(figsize=(10, 8))
gs = GridSpec(4, 4, figure=fig)

# Scatter plot in the main panel
ax_scatter = fig.add_subplot(gs[1:4, 0:3])
ax_hist_x = fig.add_subplot(gs[0, 0:3], sharex=ax_scatter)
ax_hist_y = fig.add_subplot(gs[1:4, 3], sharey=ax_scatter)


for i in permuted_indices[:5]:
    file_path = files[i]
    df = pd.read_csv(file_path)
    df = df[df["Depth"] > 9000]
    df["BranchingFactor"] = np.log10(df["BranchingFactor"])
    df['FeatureCost'] = df.apply(lambda row: df[df['Depth'] > row['Depth']]['BranchingFactor'].mean(), axis=1)
    df.dropna(inplace=True)
    ax_scatter.scatter(df['BranchingFactor'], df['FeatureCost'], label=names[i], color=colors[i], marker='o', alpha=0.6)
    
    
    kde_x = gaussian_kde(df['BranchingFactor'])
    kde_y = gaussian_kde(df['FeatureCost'])

    x_vals = np.linspace(df['BranchingFactor'].min(), df['BranchingFactor'].max(), 100)
    y_vals = np.linspace(df['FeatureCost'].min(), df['FeatureCost'].max(), 100)
    
    ax_hist_x.plot(x_vals, kde_x(x_vals), color=colors[i], alpha=0.7)
    ax_hist_y.plot(kde_y(y_vals), y_vals, color=colors[i], alpha=0.7)

# Customize plots
ax_scatter.set_title('Current Branching Factor vs Feature Branching Factor')
ax_scatter.set_xlabel('Branching Factor')
ax_scatter.set_ylabel('Mean Future Branching Factor')
ax_scatter.legend()
ax_scatter.grid(True)

ax_hist_x.set_ylabel('Density')
ax_hist_y.set_xlabel('Density')

# Remove axis labels for cleaner look
plt.setp(ax_hist_x.get_xticklabels(), visible=False)
plt.setp(ax_hist_y.get_yticklabels(), visible=False)

plt.tight_layout()
plt.savefig("plots/future.png",dpi=300)

# -------------------------
# Part 2: Reading Log Files and Plotting Metrics
# -------------------------

def time_to_seconds(time_str):
    """Convert a time string formatted as 'M:SS.ss' to total seconds."""
    try:
        parts = time_str.split(":")
        minutes = float(parts[0])
        seconds = float(parts[1])
        return minutes * 60 + seconds
    except Exception as e:
        return np.nan

# Define the directories (groups) and labels for the groups
group_dirs = ["dat/logs-wiki2009", "dat/logs-IMDB3", "dat/logs-flybrain3"]
group_labels = ["wiki20009", "IMDB", "flybrain"]

n_priors = 6  # each group has 6 priors (output_prior_0_0.csv, output_prior_1_0.csv, ..., output_prior_5_0.csv)

# Prepare containers to hold the metrics for each group and each prior.
# Each is a 2D array with shape (n_groups, n_priors)
metrics_success_rate = np.zeros((len(group_dirs), n_priors))
metrics_memory_success = np.zeros((len(group_dirs), n_priors))
metrics_time_success = np.zeros((len(group_dirs), n_priors))

# Loop over groups and within each group loop over each prior file.
for g_idx, group_dir in enumerate(group_dirs):
    for i in range(n_priors):
        log_file = f'{group_dir}/output_prior_{i}_0.csv'
        df_log = pd.read_csv(log_file)
        
        # Convert the Time column to seconds.
        df_log['Time'] = df_log['Time'].apply(time_to_seconds)
        
        # In this dataset a success is indicated when Success == 0.
        # Thus, we calculate the success rate as 100 - (mean(Success)*100)
        success_rate = 100 - df_log['Success'].mean() * 100
        metrics_success_rate[g_idx, i] = success_rate
        
        # For memory and time, consider only successful outcomes (Success == 0)
        df_success = df_log[df_log['Success'] == 0]
        avg_memory = df_success['Memory'].mean() if not df_success.empty else np.nan
        avg_time = df_success['Time'].mean() if not df_success.empty else np.nan
        
        metrics_memory_success[g_idx, i] = avg_memory
        metrics_time_success[g_idx, i] = avg_time

# ---------------------------
# Plotting Grouped Bar Charts
# ---------------------------
# We will create one grouped bar chart per metric.
x = np.arange(len(group_dirs))    # x-axis positions for each group
bar_width = 0.1                  # width of each bar (adjust as needed)

n_methods = len(names)


# Create a figure with three subplots (one for each metric)
fig, axs = plt.subplots(1, 3, figsize=(18, 6))

# Success Rate Plot
k=0
for i in permuted_indices:
    k += 1
    offset = (k - ((n_methods+1)/ 2)) * bar_width
    axs[0].bar(x + offset, 100 -metrics_success_rate[:, i], width=bar_width,
               label=names[i],color=colors[i])
axs[0].set_ylabel("Failure Rate")
axs[0].set_title("Failure Rate per Method")
axs[0].set_xticks(x)
axs[0].set_xlim(-0.5,2.5)
axs[0].set_xticklabels(group_labels)

# Memory When Success Plot
k=0
for i in permuted_indices:
    k += 1
    offset = (k - ((n_methods+1)/ 2)) * bar_width
    axs[1].bar(x + offset, metrics_memory_success[:, i], width=bar_width,
               label=names[i],color=colors[i])
axs[1].set_ylabel("Average Memory at Success (Kilobytes)")
axs[1].set_title("Memory at Success per Method")
axs[1].set_xticks(x)
axs[1].set_xlim(-0.5,2.5)
axs[1].set_xticklabels(group_labels)

# Time When Success Plot
k=0
for i in permuted_indices:
    k += 1
    offset = (k - ((n_methods+1)/ 2)) * bar_width
    axs[2].bar(x + offset, metrics_time_success[:, i], width=bar_width,
               label=names[i],color=colors[i])
axs[2].set_ylabel("Average Time at Success (s)")
axs[2].set_yscale("log")
axs[2].set_title("Time at Success per Method")
axs[2].set_xticks(x)
axs[2].set_xticklabels(group_labels)
axs[2].set_xlim(-0.5,2.5)
axs[2].legend()

plt.tight_layout()
plt.savefig("plots/metrics.png", dpi=300)


# ---------------------------
# Plotting GNP
# ---------------------------
import glob
import re
import os

# List of base directories
base_dirs = [
    "dat/logs-Gnp3/constant_n_avg_degree",
    "dat/logs-Gnp3/fixed_avg_degree",
    "dat/logs-Gnp3/fixed_n"
]
# Pattern to extract parameters from filenames like:
# output_prior_{method}_{n}_{n*p}_prop_{proportion}.csv
pattern = re.compile(r"output_prior_(\d+)_(\d+)_(\d+)_prop_([\d.]+)\.csv")

# List to collect all aggregated data
data_list = []

# Loop over each base directory and then over all CSV files in its 'files' subfolder
for base in base_dirs:
    csv_pattern = os.path.join(base, "*.csv")
    for filepath in glob.glob(csv_pattern):
        filename = os.path.basename(filepath)
        match = pattern.search(filename)
        if not match:
            print(f"Filename does not match pattern: {filename}")
            continue

        method_str, n_str, np_str, prop_str = match.groups()
        # Convert extracted values to proper types
        method = int(method_str)
        n = int(n_str)
        n_p = int(np_str)
        proportion = float(prop_str)

        # Read the CSV data
        df = pd.read_csv(filepath)
        df['Time'] = df['Time'].apply(time_to_seconds)

        # Assuming the CSV has columns: 'success', 'memory', 'runtime'
        # where 'success' is a binary indicator (1 for success, 0 otherwise)
        avg_success_rate = 1 - df["Success"].mean()

        # For memory and runtime, we compute averages only on successful runs.
        success_df = df[df["Success"] == 0]
        avg_memory = success_df["Memory"].mean() if not success_df.empty else None
        avg_runtime = success_df["Time"].mean() if not success_df.empty else None

        # Append the extracted parameters and computed averages
        data_list.append({
            "directory": base.split('/')[-1],
            "method": method,
            "n": n,
            "n*p": n_p,
            "proportion": proportion,
            "avg_success_rate": avg_success_rate,
            "avg_memory": avg_memory,
            "avg_runtime": avg_runtime
        })

df = pd.DataFrame(data_list)

directories = [
    "constant_n_avg_degree",
    "fixed_avg_degree",
    "fixed_n"
]

dir_titles =  [
    "Varying Proportion",
    "Varying n",
    "Varying Average Degree"
]

x_mapping = {
    "constant_n_avg_degree": "proportion",  # typically the average degree
    "fixed_avg_degree": "n",        # n changes while n*p ~ constant
    "fixed_n": "n*p"    # or "n*p", or "p", depending on your data
}

row_info = [
    ("Failure", "avg_success_rate"),
    ("Log(Memory)", "avg_memory"),
    ("Log(Run Time)", "avg_runtime"),
]

# Create the 3×3 subplots
fig, axes = plt.subplots(nrows=3, ncols=3, figsize=(15, 12))
# Make some space
plt.subplots_adjust(hspace=0.25, wspace=0.10)

for col_index, directory in enumerate(directories):

    subset_dir = df[df["directory"] == directory].copy()

    # Figure out which column to use on the x-axis
    x_col = x_mapping[directory]

    for row_index, (row_label, metric_col) in enumerate(row_info):
        ax = axes[row_index, col_index]

        if row_label == "Failure":
            subset_dir["metric"] = 1.0 - subset_dir[metric_col]
        else:
            subset_dir["metric"] = subset_dir[metric_col]

        subset_dir.sort_values(by=[x_col, "method"], inplace=True)

        n_methods = len(names)
        bar_width = 0.8 / n_methods

        x_vals = sorted(subset_dir[x_col].unique())

        for x_index, x_val in enumerate(x_vals):
            # For each method, plot one bar
            k = 0
            for method_idx in permuted_indices:
                k += 1
                method_name = names[method_idx]
                # Filter to the row(s) matching this x_val and method
                y_subset = subset_dir[
                    (subset_dir[x_col] == x_val) & (subset_dir["method"] == method_idx)
                ]

                # If there's a row, extract the metric value; else use 0 or np.nan
                if not y_subset.empty:
                    y_val = y_subset["metric"].iloc[0]
                else:
                    y_val = 0  # or np.nan, if you prefer

                x_position = (x_index + (k - ((n_methods+1)/ 2)) * bar_width)
                
                label = method_name if x_index == 0 else None
                
                ax.bar(x_position, y_val, bar_width, color=colors[method_idx], label=label)

        # 3) Set the x-ticks to be at each group position (the center of each group)
        ax.set_xticks(range(len(x_vals)))
        ax.set_xticklabels(x_vals)

        # Title only on top row
        if row_index == 0:
            ax.set_title(dir_titles[col_index])

        # Y-axis label
        if row_index == 0 and col_index == 0:
            ax.set_ylabel("Failure")
        elif row_index == 1 and col_index == 0:
            ax.set_ylabel("Memory (Kilobytes)") 
        elif row_index == 2 and col_index == 0:
            ax.set_ylabel("Runtime (Seconds)")

        
        if row_index == 0:
            ax.set_ylim(0, 1)
        if row_index == 1:
            ax.set_yscale("log")
            ax.set_ylim(1e2, 1e6)
        if row_index == 2:
            ax.set_yscale("log")
            ax.set_ylim(1e-5, 1e4)
            # ax.set_yscale("log")

        if row_index == 2 and col_index == 0:
            ax.set_xlabel("Proportion")
        elif row_index == 2 and col_index == 1: 
            ax.set_xlabel("G size")
        elif row_index == 2 and col_index == 2:
            ax.set_xlabel("Average Degree")
        # Show legend only for the top-right plot (or wherever you prefer)
        if row_index == 0 and col_index == 2:
            ax.legend()
        if col_index != 0:
            ax.set_yticklabels([])
            ax.set_yticks([])

# Final layout and display
plt.tight_layout()
plt.savefig("plots/gnp.png",dpi=300)