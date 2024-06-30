import os
import json

def check_inclusion(ground_truth_dir, output_dir):
    # Loop over all files in the ground truth directory
    for gt_file in os.listdir(ground_truth_dir):
        if gt_file.endswith(".json"):
            # Construct the corresponding output file name
            output_file = "output" + gt_file.replace("graph_", "").replace("-", "_").replace(".json", ".txt.json")
            
            # Construct the full paths to the ground truth file and the output file
            gt_path = os.path.join(ground_truth_dir, gt_file)
            output_path = os.path.join(output_dir, output_file)
            
            # Load the ground truth and output data
            with open(gt_path, 'r') as f:
                gt_data = json.load(f)
            with open(output_path, 'r') as f:
                output_data = json.load(f)
            
            # Check if all dictionaries in the ground truth data are included in the output data
            for gt_dict in gt_data:
                if gt_dict not in output_data:
                    print(f"The ground truth file {gt_file} is not fully included in the output file {output_file}")
                    break
            else:
                print(f"The ground truth file {gt_file} is fully included in the output file {output_file}")

# Call the function to check the inclusion of ground truths in the output
check_inclusion('ground_truth', 'output')

