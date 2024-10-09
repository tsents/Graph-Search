import pandas as pd
import matplotlib.pyplot as plt
import os
import glob
import numpy as np

# Define the directory containing the CSV files
directory_path = 'dat/flybrain-0.4sparse/'
depth_file_pattern = os.path.join(directory_path, 'depth*.csv')
branching_file_pattern = os.path.join(directory_path, 'branching*.csv')

# Initialize empty lists to store DataFrames
depth_dataframes = []
branching_dataframes = []

# Mapping of filenames to legend names for depth files
depth_legend_mapping = {
    'depth0.csv': 'Our prior based on S',
    'depth1.csv': 'Our prior based on G',
    'depth2.csv': 'Gready prior based on G',
    'depth3.csv': 'Random prior',
    'depth4.csv': 'Gready prior based on S',
}

branching_legend_mapping = {
    'branching0.csv': 'Our prior based on S',
    'branching1.csv': 'Our prior based on G',
    'branching2.csv': 'Gready prior based on G',
    'branching3.csv': 'Random prior',
    'branching4.csv': 'Gready prior based on S',
}


deg_names = {
    'deg0.txt': 'Our prior based on S',
    'deg1.txt': 'Our prior based on G',
    'deg2.txt': 'Gready prior based on G',
    'deg3.txt': 'Random prior',
    'deg4.txt': 'Gready prior based on S',
}

# Loop through each depth file matching the pattern
for file_path in glob.glob(depth_file_pattern):
    largest_df = None
    max_rows = 0
    
    with open(file_path, 'r') as file:
        current_df = []
        header = None
        
        for line in file:
            stripped_line = line.strip()
            
            if stripped_line.startswith('Depth'):
                if current_df and header is not None:
                    temp_df = pd.DataFrame(current_df, columns=header)
                    if len(temp_df) > max_rows:
                        largest_df = temp_df
                        max_rows = len(temp_df)
                
                header = stripped_line.split(',')
                current_df = []
            elif stripped_line:
                current_df.append(stripped_line.split(','))
        
        if current_df and header is not None:
            temp_df = pd.DataFrame(current_df, columns=header)
            if len(temp_df) > max_rows:
                largest_df = temp_df

    largest_df['Depth'] = pd.to_numeric(largest_df['Depth'])
    largest_df['Time'] = pd.to_timedelta(largest_df['Time'])  # Convert Time to timedelta
    largest_df.sort_values(by='Depth', inplace=True)
    largest_df = largest_df[largest_df['Depth'] >= 0]

    if largest_df is not None:
        depth_dataframes.append((largest_df, os.path.basename(file_path)))

# Loop through each branching factor file matching the pattern
for file_path in glob.glob(branching_file_pattern):
    df = pd.read_csv(file_path)
    df['Depth'] = pd.to_numeric(df['Depth'])  # Assuming 'Depth' column exists
    df['BranchingFactor'] = pd.to_numeric(df['BranchingFactor'])  # Assuming 'Factor' column exists
    df['BranchingFactor'] = np.log2(df['BranchingFactor']).rolling(window=100, min_periods=1).mean()

    df.sort_values(by='Depth', inplace=True)
    branching_dataframes.append((df, os.path.basename(file_path)))

# Step 3: Create subplots for both sets of data
fig, (ax1, ax2,ax3) = plt.subplots(3, 1, figsize=(15, 10))

# Plot Time vs Depth
for df, file_name in depth_dataframes:
    label = depth_legend_mapping.get(file_name, file_name)
    ax1.plot(df['Depth'], df['Time'].dt.total_seconds(), label=label)  # Plot time in seconds

    max_x = df['Depth'].max()
    ax1.axvline(x=max_x, linestyle='dotted', color='gray')

ax1.set_xscale('log')
ax1.set_title('Depth vs Time')
ax1.set_xlabel('Depth')
ax1.set_ylabel('Time (seconds)')
ax1.legend()
ax1.grid(True)

# Plot Branching Factor vs Depth
for df, file_name in branching_dataframes:
    label = branching_legend_mapping.get(file_name, file_name)
    ax2.plot(df['Depth'], df['BranchingFactor'], label=label)

ax2.set_xscale('log')
ax2.set_title('Branching Factor vs Depth in Log Scale')
ax2.set_xlabel('Depth')
ax2.set_ylabel('Branching Factor')
ax2.legend()
ax2.grid(True)

critical_depths = []

for df, file_name in branching_dataframes:
    half_max_depth = df['Depth'].max() / 2
    critical_section = df[df['Depth'] > half_max_depth]
    
    # Find the point where the branching factor first exceeds 0
    first_above_zero = critical_section[critical_section['BranchingFactor'] > 0]
    
    if not first_above_zero.empty:
        critical_depths.append(first_above_zero['Depth'].min())

# Get the minimum critical depth across all dataframes
if critical_depths:
    critical_depth = min(critical_depths)
else:
    critical_depth = 0  # fallback if no critical depth found

for df, file_name in branching_dataframes:
    label = branching_legend_mapping.get(file_name, file_name)
    zoomed_data = df[df['Depth'] >= critical_depth]
    ax3.plot(zoomed_data['Depth'], zoomed_data['BranchingFactor'], label=label)

ax3.set_xscale('log')
ax3.set_title('Zoomed Branching Factor vs Depth (Post-Critical Section)')
ax3.set_xlabel('Depth')
ax3.set_ylabel('Branching Factor')
ax3.legend()
ax3.grid(True)

from matplotlib.tri import Triangulation, TriAnalyzer
import matplotlib.ticker as mticker

# My axis should display 10⁻¹ but you can switch to e-notation 1.00e+01
def log_tick_formatter(val, pos=None):
    return f"$10^{{{int(val)}}}$"  # remove int() if you don't use MaxNLocator
    # return f"{10**val:.2e}"      # e-Notation

deg_file_pattern = os.path.join(directory_path, 'deg*.txt')

num_files = len(glob.glob(deg_file_pattern))
fig, axes = plt.subplots(nrows=1, ncols=num_files, figsize=(15, 5), subplot_kw={'projection': '3d'})


# # Iterate over files matching the pattern
for file_name, ax4 in zip(glob.glob(deg_file_pattern),axes):
    data_dict = {}
    with open(file_name, 'r') as file:
        for line_number, line in enumerate(file):
            if line_number % 50 == 0 or (line_number > 8500 and line_number % 5 == 0) or (line_number > 9500) :
                if line.strip():  # Ignore empty lines
                    line = line[4:-2]  # Remove "map[" from start and "]" from end
                    pairs = line.split()  # Split by whitespace
                    x_value = 10000 - sum(float(pair.split(':')[1]) for pair in pairs)  # Sum of Z values for X
                    pair_data = []
                    for pair in pairs:
                        split_pair = pair.split(':')
                        key = float(split_pair[0]) 
                        value = float(split_pair[1])
                        pair_data.append((key,value)) 
                    data_dict[x_value] = pair_data # there was a bug that there where multiple instences of the same depth, this is why we need this

    x = []
    y = []
    z = []

    for x_val, arr in data_dict.items():
        for (y_val,z_val) in arr:
            x.append(x_val)
            y.append(y_val)  
            z.append(z_val)

    y = np.log10(y)
    z = np.log10(z)
    x = np.array(x,dtype=int)
    y = np.array(y,dtype=float)
    z = np.array(z,dtype=float)

    triang = Triangulation(x, y)


    tri_analyzer = TriAnalyzer(triang)
    valid_triangles = []
    for i in range(len(triang.triangles)):
        tri_vertices = triang.triangles[i]
        y_coords = y[tri_vertices]
        x_coords = y[tri_vertices]
        
        y_distance = np.max(y_coords) - np.min(y_coords)
        x_distance = np.max(x_coords) - np.min(x_coords)

        if (x_distance < 50 and y_distance < 0.5) or (x[tri_vertices] > 9500).all():
            valid_triangles.append(i)


    filtered_triangles = triang.triangles[valid_triangles]

    ax4.plot_trisurf(x,y,z,triangles=filtered_triangles, cmap='viridis')


    ax4.set_xlim3d(0,11000)
    ax4.set_ylim3d(max(y),min(y))
    ax4.set_title(deg_names[os.path.basename(file_name)])
    ax4.zaxis.set_major_formatter(mticker.FuncFormatter(log_tick_formatter))
    ax4.zaxis.set_major_locator(mticker.MaxNLocator(integer=True))
    ax4.yaxis.set_major_formatter(mticker.FuncFormatter(log_tick_formatter))
    ax4.yaxis.set_major_locator(mticker.MaxNLocator(integer=True))


# # Adjust layout and show the plots
plt.tight_layout()
plt.show()