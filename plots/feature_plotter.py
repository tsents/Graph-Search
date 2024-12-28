import pandas as pd
import matplotlib.pyplot as plt
import numpy as np

data_list = []

files = [f'dat/IMDB/branching{i}.csv' for i in range(5)]

colors = ['b', 'g', 'r', 'c', 'm'] 
names = [
    "Our method in S",
    "Our method in G",
    "Gready method in G",
    "Random",
    "Gready method in S",
]

plt.figure(figsize=(10, 6))

for i, file_path in enumerate(files):
    df = pd.read_csv(file_path)
    df = df[df["Depth"] > 78000]
    df['BranchingFactor'] = np.log10(df['BranchingFactor'])
    df['FeatureCost'] = df.apply(lambda row: df[df['Depth'] > row['Depth']]['BranchingFactor'].mean(), axis=1)
    print("easy")
    plt.scatter(df['BranchingFactor'], df['FeatureCost'], 
                label=names[i], color=colors[i], marker='o')

# Customize the plot
plt.title('Current cost vs Average Feature cost')
plt.xlabel('Branching Factor')
plt.ylabel('Mean Future Branching Cost')
plt.legend(title='Files')  # Show legend to identify each line
plt.grid(True)

# Show the plot
plt.show()
