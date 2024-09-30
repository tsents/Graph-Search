import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt
import numpy as np

# Load the CSV files
df1 = pd.read_csv('dat/IMDB/branching0.csv')
df2 = pd.read_csv('dat/IMDB/branching2.csv')
df3 = pd.read_csv('dat/IMDB/branching3.csv')
df4 = pd.read_csv('dat/IMDB/branching4.csv')

# Ensure the Depth column is treated as numeric
df1['Depth'] = pd.to_numeric(df1['Depth'])
df2['Depth'] = pd.to_numeric(df2['Depth'])
df3['Depth'] = pd.to_numeric(df3['Depth'])
df4['Depth'] = pd.to_numeric(df4['Depth'])

# window_size = 5
# df1['BranchingFactor_Smoothed'] = df1['BranchingFactor'].rolling(window=window_size).mean()
# df2['BranchingFactor_Smoothed'] = df2['BranchingFactor'].rolling(window=window_size).mean()

sns.set_theme()

# Create a figure with three subplots
fig, axes = plt.subplots(1, 4, figsize=(15,5))

# Plot the data using seaborn regplot with trend line and variance
# sns.regplot(x='Depth', y='BranchingFactor', data=df1,
#             ax=axes[0], scatter=True,
#             lowess=True,
#             line_kws={'color': 'blue'},
#             scatter_kws={'alpha':0.5})
# sns.regplot(x='Depth', y='BranchingFactor', data=df2,
#             ax=axes[1], scatter=True,
#             lowess=True,
#             line_kws={'color': 'blue'},
#             scatter_kws={'alpha':0.5})
# df1['Depth'] = np.log(df1['Depth'])
# df2['Depth'] = np.log(df2['Depth'])

# df1 = df1[df1['BranchingFactor'] != 1]
# df2 = df2[df2['BranchingFactor'] != 1]

print(df2['BranchingFactor'])

df1 = df1.dropna()
df2 = df2.dropna()

window_size = 100

df1['smoothed_values'] = np.log2(df1['BranchingFactor'])
df2['smoothed_values'] = np.log2(df2['BranchingFactor'])
df3['smoothed_values'] = np.log2(df3['BranchingFactor'])
df4['smoothed_values'] = np.log2(df4['BranchingFactor'])

df1['smoothed_values'] = df1['smoothed_values'].rolling(window=window_size, min_periods=1).mean()
df2['smoothed_values'] = df2['smoothed_values'].rolling(window=window_size, min_periods=1).mean()
df3['smoothed_values'] = df3['smoothed_values'].rolling(window=window_size, min_periods=1).mean()
df4['smoothed_values'] = df4['smoothed_values'].rolling(window=window_size, min_periods=1).mean()

sns.lineplot(x='Depth',y='smoothed_values', data=df1,
            ax=axes[0])
sns.lineplot(x='Depth',y='smoothed_values', data=df2,
            ax=axes[1])
sns.lineplot(x='Depth',y='smoothed_values', data=df3,
            ax=axes[0])
sns.lineplot(x='Depth',y='smoothed_values', data=df4,
            ax=axes[1])


# Set titles for each subplot
axes[0].set_title('our prior')
axes[1].set_title('naive prior')

# # Reverse the x-axis to match the depth order
# for ax in axes:
#     ax.invert_xaxis()

# Display the plot
plt.tight_layout()
plt.show()

