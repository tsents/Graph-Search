import re
import math

def scientific_to_int(match):
    base, exponent = match.groups()
    base = float(base)
    exponent = int(exponent)
    return str(int(base * math.pow(10, exponent)))

def replace_scientific_notation(file_path, output_path):
    with open(file_path, 'r') as file:
        content = file.read()

    # Regular expression to match scientific notation
    scientific_notation_pattern = re.compile(r'(\d+\.\d+)[eE]\+?([+-]?\d+)')
    new_content = scientific_notation_pattern.sub(scientific_to_int, content)

    with open(output_path, 'w') as file:
        file.write(new_content)

if __name__ == "__main__":
    input_file = 'Mutag.node_labels'  # Replace with your input file path
    output_file = 'Mutag2.node_labels'  # Replace with your desired output file path
    replace_scientific_notation(input_file, output_file)
