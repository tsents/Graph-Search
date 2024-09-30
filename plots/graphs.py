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

# Load the branching CSV files
df1 = pd.read_csv('dat/IMDB/branching0.csv')
df2 = pd.read_csv('dat/IMDB/branching2.csv')
df3 = pd.read_csv('dat/IMDB/branching3.csv')
df4 = pd.read_csv('dat/IMDB/branching4.csv')

# Ensure the Depth column is treated as numeric
for df in [df1, df2, df3, df4]:
    df['Depth'] = pd.to_numeric(df['Depth'])

# Read the hair data
hair_array1 = read_array_from_file('dat/IMDB/hair0.txt')
hair_depth1 = np.arange(len(hair_array1))
hair_array2 = read_array_from_file('dat/IMDB/hair2.txt')
hair_depth2 = np.arange(len(hair_array2))
hair_array3 = read_array_from_file('dat/IMDB/hair3.txt')
hair_depth3 = np.arange(len(hair_array3))
hair_array4 = read_array_from_file('dat/IMDB/hair4.txt')
hair_depth4 = np.arange(len(hair_array4))

# Create a figure with three rows of subplots
fig, axes = plt.subplots(2, 4, figsize=(15, 15))

# Process and plot the branching data
for i, (df, hair_array, hair_depth) in enumerate(zip([df1, df2, df3, df4], 
                                                      [hair_array1, hair_array2, hair_array3, hair_array4],
                                                      [hair_depth1, hair_depth2, hair_depth3, hair_depth4])):
    
    # Smooth the branching factor
    window_size = 1000
    df['smoothed_values'] = np.log2(df['BranchingFactor']).rolling(window=window_size, min_periods=1).mean()

    # Plot branching data
    axes[0, i].plot(df['Depth'], df['smoothed_values'], label='Branching Factor', linewidth=2)
    
    # Plot the hair data on the same subplot with a secondary y-axis
    ax2 = axes[0, i].twinx()
    ax2.plot(hair_depth, hair_array, color='green', label='Hairs Left (Normalized)', linewidth=1)
    ax2.fill_between(hair_depth, hair_array, alpha=0.3, color='green')
    ax2.set_ylim(0, 1)
    ax2.set_ylabel('Hairs Left')
    
    # Set titles and labels
    axes[0, i].set_title(f'Branching Plot {i+1}')
    axes[0, i].set_xlabel('Depth')
    axes[0, i].set_ylabel('Branching Factor (log scale)')

# Read the calls vs depth data for each depth dataset
depth_data_files = [
    'dat/IMDB/depth0.csv',
    'dat/IMDB/depth2.csv',
    'dat/IMDB/depth3.csv',
    'dat/IMDB/depth4.csv'
]

# Initialize lists to store depth values for scaling
all_depths = []
all_times = []

# Plotting calls vs depth
for i, depth_data_file in enumerate(depth_data_files):
    print(i)
    depth_data = pd.read_csv(depth_data_file)  # Assuming each CSV has 'Calls', 'Depth', and 'Time' columns
    
    depth_data.sort_values(by='Depth', inplace=True)
    
    # Extract the x and y values for Calls
    x_values_calls = depth_data['Calls'].tolist()
    y_values_depth = depth_data['Depth'].tolist()
    
    # Append depth values for y-axis scaling
    all_depths.extend(y_values_depth)

    # Plot the sorted data
    axes[1, i].plot(x_values_calls, y_values_depth, marker='o')
    axes[1, i].set_xlabel('Time')
    axes[1, i].set_ylabel('Depth')
    axes[1, i].grid(True)

# Set common y-axis limits for the calls vs depth plots
y_min = min(all_depths)
y_max = max(all_depths)
for ax in axes[1, :]:
    ax.set_ylim(y_min, y_max)

plt.show()
# Plotting Time vs Depth
# for i, depth_data_file in enumerate(depth_data_files):
#     print(i)
#     depth_data = pd.read_csv(depth_data_file)
    
#     # Extract the x and y values for Time
#     x_values_time = depth_data['Time'].tolist()
#     y_values_depth = depth_data['Depth'].tolist()
    
#     # Append depth values for y-axis scaling
#     all_times.extend(x_values_time)

#     # Plot the sorted data
#     axes[2, i].plot(x_values_time, y_values_depth, marker='o', color='orange')
#     axes[2, i].set_xlabel('Time')
#     axes[2, i].set_ylabel('Depth')
#     axes[2, i].set_title(f'Time vs Depth {i+1}')
#     axes[2, i].grid(True)

# # Set common y-axis limits for the time vs depth plots
# for ax in axes[2, :]:
#     ax.set_ylim(y_min, y_max)

# # Adjust layout

# plt.savefig('plot.png', dpi=300, bbox_inches='tight')
