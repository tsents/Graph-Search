
import csv


def convert_spaces_to_commas_inplace(filename):
    with open(filename, 'r') as f:
        lines = f.readlines()

    with open(filename, 'w') as f:
        for line in lines:
            new_line = ','.join(line.split())
            f.write(new_line + '\n')


def reindex_node_file(filename):
    """
    Reindex the left column (node IDs) of the node file so that IDs start from 0 and are consecutive.
    Returns a mapping from old node IDs to new node IDs.
    """
    with open(filename, 'r') as f:
        reader = csv.reader(f)
        rows = [row for row in reader]

    old_ids = [row[0] for row in rows]
    unique_ids = sorted(set(old_ids), key=int)
    id_map = {old_id: str(new_id) for new_id, old_id in enumerate(unique_ids)}

    # Update rows with new IDs
    new_rows = [[id_map[row[0]], row[1]] for row in rows]

    with open(filename, 'w', newline='') as f:
        writer = csv.writer(f)
        writer.writerows(new_rows)

    return id_map


def reindex_edge_file(filename, node_map, node_file):

    # Step 1: Read existing labels to determine the next available label
    existing_labels = set()
    try:
        with open(node_file, 'r') as f:
            reader = csv.reader(f)
            for row in reader:
                if len(row) > 1:
                    try:
                        existing_labels.add(int(row[1]))
                    except ValueError:
                        pass  # Skip rows with non-integer labels
    except FileNotFoundError:
        pass  # No node file yet

    next_label = max(existing_labels) + 1 if existing_labels else 0

    # Step 2: Read the edge file
    with open(filename, 'r') as f:
        reader = csv.reader(f)
        rows = [row for row in reader]

    current_max_id = max(int(v) for v in node_map.values())
    next_id = current_max_id + 1

    new_nodes = []  # List of [new_node_id, new_label]
    new_rows = []

    for row in rows:
        u, v = row[0], row[1]

        for node in (u, v):
            if node not in node_map:
                node_map[node] = str(next_id)
                new_nodes.append([str(next_id), str(next_label)])
                next_id += 1

        new_rows.append([node_map[u], node_map[v]])

    # Step 3: Overwrite the edge file with reindexed edges
    with open(filename, 'w', newline='') as f:
        writer = csv.writer(f)
        writer.writerows(new_rows)

    # Step 4: Append all new nodes (with same label) to node file
    if new_nodes:
        with open(node_file, 'a', newline='') as f:
            writer = csv.writer(f)
            writer.writerows(new_nodes)



def convert_formats(name):
    convert_spaces_to_commas_inplace(f"{name}/g/g.node_labels")
    convert_spaces_to_commas_inplace(f"{name}/g/g.edges")
    node_map = reindex_node_file(f"{name}/g/g.node_labels")
    reindex_edge_file(f"{name}/g/g.edges", node_map, f"{name}/g/g.node_labels")


if __name__ == "__main__":
    convert_formats("flybrain")
