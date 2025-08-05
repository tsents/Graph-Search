import networkx as nx
import random
import re
import sys
from pathlib import Path
import os
from collections import defaultdict
import itertools

def get_bfs_subgraph(G, size_ratio=0.05):
    n = G.number_of_nodes()
    target_size = int(n * size_ratio)
    start_node = random.choice(list(G.nodes()))
    bfs_nodes = list(nx.bfs_tree(G, start_node).nodes)
    bfs_nodes = bfs_nodes[:target_size]
    G_sub = G.subgraph(bfs_nodes).copy()
    return G_sub


def build_networkx_graph(labels_file, edges_file, is_directed=False):
    G = nx.DiGraph() if is_directed else nx.Graph()
    node_to_label = {}

    with open(edges_file, 'r') as f:
        for line in f:
            u, v = re.split(r'[, \t]+', line.strip())
            G.add_edge(int(u), int(v))

    with open(labels_file, 'r') as f:
        for line in f:
            node, label = re.split(r'[, \t]+', line.strip())
            node_to_label[int(node)] = label

    nx.set_node_attributes(G, node_to_label, name='label')
    return G


def save_graph_to_files(graph, path, prefix):
    save_dir = path / prefix  # e.g., path/S
    save_dir.mkdir(parents=True, exist_ok=True)

    if prefix == "g_induced":
        prefix = "g"

    nodes_file = save_dir / f"{prefix}.node_labels"
    edges_file = save_dir / f"{prefix}.edges"

    # Step 1: Reindex node IDs
    old_nodes = list(graph.nodes())
    old_to_new = {old_id: new_id for new_id, old_id in enumerate(old_nodes)}

    # Step 2: Write nodes with labels
    with open(nodes_file, "w") as f:
        for old_id in old_nodes:
            if "label" in graph.nodes[old_id]:
                label = graph.nodes[old_id]["label"]
                f.write(f"{old_to_new[old_id]},{label}\n")

    # Step 3: Write edges using new IDs
    with open(edges_file, "w") as f:
        for u, v in graph.edges():
            f.write(f"{old_to_new[u]},{old_to_new[v]}\n")


def filter_edges_file_keep_only_induced(edges_file_g, S, output_file):
    nodes_S_set = set(S.nodes())
    induced_edges = set(S.edges())
    with open(edges_file_g, 'r') as fin, open(output_file, 'w') as fout:
        for line in fin:
            if not line.strip():
                continue
            u_str, v_str = line.strip().split(',')
            u, v = int(u_str), int(v_str)

            # Keep edge if:
            # - at least one node NOT in S
            # OR
            # - both in S AND edge is in induced edges
            if (u not in nodes_S_set or v not in nodes_S_set) or ((u, v) in induced_edges or (v, u) in induced_edges):
                fout.write(line)
            # else: skip this edge (remove from file)


def plant_rainbow_clique(folder_path):
    # File paths
    node_file = folder_path / "g" / "g.node_labels"
    edge_file = folder_path / "g" / "g.edges"

    # Step 1: Load node labels
    node_to_label = {}
    label_to_nodes = defaultdict(list)

    with open(node_file, "r") as f:
        for line in f:
            node, label = line.strip().split(",")
            node = int(node)
            label = int(label)
            node_to_label[node] = label
            label_to_nodes[label].append(node)

    # Step 2: Load original graph
    G = nx.Graph()
    G.add_nodes_from(node_to_label.keys())

    with open(edge_file, "r") as f:
        for line in f:
            u, v = map(int, line.strip().split(","))
            G.add_edge(u, v)

    # Step 3: Pick one node per label
    rainbow_clique_nodes = []
    for label in sorted(label_to_nodes.keys()):
        node = random.choice(label_to_nodes[label])
        rainbow_clique_nodes.append(node)

    # Step 4: Add edges to make the clique
    for u, v in itertools.combinations(rainbow_clique_nodes, 2):
        G.add_edge(u, v)

    # Step 5: Save to g_for_clique/
    out_folder = folder_path / "g_for_clique"
    os.makedirs(out_folder, exist_ok=True)

    # Save node labels
    with open(out_folder / "g.node_labels", "w") as f:
        for node in sorted(G.nodes()):
            f.write(f"{node},{node_to_label[node]}\n")

    # Save edges
    with open(out_folder / "g.edges", "w") as f:
        for u, v in G.edges():
            f.write(f"{u},{v}\n")

    print(f"Rainbow clique of size {len(rainbow_clique_nodes)} planted and saved to {out_folder}")


if __name__ == "__main__":
    if len(sys.argv) < 2 or len(sys.argv) > 3:
        print("Usage: python run.py <folder_path> [--induced]")
        sys.exit(1)

    folder_path = Path(sys.argv[1])
    induced = False

    if len(sys.argv) == 3:
        if sys.argv[2] == "--induced":
            induced = True
        else:
            print(f"Unknown option: {sys.argv[2]}")
            print("Usage: python run.py <folder_path> [--induced]")
            sys.exit(1)

    graph_nodes = folder_path / "g" / "g.node_labels"
    graph_edges = folder_path / "g" / "g.edges"
    new_graph_path = folder_path / "g_induced" / "g.edges"
    os.makedirs(new_graph_path.parent, exist_ok=True)
    temp_graph = build_networkx_graph(graph_nodes, graph_edges)
    temp_s = get_bfs_subgraph(temp_graph, 0.05)
    print("a")
    if induced:
        filter_edges_file_keep_only_induced(graph_edges, temp_s, new_graph_path)
        print("b")
        # save_graph_to_files(temp_graph, folder_path, "g_induced")
    save_graph_to_files(temp_s, folder_path, "s")

