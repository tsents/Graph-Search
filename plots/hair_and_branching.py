import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# Function to read array from a text file
def read_array_from_file(file_path):
    with open(file_path, 'r') as file:
        data = file.read()
        data = data.strip('[]').split()
        array = [27720 - int(item.strip('[],')) for item in data]
        leftover = np.flip(np.linspace(pow(10,5) - len(array),pow(10,5),len(array)))
        print(len(leftover),len(array))
        array = array/leftover
    return array

# Load the CSV files
df1 = pd.read_csv('IMDB-branching.csv')
df2 = pd.read_csv('IMDB-branching2.csv')

# Ensure the Depth column is treated as numeric
df1['Depth'] = pd.to_numeric(df1['Depth'])
df2['Depth'] = pd.to_numeric(df2['Depth'])

# Read the hair data
hair_array = read_array_from_file('IMDB-hair.txt')
hair_depth = np.arange(len(hair_array))

hair_array2 = read_array_from_file('IMDB-hair2.txt')
hair_depth2 = np.arange(len(hair_array2))

# Create a figure with two subplots
fig, axes = plt.subplots(1, 2, figsize=(15,5))

# Process and plot the branching data
df1 = df1.dropna()
df2 = df2.dropna()

window_size = 1000

df1['smoothed_values'] = np.log2(df1['BranchingFactor'])
df2['smoothed_values'] = np.log2(df2['BranchingFactor'])

df1['smoothed_values'] = df1['smoothed_values'].rolling(window=window_size, min_periods=1).mean()
df2['smoothed_values'] = df2['smoothed_values'].rolling(window=window_size, min_periods=1).mean()

axes[0].plot(df1['Depth'], df1['smoothed_values'], label='Branching Factor', linewidth=2)
axes[1].plot(df2['Depth'], df2['smoothed_values'], label='Branching Factor', linewidth=2)

# Plot the hair data on both subplots with a secondary y-axis
ax2_0 = axes[0].twinx()
ax2_0.plot(hair_depth, hair_array, color='green', label='Hairs Left (Normalized)', linewidth=1)
ax2_0.fill_between(hair_depth, hair_array, alpha=0.3, color='green')
ax2_0.set_ylim(0, 1)
ax2_0.set_ylabel('Hairs Left')

ax2_1 = axes[1].twinx()
ax2_1.plot(hair_depth2, hair_array2, color='green', label='Hairs Left (Normalized)', linewidth=1)
ax2_1.fill_between(hair_depth2, hair_array2, alpha=0.3, color='green')
ax2_1.set_ylim(0, 1)
ax2_1.set_ylabel('Hairs Left')

# Set titles for each subplot
axes[0].set_title('our prior')
axes[1].set_title('naive prior')

# Set labels for each subplot
axes[0].set_xlabel('Depth')
axes[0].set_ylabel('Branching Factor (log scale)')
axes[1].set_xlabel('Depth')
axes[1].set_ylabel('Branching Factor (log scale)')

# Add legends to each subplot
lines_labels_0 = [axes[0].get_legend_handles_labels(), ax2_0.get_legend_handles_labels()]
lines_0, labels_0 = [sum(lol, []) for lol in zip(*lines_labels_0)]

lines_labels_1 = [axes[1].get_legend_handles_labels(), ax2_1.get_legend_handles_labels()]
lines_1, labels_1 = [sum(lol, []) for lol in zip(*lines_labels_1)]

# Create combined legends for each pair of subplots
axes[0].legend(lines_0, labels_0, loc='upper left')
axes[1].legend(lines_1, labels_1, loc='upper left')

# Optionally, remove individual legends if they are not needed
ax2_0.legend().remove()
ax2_1.legend().remove()

# Display the plot
plt.tight_layout()
plt.show()
