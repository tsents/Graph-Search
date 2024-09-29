import matplotlib.pyplot as plt

# Use a generator to read the file in chunks to handle large data
def read_large_file_in_chunks(file_path, chunk_size=1024):
    with open(file_path, 'r') as file:
        while True:
            chunk = file.read(chunk_size)
            if not chunk:
                break
            yield chunk

# Initialize lists for x and y values
x_values = []
y_values = []

# Read the file and process the data in chunks
bracket_data = ""
for chunk in read_large_file_in_chunks('dat/depth2.txt'):
    bracket_data += chunk
    while '[' in bracket_data and ']' in bracket_data:
        data = bracket_data[bracket_data.find('[')+1:bracket_data.find(']')]
        bracket_data = bracket_data[bracket_data.find(']')+1:]

        # Split the data into pairs
        pairs = data.split()

        # Iterate over each pair and split into x and y
        for pair in pairs:
            y, x = pair.split(':')
            x_values.append(int(x))
            y_values.append(int(y))

# Plot the graph
plt.plot(x_values, y_values, marker='o')
plt.xlabel('number of calls')
plt.ylabel('depth')
plt.title('progression through time')
plt.grid(True)
plt.show()
