import argparse
import json
import os
import random
import re
import sys
import threading
from concurrent.futures import ThreadPoolExecutor, as_completed
from collections import defaultdict
import time

DEFAULT_COLOR = 0xFFFFFFFF


class Graph:
    def __init__(self, directed=False):
        self.vertices = {}
        self.directed = directed

    def add_vertex(self, u, color=DEFAULT_COLOR):
        if u not in self.vertices:
            self.vertices[u] = {
                'color': color,
                'neighborhood': set(),
                'neighborhood_in': set(),
                'neighborhood_out': set()
            }

    def add_edge(self, u, v):
        if u == v:  # Ignore self loops
            return

        # Add vertices if they don't exist
        self.add_vertex(u)
        self.add_vertex(v)

        # Add to neighborhood (both directions for undirected)
        self.vertices[u]['neighborhood'].add(v)
        self.vertices[v]['neighborhood'].add(u)

        if self.directed:
            self.vertices[u]['neighborhood_out'].add(v)
            self.vertices[v]['neighborhood_in'].add(u)

    def get_vertex_color(self, u):
        return self.vertices[u]['color']

    def get_neighborhood(self, u):
        return self.vertices[u]['neighborhood']

    def get_neighborhood_in(self, u):
        return self.vertices[u]['neighborhood_in']

    def get_neighborhood_out(self, u):
        return self.vertices[u]['neighborhood_out']

    def has_vertex(self, u):
        return u in self.vertices


class Context:
    def __init__(self, graph, subgraph, prior):
        self.graph = graph
        self.subgraph = subgraph
        self.restrictions = {}
        self.path = {}
        self.chosen = set()
        self.prior = prior


def read_graph(filename, input_fmt, input_parse="%d\t%d", directed=False):
    if input_fmt == "json":
        if not filename.endswith(".json"):
            raise ValueError("Filename must end with .json for JSON format")
        return json_to_graph(filename, directed)
    elif input_fmt == "folder":
        return load_graph_from_folder(filename, input_parse, directed)
    else:
        raise ValueError(f"Unsupported input format: {input_fmt}")


def json_to_graph(graph_file, directed=False):
    with open(graph_file) as f:
        js_graph = json.load(f)

    graph = Graph(directed)

    # Add nodes with attributes
    for node in js_graph["nodes"]:
        graph.add_vertex(node["id"], node["color"])

    # Add edges
    for edge in js_graph["links"]:
        src = edge["source"]
        tgt = edge["target"]

        # Add missing nodes with default color
        if not graph.has_vertex(src):
            graph.add_vertex(src, DEFAULT_COLOR)
        if not graph.has_vertex(tgt):
            graph.add_vertex(tgt, DEFAULT_COLOR)

        graph.add_edge(src, tgt)

    return graph


def parse_line(line, input_parse):
    """Parse a line using the input_parse string"""
    line = line.strip()
    if not line:  # Skip empty lines
        raise ValueError("Empty line")

    # Handle common formats more robustly
    if input_parse == "%d\t%d":
        parts = line.split('\t')
        if len(parts) != 2:
            parts = line.split()  # Fallback to whitespace split
        if len(parts) != 2:
            raise ValueError(f"Expected 2 parts, got {len(parts)}")
        return parts[0], parts[1]
    elif input_parse == "%d %d":
        parts = line.split()
        if len(parts) != 2:
            raise ValueError(f"Expected 2 parts, got {len(parts)}")
        return parts[0], parts[1]
    else:
        # Use the original regex approach for other formats
        regex_pattern = (
            re.escape(input_parse)
            .replace(r"%d", r"(\S+)")
            .replace(r"\ ", r"[ ,]+")
        )
        match = re.match(regex_pattern, line)
        if match:
            return match.groups()
        else:
            raise ValueError(f"Line '{line}' does not match format '{input_parse}'")


def load_graph_from_folder(folder_path, input_parse="%d\t%d", directed=False):
    graph = Graph(directed)

    # Find files ending with .node_labels and .edges
    node_labels_file = None
    edges_file = None

    for filename in os.listdir(folder_path):
        if filename.endswith('.node_labels'):
            node_labels_file = os.path.join(folder_path, filename)
        elif filename.endswith('.edges'):
            edges_file = os.path.join(folder_path, filename)

    # Add labeled nodes
    if node_labels_file and os.path.exists(node_labels_file):
        with open(node_labels_file) as f:
            for line in f:
                try:
                    node_str, label_str = parse_line(line, input_parse)
                    node = int(node_str)
                    label = int(label_str)
                    graph.add_vertex(node, label)
                except ValueError as e:
                    print(f"Warning in labels file: {e}")

    # Add edges
    if edges_file and os.path.exists(edges_file):
        with open(edges_file) as f:
            for line in f:
                try:
                    u_str, v_str = parse_line(line, input_parse)
                    u, v = int(u_str), int(v_str)
                    if not graph.has_vertex(u):
                        graph.add_vertex(u, DEFAULT_COLOR)
                    if not graph.has_vertex(v):
                        graph.add_vertex(v, DEFAULT_COLOR)
                    graph.add_edge(u, v)
                except ValueError as e:
                    print(f"Warning in edges file: {e}")

    return graph


def calculate_prior(S, G, prior_policy):
    prior = {}

    if prior_policy == 0:  # d^2 in S
        for v in S.vertices:
            prior[v] = 0
            for u in S.get_neighborhood(v):
                prior[v] += len(S.get_neighborhood(u))
    elif prior_policy == 1:  # d^2 in G
        for v in G.vertices:
            prior[v] = 0
            for u in G.get_neighborhood(v):
                prior[v] += len(G.get_neighborhood(u))
    elif prior_policy == 4:  # d in S
        for v in S.vertices:
            prior[v] = len(S.get_neighborhood(v))

    return prior


def restriction_score(restrictions, prior, u, prior_policy):
    if prior_policy == 0:
        return prior.get(u, 0)
    elif prior_policy == 1:
        score = 0
        if u in restrictions:
            for u_instance in restrictions[u]:
                score += prior.get(u_instance, 0)
        return -score
    elif prior_policy == 2:
        return -len(restrictions.get(u, []))
    elif prior_policy == 3:
        return random.random()
    elif prior_policy in [4, 5]:
        return prior.get(u, 0)
    return 0


def choose_next(restrictions, chosen, subgraph, prior, prior_policy):
    max_score = float('-inf')
    idx = None

    for u in restrictions:
        if u not in chosen:
            if len(restrictions[u]) <= 1:
                return u
            score = restriction_score(restrictions, prior, u, prior_policy)
            if score > max_score:
                max_score = score
                idx = u

    if idx is None:
        for v in subgraph.vertices:
            if v not in chosen:
                return v

    return idx


def choose_start(subgraph, prior, prior_policy):
    if not subgraph.vertices:
        return None
    if prior_policy in [1, 2]:
        return next(iter(subgraph.vertices))

    max_score = float('-inf')
    idx = None

    for u in subgraph.vertices:
        score = restriction_score(None, prior, u, prior_policy)
        if score > max_score:
            max_score = score
            idx = u

    return idx


def colored_neighborhood(graph, u, color, deg, induced, directed_mode=None):
    """Get colored neighborhood with degree constraints"""
    output = set()

    if directed_mode == "out":
        neighbors = graph.get_neighborhood_out(u)
    elif directed_mode == "in":
        neighbors = graph.get_neighborhood_in(u)
    else:
        neighbors = graph.get_neighborhood(u)

    for v in neighbors:
        if graph.get_vertex_color(v) == color:
            if directed_mode == "out":
                v_deg = len(graph.get_neighborhood_out(v))
            elif directed_mode == "in":
                v_deg = len(graph.get_neighborhood_in(v))
            else:
                v_deg = len(graph.get_neighborhood(v))

            if (not induced and deg <= v_deg) or deg == v_deg:
                output.add(v)

    return output


def single_update(context, u, v_s, v_g, directed, induced):
    """Update restrictions for a single vertex"""
    single_inverse = set()

    if u not in context.restrictions:
        # First time - initialize restrictions
        single_inverse.add(None)  # Special marker for new restriction

        if not directed:
            single_rest = colored_neighborhood(
                context.graph, v_g,
                context.subgraph.get_vertex_color(u),
                len(context.subgraph.get_neighborhood(u)),
                induced
            )
        else:
            # Handle directed case
            single_rest = set()

            # Check if v_s has outgoing edge to u
            if u in context.subgraph.get_neighborhood_out(v_s):
                single_rest = colored_neighborhood(
                    context.graph, v_g,
                    context.subgraph.get_vertex_color(u),
                    len(context.subgraph.get_neighborhood_out(u)),
                    induced, "out"
                )

            # Check if v_s has incoming edge from u
            if u in context.subgraph.get_neighborhood_in(v_s):
                if u not in context.subgraph.get_neighborhood_out(v_s):
                    # Only incoming edge
                    single_rest = colored_neighborhood(
                        context.graph, v_g,
                        context.subgraph.get_vertex_color(u),
                        len(context.subgraph.get_neighborhood_in(u)),
                        induced, "in"
                    )
                else:
                    # Both directions - compute intersection
                    in_neighbors = colored_neighborhood(
                        context.graph, v_g,
                        context.subgraph.get_vertex_color(u),
                        len(context.subgraph.get_neighborhood_in(u)),
                        induced, "in"
                    )
                    single_rest = single_rest.intersection(in_neighbors)

        context.restrictions[u] = single_rest
    else:
        # Update existing restrictions
        single_rest = context.restrictions[u].copy()

        if not directed:
            # Remove vertices that don't have edge to v_g
            to_remove = []
            for u_instance in single_rest:
                if u_instance not in context.graph.get_neighborhood(v_g):
                    to_remove.append(u_instance)
                    single_inverse.add(u_instance)

            for item in to_remove:
                single_rest.remove(item)
        else:
            # Handle directed updates
            if u in context.subgraph.get_neighborhood_out(v_s):
                to_remove = []
                for u_instance in single_rest:
                    if u_instance not in context.graph.get_neighborhood_out(v_g):
                        to_remove.append(u_instance)
                        single_inverse.add(u_instance)
                for item in to_remove:
                    single_rest.remove(item)

            if u in context.subgraph.get_neighborhood_in(v_s):
                to_remove = []
                for u_instance in single_rest:
                    if u_instance not in context.graph.get_neighborhood_in(v_g):
                        to_remove.append(u_instance)
                        single_inverse.add(u_instance)
                for item in to_remove:
                    single_rest.remove(item)

        context.restrictions[u] = single_rest

    return single_inverse, len(single_rest) == 0


def update_restrictions(context, v_g, v_s, directed, induced):
    """Update restrictions for all neighbors"""
    empty = False
    inverse_restrictions = {}

    # Use threading for parallel updates like in Go
    def update_neighbor(u):
        if u in context.chosen:
            return u, None, False

        single_inverse, is_empty = single_update(context, u, v_s, v_g, directed, induced)
        return u, single_inverse, is_empty

    # Process neighbors in parallel
    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = []
        for u in context.subgraph.get_neighborhood(v_s):
            futures.append(executor.submit(update_neighbor, u))

        for future in as_completed(futures):
            u, single_inverse, is_empty = future.result()
            if single_inverse is not None:
                inverse_restrictions[u] = single_inverse
                if is_empty:
                    empty = True

    return inverse_restrictions, empty


def recursion_search(context, v_g, v_s, output_file, directed, induced, prior_policy):
    """Main recursive search function"""
    if v_g in context.path:
        return 0

    if len(context.subgraph.vertices) == len(context.chosen) + 1:
        context.path[v_g] = v_s
        with threading.Lock():
            output_file.write(f"{context.path}\n")
            output_file.flush()
        del context.path[v_g]
        return 1

    ret = 0
    context.path[v_g] = v_s
    context.chosen.add(v_s)

    # Save current state of restrictions for this vertex
    self_list = context.restrictions.get(v_s, set())
    if v_s in context.restrictions:
        del context.restrictions[v_s]

    # Update restrictions
    inverse_restrictions, empty = update_restrictions(context, v_g, v_s, directed, induced)
    inverse_restrictions[v_s] = self_list

    if not empty:
        new_v_s = choose_next(context.restrictions, context.chosen, context.subgraph, context.prior, prior_policy)

        # Debug output
        print(
            f"depth {len(context.chosen)}, target size {len(context.restrictions.get(new_v_s, []))}, open {len(context.restrictions)}")

        # Recursive calls
        if new_v_s in context.restrictions:
            for u_instance in context.restrictions[new_v_s].copy():
                if ret == 20:
                    break
                ret += recursion_search(context, u_instance, new_v_s, output_file, directed, induced, prior_policy)
                if ret == 20:
                    break

    # Restore restrictions
    for u in inverse_restrictions:
        if None in inverse_restrictions[u]:  # Special marker for new restriction
            if u in context.restrictions:
                del context.restrictions[u]
        else:
            if u not in context.restrictions:
                context.restrictions[u] = set()
            context.restrictions[u].update(inverse_restrictions[u])

    # Cleanup
    context.chosen.remove(v_s)
    del context.path[v_g]

    return ret


def find_all(graph, subgraph, prior, output_file, directed, induced, prior_policy):
    """Find all subgraph matches"""
    if not subgraph.vertices:
        print("Empty subgraph - no matches to find")
        return 0

    total_matches = 0
    match_lock = threading.Lock()

    def search_from_vertex(u):
        nonlocal total_matches
        context = Context(graph, subgraph, prior)
        matches = recursion_search(context, u, v_0, output_file, directed, induced, prior_policy)
        with match_lock:
            total_matches += matches

    v_0 = choose_start(subgraph, prior, prior_policy)
    if v_0 is None:
        return 0
    # Use threading for parallel execution
    with ThreadPoolExecutor(max_workers=8) as executor:
        futures = []
        batch_size = 0

        for u in graph.vertices:
            if graph.get_vertex_color(u) == subgraph.get_vertex_color(v_0):
                futures.append(executor.submit(search_from_vertex, u))
                batch_size += 1

                # Process in batches like Go code (every 512 vertices)
                if batch_size % 512 == 0:
                    # Wait for current batch to complete
                    for future in as_completed(futures):
                        future.result()
                    futures.clear()

        # Wait for remaining futures
        for future in as_completed(futures):
            future.result()

    return total_matches


def main():
    parser = argparse.ArgumentParser(description="Subgraph isomorphism finder")
    parser.add_argument("--out", default="dat/output.txt", help="Output location")
    parser.add_argument("--fmt", default="json", help="File format (json/folder)")
    parser.add_argument("--parse", default="%d\t%d", help="Parse format for folder")
    parser.add_argument("--prior", type=int, default=0, help="Prior policy (0-5)")
    parser.add_argument("--directed", action="store_true", help="Use directed graphs")
    parser.add_argument("--induced", action="store_true", help="Use induced subgraphs")
    parser.add_argument("--recursion", type=int, help="Set recursion limit (default: 1000)")
    parser.add_argument("graph_file", help="Main graph file")
    parser.add_argument("subgraph_file", help="Subgraph file")

    args = parser.parse_args()
    if args.recursion:
        sys.setrecursionlimit(args.recursion)
    print(f"Recursion limit set to: {sys.getrecursionlimit()}")
    print(f"output -> {args.out}")
    print(f"parsing : {args.parse}")
    print(f"prior : {args.prior}")

    # Read graphs
    G = read_graph(args.graph_file, args.fmt, args.parse, args.directed)
    S = read_graph(args.subgraph_file, args.fmt, args.parse, args.directed)

    # Open output file
    with open(args.out, 'w') as output_file:
        if args.prior == 5:
            # Combined method - run both prior policies
            def run_search(policy):
                prior = calculate_prior(S, G, policy)
                return find_all(G, S, prior, output_file, args.directed, args.induced, policy)

            with ThreadPoolExecutor(max_workers=2) as executor:
                future1 = executor.submit(run_search, 0)
                future2 = executor.submit(run_search, 4)

                future1.result()
                future2.result()

            print("done")
        else:
            prior = calculate_prior(S, G, args.prior)
            matches = find_all(G, S, prior, output_file, args.directed, args.induced, args.prior)
            print(f"matches {matches}")


if __name__ == "__main__":
    # import sys
    #
    # sys.setrecursionlimit(8000)
    # print(sys.getrecursionlimit())
    main()