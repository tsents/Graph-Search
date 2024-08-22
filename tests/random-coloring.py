import random

# Define the function to create the file
def create_file_with_random_numbers(filename, n,c):
    with open(filename, 'w') as file:
        for i in range(1, n + 1):
            file.write(f"{i} {random.randint(1, c)}\n")

create_file_with_random_numbers("numbers_with_random.node_labels", 16386,5)

# Print a success message
print("File 'numbers_with_random.txt' was successfully created with numbers from 1 to 3,774,768 and random numbers.")
