import numpy as np
import matplotlib.pyplot as plt

# Function to read array from a text file
def read_array_from_file(file_path):
    with open(file_path, 'r') as file:
        data = file.read()
        data = data.strip('[]').split()
        array = [27720 - int(item.strip('[],')) for item in data]
        leftover = np.flip(np.arange(len(array)))
        array = array/leftover
    return array

# Function to plot the array
def plot_array(array):
    x = np.arange(len(array))
    y = array
    plt.plot(x, y, marker='o')
    plt.xlabel('depth')
    plt.ylabel('hairs left')
    plt.title('Array Plot')
    plt.grid(True)
    plt.show()

# Example usage
file_path = 'hair.txt'
array = read_array_from_file(file_path)
plot_array(array)
