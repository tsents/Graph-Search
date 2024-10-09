import pandas as pd
import matplotlib.pyplot as plt
import os
import glob
import numpy as np

# Define the directory containing the CSV files
directory_path = 'dat/bn-human-Jung2015_M87125334/'
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

from matplotlib import cm
from scipy.interpolate import griddata

deg_file_pattern = os.path.join(directory_path, 'deg*.txt')

num_files = len(glob.glob(deg_file_pattern))
fig, axes = plt.subplots(nrows=1, ncols=num_files, figsize=(15, 5), subplot_kw={'projection': '3d'})


# # Iterate over files matching the pattern
for file_name, ax4 in zip(glob.glob(deg_file_pattern),axes):
    data_dict = {}
    with open(file_name, 'r') as file:
        for line_number, line in enumerate(file):
            if line_number % 100 == 0:
                if line.strip():  # Ignore empty lines
                    line = line[4:-2]  # Remove "map[" from start and "]" from end
                    pairs = line.split()  # Split by whitespace
                    x_value = 100000 - sum(float(pair.split(':')[1]) for pair in pairs)  # Sum of Z values for X

                    for pair in pairs:
                        split_pair = pair.split(':')
                        key = float(split_pair[0])  # This will be the Y value (name)
                        value = float(split_pair[1])  # Convert Z value to float
                        # print(x_value,key,value)

                        # Store the latest entry in the dictionary
                        data_dict[key] = (x_value, value)  # Override if key exists
    # data_dict = [(x_value, value) for (x_value, value) in data_dict if x_value % 100 == 0]

    x = []
    y = []
    z = []

    for key, (x_val, z_val) in data_dict.items():
        x.append(x_val)
        y.append(key)  # Y is the key (name)
        z.append(z_val)


    ax4.scatter(x,y,z)
    # grid_x, grid_y = np.mgrid[min(x):max(x), min(y):max(y)]

    # # Interpolate the data
    # grid_z = griddata((x, y), z, (grid_x, grid_y), method='linear')

    # ax4.plot_surface(grid_x, grid_y, grid_z, cmap='viridis')

#     for key, (x_val, z_val) in data_dict.items():
#         x.append(x_val)
#         y.append(key)  # Y is the key (name)
#         z.append(z_val)

#     # Convert lists to numpy arrays
#     x = np.array(x,dtype=float)
#     y = np.array(y,dtype=float)
#     z = np.array(z,dtype=float)

#     # Create a grid for surface plot
#     # Use unique Y values to create a meshgrid
#     # y_unique = np.unique(y)
#     # x_unique = np.unique(x)
    
#     # # Create a grid for X and Y
#     # X, Y = np.meshgrid(x_unique, y_unique)
#     ax4.plot_trisurf(x, y, z, cmap=cm.coolwarm,
#                         linewidth=1, antialiased=False)
#     # # Initialize Z with NaNs for surface plotting
#     # Z = np.full(X.shape,np.nan)

#     # for i, unique_y in enumerate(y_unique):
#     #     for j, unique_x in enumerate(x_unique):
#     #         # Find corresponding Z values
#     #         z_values = [data_dict[key][1] for key in data_dict if key == unique_y and data_dict[key][0] == unique_x]
#     #         if z_values:
#     #             Z[i, j] = z_values[-1]  # Take the last value if duplicates exist


#     # # Z = np.log1p(Z)
#     # # Y = np.log2(Y)
#     # ax4.plot_surface(X, Y, Z, cmap=cm.coolwarm,
#     #                     linewidth=1, antialiased=False)

#     # # Create a mesh grid for surface plotting
#     # X_unique = np.unique(X)
#     # Y_unique = np.unique(Y)
#     # X_grid, Y_grid = np.meshgrid(X_unique, Y_unique)
#     # Z_grid = np.zeros_like(X_grid)

#     # for i in range(len(X)):
#     #     idx_x = np.where(X_unique == X[i])[0][0]
#     #     idx_y = np.where(Y_unique == Y[i])[0][0]
#     #     Z_grid[idx_y, idx_x] = Z[i]

#     # Plot the surface
#     # ax4.plot_surface(X, Y, Z, cmap=cm.coolwarm,
#     #                  linewidth=0, antialiased=False)

#     # Set titles and labels for each subplot
#     ax4.set_title(f'3D Plot of Degree Data from {file_name}')
#     ax4.set_xlabel('Depth')
#     ax4.set_ylabel('Split[0]')
#     ax4.set_zlabel('Split[1]')
#     ax4.grid(True)

#     # Flip the Y-axis
#     ax4.set_ylim(ax4.get_ylim()[::-1])  # Reverse the Y-axis direction

# # Adjust layout and show the plots
plt.tight_layout()
plt.show()