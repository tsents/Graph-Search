import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# Function to read array from a text file
def read_array_from_file(file_path):
    with open(file_path, 'r') as file:
        data = file.read()
        data = data.strip('[]').split()
        array = [27720 - int(item.strip('[],')) for item in data]

        zero_start_index = None
        for i in range(len(array)):
            if array[i] == 27720:
                if i > 0 and array[i-1] != 27720:
                    zero_start_index = i
                    break

        leftover = np.flip(np.linspace(pow(10, 5) - len(array), pow(10, 5), len(array)))
        array = array / leftover

        if zero_start_index is not None:
            last_non_zero_value = array[zero_start_index - 1]
            for j in range(zero_start_index, len(array)):
                array[j] = last_non_zero_value
        else:
            print("No transition to zeros found after non-zero values.")

    return array

# Load the CSV files
df1 = pd.read_csv('dat/flybrain/branching0.csv')
df2 = pd.read_csv('dat/flybrain/branching2.csv')
df3 = pd.read_csv('dat/flybrain/branching3.csv')
df4 = pd.read_csv('dat/flybrain/branching4.csv')

# Ensure the Depth column is treated as numeric
df1['Depth'] = pd.to_numeric(df1['Depth'])
df2['Depth'] = pd.to_numeric(df2['Depth'])
df3['Depth'] = pd.to_numeric(df3['Depth'])
df4['Depth'] = pd.to_numeric(df4['Depth'])

# Read the hair data
hair_array1 = read_array_from_file('dat/flybrain/hair0.txt')
hair_depth1 = np.arange(len(hair_array1))
hair_array2 = read_array_from_file('dat/flybrain/hair2.txt')
hair_depth2 = np.arange(len(hair_array2))
hair_array3 = read_array_from_file('dat/flybrain/hair3.txt')
hair_depth3 = np.arange(len(hair_array3))
hair_array4 = read_array_from_file('dat/flybrain/hair4.txt')
hair_depth4 = np.arange(len(hair_array4))

# Create a figure with two subplots
fig, axes = plt.subplots(1, 4, figsize=(15, 5))

# Process and plot the branching data
df1 = df1.dropna()
df2 = df2.dropna()
df3 = df3.dropna()
df4 = df4.dropna()

window_size = 1000

df1['smoothed_values'] = np.log2(df1['BranchingFactor']).rolling(window=window_size, min_periods=1).mean()
df2['smoothed_values'] = np.log2(df2['BranchingFactor']).rolling(window=window_size, min_periods=1).mean()
df3['smoothed_values'] = np.log2(df3['BranchingFactor']).rolling(window=window_size, min_periods=1).mean()
df4['smoothed_values'] = np.log2(df4['BranchingFactor']).rolling(window=window_size, min_periods=1).mean()

axes[0].plot(df1['Depth'], df1['smoothed_values'], label='Branching Factor', linewidth=2)
axes[1].plot(df2['Depth'], df2['smoothed_values'], label='Branching Factor', linewidth=2)
axes[2].plot(df3['Depth'], df3['smoothed_values'], label='Branching Factor', linewidth=2)
axes[3].plot(df4['Depth'], df4['smoothed_values'], label='Branching Factor', linewidth=2)

# Set common Y-axis limits
y_min = min(df1['smoothed_values'].min(), df2['smoothed_values'].min(), df3['smoothed_values'].min(), df4['smoothed_values'].min())
y_max = max(df1['smoothed_values'].max(), df2['smoothed_values'].max(), df3['smoothed_values'].max(), df4['smoothed_values'].max())
for ax in axes:
    ax.set_ylim(y_min, y_max*1.1)

# Plot the hair data on both subplots with a secondary y-axis
ax2_0 = axes[0].twinx()
ax2_0.plot(hair_depth1, hair_array1, color='green', label='Hairs Left (Normalized)', linewidth=1)
ax2_0.fill_between(hair_depth1, hair_array1, alpha=0.3, color='green')
ax2_0.set_ylim(0, 1)
ax2_0.set_ylabel('Hairs Left')

ax2_1 = axes[1].twinx()
ax2_1.plot(hair_depth2, hair_array2, color='green', label='Hairs Left (Normalized)', linewidth=1)
ax2_1.fill_between(hair_depth2, hair_array2, alpha=0.3, color='green')
ax2_1.set_ylim(0, 1)
ax2_1.set_ylabel('Hairs Left')

ax2_2 = axes[2].twinx()
ax2_2.plot(hair_depth3, hair_array3, color='green', label='Hairs Left (Normalized)', linewidth=1)
ax2_2.fill_between(hair_depth3, hair_array3, alpha=0.3, color='green')
ax2_2.set_ylim(0, 1)
ax2_2.set_ylabel('Hairs Left')

ax2_3 = axes[3].twinx()
ax2_3.plot(hair_depth4, hair_array4, color='green', label='Hairs Left (Normalized)', linewidth=1)
ax2_3.fill_between(hair_depth4, hair_array4, alpha=0.3, color='green')
ax2_3.set_ylim(0, 1)
ax2_3.set_ylabel('Hairs Left')

# Set titles for each subplot
axes[0].set_title('Our Prior')
axes[1].set_title('Gready Prior based on G')
axes[2].set_title('Random Prior')
axes[3].set_title('Gready Prior based on S')


# Set labels for each subplot
for ax in axes:
    ax.set_xlabel('Depth')
    ax.set_ylabel('Branching Factor (log scale)')

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
ax2_2.legend().remove()
ax2_3.legend().remove()

# Display the plot
plt.tight_layout()
plt.show()
