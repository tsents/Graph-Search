import pandas as pd
import matplotlib.pyplot as plt
import numpy as np
# Initialize an empty list to store the data for each file
data_list = []

# List of file names to loop through
files = [f'dat/IMDB-new-dat/factor{i}.csv' for i in range(5)]

# Create a list of colors for distinct lines
colors = ['b', 'g', 'r', 'c', 'm']  # You can customize the colors

# Plot each file separately
plt.figure(figsize=(10, 6))

for i, file_path in enumerate(files):
    # Load the data for each file
    df = pd.read_csv(file_path)
    
    # Check if 'Depth' and 'Factor' columns exist in the dataframe
    if 'Depth' in df.columns and 'Factor' in df.columns:
        # Group by 'Depth' and calculate the mean of 'Factor'
        average_factors = df.groupby('Depth')['Factor'].mean().reset_index()
        average_factors['Transformed'] = -np.log10(average_factors['Factor'])
        average_factors = average_factors[df['Depth'] > 39000]
        average_factors['FeatureCost'] = average_factors.apply(lambda row: average_factors[average_factors['Depth'] > row['Depth']]['Transformed'].mean(), axis=1)

        plt.scatter(average_factors['Transformed'], average_factors['FeatureCost'], 
                 label=f'Factor {i}', color=colors[i], marker='o', linestyle='-', linewidth=2)

# Customize the plot
plt.title('Culmative Reduction Factor vs Depth for Different Files')
plt.xlabel('Depth')
plt.ylabel('Culmative Reduction Factor')
plt.legend(title='Files')  # Show legend to identify each line
plt.grid(True)

# Show the plot
plt.show()
