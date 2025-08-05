import time
import tracemalloc
import subprocess
import pandas as pd
import random
import networkx as nx
import sys
from pathlib import Path
import re
from collections import deque
from collections import defaultdict
import for_s
import signal
import os
import itertools


def run_with_measurement(func, name, g, s):
    tracemalloc.start()
    start_time = time.time()

    found = func(g, s, name)

    end_time = time.time()
    current, peak = tracemalloc.get_traced_memory()
    tracemalloc.stop()

    if found:
        process_time = round(end_time - start_time, 4)
        mem = round(peak / 1024, 2)
    else:
        process_time, mem = -1, -1

    return {
        "algorithm": name,
        "runtime_sec": process_time,
        "memory_peak_kb": mem
    }


# def get_bfs_subgraph(G, size_ratio=0.05):
#     n = G.number_of_nodes()
#     target_size = max(1, int(n * size_ratio))
#     start_node = random.choice(list(G.nodes()))
#
#     visited = set()
#     queue = deque([start_node])
#     bfs_nodes = []
#
#     while queue and len(bfs_nodes) < target_size:
#         node = queue.popleft()
#         if node in visited:
#             continue
#         visited.add(node)
#         bfs_nodes.append(node)
#         queue.extend(n for n in G.neighbors(node) if n not in visited)
#
#     G_sub = G.subgraph(bfs_nodes).copy()

# def get_bfs_subgraph(G, size_ratio=0.05):
#     n = G.number_of_nodes()
#     target_size = int(n * size_ratio)
#     start_node = random.choice(list(G.nodes()))
#     bfs_nodes = list(nx.bfs_tree(G, start_node).nodes)
#     bfs_nodes = bfs_nodes[:target_size]
#     G_sub = G.subgraph(bfs_nodes).copy()
#     print("is?", nx.is_connected(G_sub), G.number_of_nodes(), G_sub.number_of_nodes())
#     return G_sub
#
#
# # This function builds the graph from the given edge and label files
# def build_networkx_graph(labels_file, edges_file):
#     G = nx.Graph()
#     node_to_label = {}
#
#     with open(edges_file, 'r') as f:
#         for line in f:
#             u, v = re.split(r'[, \t]+', line.strip())
#             G.add_edge(int(u), int(v))
#
#     with open(labels_file, 'r') as f:
#         for line in f:
#             node, label = re.split(r'[, \t]+', line.strip())
#             node_to_label[int(node)] = label
#
#     nx.set_node_attributes(G, node_to_label, name='label')
#     return G


class TimeoutException(Exception):
    pass


def timeout_handler(signum, frame):
    raise TimeoutException()


def vf2_algo(graph, s, name):
    # 1. Measure graph building time
    start_build = time.time()
    (g_nodes, g_edges) = graph
    (s_nodes, s_edges) = s
    G = for_s.build_networkx_graph(g_nodes, g_edges)
    s_graph = for_s.build_networkx_graph(s_nodes, s_edges)
    end_build = time.time()
    build_time = end_build - start_build
    # Append build time to log file
    log_file_path = folder_path / "python_times_clq.txt"
    with open(log_file_path, "a") as f:
        f.write(f"{build_time:.4f}\n")

    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm(timeout)

    try:
        # (g_nodes, g_edges) = graph
        # (s_nodes, s_edges) = s
        # G = for_s.build_networkx_graph(g_nodes, g_edges)
        # s_graph = for_s.build_networkx_graph(s_nodes, s_edges)

        matcher = nx.algorithms.isomorphism.GraphMatcher(
            G, s_graph,
            node_match=lambda x, y: x['label'] == y['label']
        )

        for _ in matcher.subgraph_isomorphisms_iter():
            signal.alarm(0)  # cancel timeout
            return True  # found a match

    except TimeoutException:
        print(f"VF2++: Timeout after {timeout} seconds.")
    finally:
        signal.alarm(0)  # Cancel alarm
    return False


# Example C++ algorithm (as an external command)
def algo_cpp_command(g, s, name):
    results_file_velcro = Path(folder_path) / "velcro_results_clq.txt"

    if name == "VF3":
        start_build = time.time()
        g_vf3_path = folder_path / "g" / "g_vf3.grf"
        (g_nodes, g_edges) = g
        to_vf3_format(g_nodes, g_edges, g_vf3_path)
        s_vf3_path = folder_path / "s" / "s_vf3.sub.grf"
        (s_nodes, s_edges) = s
        to_vf3_format(s_nodes, s_edges, s_vf3_path)
        end_build = time.time()
        build_time = end_build - start_build
        # Append build time to log file
        log_file_path = folder_path / "vf3_format_times_clq.txt"
        with open(log_file_path, "a") as f:
            f.write(f"{build_time:.4f}\n")

        cmd = ["timeout", "-s", "SIGINT", f"{timeout}s",
               "./vf3_new", str(g_vf3_path), str(s_vf3_path), "-u", "-e", "-F"]
    else:
        cmd = [
            "timeout", "-s", "SIGINT", f"{timeout}s",  # <-- Timeout part
            "./velcro",
            "-fmt=folder",
            '-parse=%d,%d',
            "-prior=1",
            f"-out={str(results_file_velcro)}",
            g,
            s
        ]
    try:
        result = subprocess.run(cmd, check=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
        if name == "VF3":
            print(int(result.stdout.strip().split()[0]), "vf3")
            return int(result.stdout.strip().split()[0])
        elif name == "velcro":
            #return results_file_velcro.exists() and results_file_velcro.stat().st_size > 0
            if results_file_velcro.exists() and results_file_velcro.stat().st_size > 0:
                with open(results_file_velcro, 'r') as f:
                    first_line = f.readline().strip()
                    # print(first_line, "velcro_first")
                return True
            else:
                return False

    except subprocess.CalledProcessError as e:
        if e.returncode == 124:
            print(f"{name} timed out after {timeout}s.")
        else:
            print(f"{name} crashed: {e}")

        if name == "VF3" and e.stdout:
            parts = e.stdout.strip().split()
            if parts and parts[0].isdigit():
                return int(parts[0])

        elif name == "velcro":
            try:
                if results_file_velcro.exists() and results_file_velcro.stat().st_size > 0:
                    with open(results_file_velcro, 'r') as f:
                        first_line = f.readline().strip()
                        # print(first_line, "velcro_first")
                    return True
                else:
                    return False
            except Exception as file_error:
                print(f"Error reading velcro results file: {file_error}")
        return False

    finally:
        # Always clear the velcro results file after run
        if name == "velcro" and results_file_velcro.exists():
            try:
                results_file_velcro.write_text("")  # empty the fil
            except Exception as e:
                print(f"Failed to clear velcro results file: {e}")


# Main experiment runner
def main(G, s, folder):

    results = [
               run_with_measurement(algo_cpp_command, "VF3", G, s),
               run_with_measurement(algo_cpp_command, "velcro", str(folder / "g_for_clique"), str(folder / "s_clique")),
               run_with_measurement(vf2_algo, "VF2++", G, s)]

    for result in results:
        save_result_by_algorithm(result, folder)


def save_result_by_algorithm(result, folder):
    folder = Path(folder)
    folder.mkdir(parents=True, exist_ok=True)

    algo_name = result["algorithm"]
    output_csv = folder / f"{algo_name}_clique.csv"

    df = pd.DataFrame([{
        "time": result["runtime_sec"],
        "memory": result["memory_peak_kb"]
    }])

    if output_csv.exists():
        df.to_csv(output_csv, mode="a", header=False, index=False)
    else:
        df.to_csv(output_csv, index=False)

    print(f"Saved to {output_csv}")


def to_vf3_format(labels_file, edges_file, filepath_to_save):
    # Read nodes and labels
    node_labels = {}
    with open(labels_file, 'r') as f:
        for line_num, line in enumerate(f, 1):
            if line.strip() == "":
                continue
            parts = re.split(r'[, \t]+', line.strip())
            if len(parts) < 2:
                raise ValueError(f"Invalid node label format at line {line_num} in {labels_file}: '{line.strip()}'")
            node_str, label = parts[0], parts[1]
            try:
                node = int(node_str)
            except ValueError:
                raise ValueError(f"Non-integer node id at line {line_num} in {labels_file}: '{node_str}'")
            node_labels[node] = label

    # Read edges (no labels)
    edges = []
    with open(edges_file, 'r') as f:
        for line_num, line in enumerate(f, 1):
            if line.strip() == "":
                continue
            parts = re.split(r'[, \t]+', line.strip())
            if len(parts) < 2:
                raise ValueError(f"Invalid edge format at line {line_num} in {edges_file}: '{line.strip()}'")
            try:
                u, v = int(parts[0]), int(parts[1])
            except ValueError:
                raise ValueError(f"Non-integer edge nodes at line {line_num} in {edges_file}: '{line.strip()}'")
            if u != v:  # Avoid self-loops
                edges.append((u, v))

    # Check nodes continuity
    max_node_id = max(node_labels.keys())
    num_nodes = max_node_id + 1
    missing_nodes = [i for i in range(num_nodes) if i not in node_labels]
    if missing_nodes:
        raise ValueError(f"Missing node labels for nodes: {missing_nodes}")

    # Group edges by source node (undirected: add both u->v and v->u)
    edges_by_node = defaultdict(list)
    added = set()
    # Determine the next available integer label
    existing_labels = set(int(label) for label in node_labels.values())
    default_label = max(existing_labels, default=-1) + 1
    for u, v in edges:
        # if u >= num_nodes or v >= num_nodes:
        #     raise ValueError(f"Edge with invalid node id: ({u}, {v}) outside node range 0 ... {num_nodes-1}")
        # if u == v:
        #     raise ValueError(f"Self-loop edge detected: ({u}, {v})")
        for node in (u, v):
            if node not in node_labels:
                node_labels[node] = default_label
        if (u, v) not in added and (v, u) not in added:
            edges_by_node[u].append((u, v))
            # edges_by_node[v].append((v, u))
            # edges_by_node[u].append((u, v, 0))
            # edges_by_node[v].append((v, u, 0))
            added.add((u, v))
            # added.add((v, u))

    # Write output
    with open(filepath_to_save, 'w') as f:
        f.write(f"{num_nodes}\n")

        # Write node labels (tab separated)
        for node_id in range(num_nodes):
            label = node_labels.get(node_id, "0")
            f.write(f"{node_id} {label}\n")

        # Write edges grouped by node, edge lines: "src tgt<tab>label"
        for node_id in range(num_nodes):
            out_edges = edges_by_node.get(node_id, [])
            f.write(f"{len(out_edges)}\n")
            for (src, tgt) in out_edges:
                if src != node_id:
                    raise ValueError(f"Edge source node mismatch for node {node_id}: edge source {src}")
                f.write(f"{src} {tgt}\n")


def save_graph_to_files(graph, path, prefix):
    save_dir = path / prefix  # e.g., path/S
    save_dir.mkdir(parents=True, exist_ok=True)

    nodes_file = save_dir / f"{prefix}.node_labels"
    edges_file = save_dir / f"{prefix}.edges"

    # Step 1: Reindex node IDs
    old_nodes = list(graph.nodes())
    old_to_new = {old_id: new_id for new_id, old_id in enumerate(old_nodes)}

    # Step 2: Write nodes with labels
    with open(nodes_file, "w") as f:
        for old_id in old_nodes:
            label = graph.nodes[old_id].get("label", old_id)
            f.write(f"{old_to_new[old_id]},{label}\n")

    # Step 3: Write edges using new IDs
    with open(edges_file, "w") as f:
        for u, v in graph.edges():
            f.write(f"{old_to_new[u]},{old_to_new[v]}\n")


def create_clique_from_labels(label_file):
    # Read all labels from original g.node_labels
    # label_file = folder_path / "g" / "g.node_labels"
    unique_labels = set()

    with open(label_file, "r") as f:
        for line in f:
            _, label = line.strip().split(",")
            unique_labels.add(int(label))

    sorted_labels = sorted(unique_labels)
    label_to_new_node = {label: idx for idx, label in enumerate(sorted_labels)}

    # Prepare output folder
    out_folder = folder_path / "s_clique"
    os.makedirs(out_folder, exist_ok=True)

    # Write g.node_labels with node == label
    with open(out_folder / "s.node_labels", "w") as f:
        for label, new_id in label_to_new_node.items():
            f.write(f"{new_id},{label}\n")

    # Write all edges to form a clique
    n = len(sorted_labels)
    with open(out_folder / "s.edges", "w") as f:
        for u, v in itertools.combinations(range(n), 2):
            f.write(f"{u},{v}\n")

    print(f"Saved clique with {len(sorted_labels)} nodes to {out_folder}")

from sphera import find_rainbow_clique, process_graph

def find_clique(folder_path, empty1, empty2):
    label_file = folder_path / "g_for_clique" / "g.node_labels"
    edges_file = folder_path / "g_for_clique" / "g.edges"
    start_build = time.time()
    # Load the real graph using the specified edges file
    graph = process_graph.create_real_graph(edges_file)
    graph, node_to_label = process_graph.labeled_graph(graph, label_file)
    # Create the label-to-node dictionary based on node-to-label
    label_to_node = process_graph.create_label_dict(node_to_label)
    end_time = time.time()
    build_time = end_time - start_build
    log_file_path = folder_path / "python_times_sphera_clq.txt"
    with open(log_file_path, "a") as f:
        f.write(f"{build_time:.4f}\n")
    signal.signal(signal.SIGALRM, timeout_handler)
    signal.alarm(timeout)
    try:
        # Call the sphera function to find the maximum rainbow clique with the specified parameters
        max_rainbow_clique, _ = find_rainbow_clique.rc_detection(graph, node_to_label, label_to_node)
        signal.alarm(0)
        print(len(max_rainbow_clique), "cc")
        if len(max_rainbow_clique) == colors:
            return True
    except TimeoutException:
        print(f"sphera: Timeout after {timeout} seconds.")
    finally:
        signal.alarm(0)  # Cancel alarm
    return False

if __name__ == "__main__":
    print("abc")
    if len(sys.argv) != 2:
        print("Usage: python run.py <folder_path>")
        sys.exit(1)

    folder_path = Path(sys.argv[1])
    timeout = 300
    colors = 20
    # find_clique(folder_path)

    # graph_nodes = folder_path / "g" / "g.node_labels"
    # graph_edges = folder_path / "g" / "g.edges"
    # many S
    for num_plant in range(40):
        print(num_plant)
        # subprocess.run(["python", "for_s.py", str(folder_path)], check=True)
        for_s.plant_rainbow_clique(folder_path)
        # if num_plant == 0:
        #     create_clique_from_labels(folder_path / "g" / "g.node_labels")
        print("planted")
        # print("S_planted")
        # sub_nodes = folder_path / "s" / "s.node_labels"
        # sub_edges = folder_path / "s" / "s.edges"
        sub_nodes = folder_path / "s_clique" / "s.node_labels"
        sub_edges = folder_path / "s_clique" / "s.edges"
        graph_nodes = folder_path / "g_for_clique" / "g.node_labels"
        graph_edges = folder_path / "g_for_clique" / "g.edges"
        main((graph_nodes, graph_edges), (sub_nodes, sub_edges), folder_path)
        result = run_with_measurement(find_clique, "SPHERA", folder_path, -1)
        save_result_by_algorithm(result, folder_path)
