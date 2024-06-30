import json

#convert the map format into json
def convert_string_to_dict(s):
    s = s.replace("map[", "").replace("]", "")
    pairs = s.split(" ")
    return {pair.split(":")[0]: pair.split(":")[1] for pair in pairs}

def convert_file_to_json(input_file, output_file):
    with open(input_file, "r") as f:
        lines = f.readlines()
    dicts = [convert_string_to_dict(line) for line in lines]
    with open(output_file, "w") as f:
        json.dump(dicts, f)

import os

directory = 'important outputs'

for filename in os.listdir(directory):
    if filename.endswith(".txt"):
        print(os.path.join(directory, filename))
        convert_file_to_json(os.path.join(directory, filename),os.path.join('output', filename + '.json'))

