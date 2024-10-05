import matplotlib.pyplot as plt
import numpy as np
from matplotlib.colors import LogNorm

def parse_file(filename):
    data = []
    with open(filename, 'r') as f:
        for line in f:
            # Split the line to extract the dictionary
            parts = line.split('map[')
            if len(parts) > 1:
                dict_str = parts[1].strip().rstrip(']')
                dict_items = dict_str.split()
                row_data = {}
                for item in dict_items:
                    key, value = item.split(':')
                    row_data[int(key)] = int(value)
                data.append(row_data)
    return data

def create_heatmap(data):
    num_rows = len(data)
    num_cols = max(max(row.keys()) for row in data) if data else 0  # Find max column index
    num_cols += 1
    heatmap = np.zeros((num_rows, num_cols))

    # Populate the heatmap
    for i, row in enumerate(data):
        for col, value in row.items():
            heatmap[i, col] = value  # Adjust for zero-based index

    # Scale normalization: Divide each row by its maximum value
    # row_maxes = heatmap.max(axis=0, keepdims=True)
    # heatmap = np.divide(heatmap, row_maxes, where=row_maxes!=0)  # Avoid division by zero

    # Plotting the normalized heatmap
    plt.imshow(heatmap.T, cmap='hot', interpolation='nearest', aspect='auto')
    plt.colorbar(label='Normalized Count')
    plt.xlabel('Rows')  # Updated label
    plt.ylabel('Columns')  # Updated label
    plt.title('Scale Normalized Heatmap')
    plt.xlim(1, num_rows)  # Set x-limits
    plt.ylim(1, num_cols)  # Set y-limits to match the transposed dimensions
    plt.gca().invert_yaxis()  # Invert y-axis
    plt.show()

if __name__ == "__main__":
    filename = 'dat/deg'  # Replace with your actual filename
    data = parse_file(filename)
    create_heatmap(data)
