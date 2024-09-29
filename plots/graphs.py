import pandas as pd
import matplotlib.pyplot as plt

# Read the CSV file
df = pd.read_csv('IMDB.csv')

# Convert Time to seconds
def convert_to_seconds(time_str):
    if 'ms' in time_str:
        return float(time_str.replace('ms', '')) / 1000
    elif 's' in time_str:
        return float(time_str.replace('s', ''))
    else:
        return float(time_str)

df['Time'] = df['Time'].apply(convert_to_seconds)

# Sort the DataFrame by Depth
df = df.sort_values(by='Depth')

# Plot Depth vs Time as a line graph
plt.figure(figsize=(10, 5))
plt.subplot(1, 2, 1)
plt.plot(df['Time'], df['Depth'], marker='o')
plt.xlabel('Time (seconds)')
plt.ylabel('Depth')
plt.title('Depth vs Time')

# Plot Depth vs Calls as a line graph
plt.subplot(1, 2, 2)
plt.plot(df['Calls'], df['Depth'], marker='o')
plt.xlabel('Calls')
plt.ylabel('Depth')
plt.title('Depth vs Calls')

# Show plots
plt.tight_layout()
plt.show()
