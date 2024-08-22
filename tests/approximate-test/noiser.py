import networkx as nx
import random
import argparse

def load_graph_from_file(edge_file):
    """Loads a graph from an edge list file."""
    G = nx.Graph()
    with open(edge_file, 'r') as f:
        for line in f:
            u, v = map(int, line.split())
            G.add_edge(u, v)
    return G


def load_edges_from_file(edge_file):
    """Loads a list of edges from an edge list file."""
    edges = []
    with open(edge_file, 'r') as f:
        for line in f:
            u, v = map(int, line.split())
            edges.append((u, v))
    return edges


def remove_random_edges(G, S, edges_in_S, K):
    """
    Removes K random edges from the subgraph induced by S in G,
    ensuring that both S and G remain connected.

    Parameters:
    G (networkx.Graph): The main graph.
    S (list): List of vertices in the subgraph S.
    edges_in_S (list): List of edges in the subgraph S.
    K (int): The number of edges to remove.

    Returns:
    networkx.Graph: The modified graph G.
    """
    # Create a subgraph manually using only the edges specified in edges_in_S
    subgraph = nx.Graph()
    subgraph.add_edges_from(edges_in_S)

    # Ensure we don't try to remove more edges than exist in the subgraph
    if K > len(edges_in_S):
        raise ValueError("K cannot be greater than the number of edges in S.")

    edges_to_remove = random.sample(edges_in_S, K)  # Randomly select K edges to remove

    for edge in edges_to_remove:
        G.remove_edge(*edge)  # Remove edge from the main graph
        subgraph.remove_edge(*edge)  # Remove edge from the subgraph

        # Check if both G and subgraph S are still connected
        if not nx.is_connected(G) or not nx.is_connected(subgraph):
            G.add_edge(*edge)  # Add the edge back if removing it disconnects G or S
            subgraph.add_edge(*edge)  # Restore the edge in the subgraph

    return G


def save_graph_to_file(G, output_file):
    """Saves the graph G as an edge list to a file."""
    with open(output_file, 'w') as f:
        for u, v in G.edges():
            f.write(f"{u} {v}\n")
def load_vertex_list(vertex_file):
    """Loads a list of vertices from a vertex file, ignoring labels."""
    vertices = []
    with open(vertex_file, 'r') as f:
        for line in f:
            vertex, _ = map(int, line.split())
            vertices.append(vertex)
    return vertices




def process_graph(edge_file, vertex_file, edges_in_S_file, output_file, K):
    """
    Loads a graph, processes it to remove K random edges from the specified subgraph,
    and saves the modified graph to a file.

    Parameters:
    edge_file (str): Path to the file containing the edges of the graph.
    vertex_file (str): Path to the file containing the vertices in the subgraph.
    edges_in_S_file (str): Path to the file containing the edges in the subgraph.
    output_file (str): Path to the file where the modified graph will be saved.
    K (int): The number of edges to remove from the subgraph.
    """
    # Load the graph from the file
    G = load_graph_from_file(edge_file)
    # Load the vertices and edges
    S = load_vertex_list(vertex_file)
    edges_in_S = load_edges_from_file(edges_in_S_file)
    # Modify the graph
    G_modified = remove_random_edges(G, S, edges_in_S, K)
    # Save the modified graph to a file
    save_graph_to_file(G_modified, output_file)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Process a graph by removing random edges while maintaining connectivity.")
    parser.add_argument("edge_file", type=str, help="Path to the file containing the edges of the graph.")
    parser.add_argument("vertex_file", type=str, help="Path to the file containing the vertices in the subgraph.")
    parser.add_argument("edges_in_S_file", type=str, help="Path to the file containing the edges in the subgraph.")
    parser.add_argument("output_file", type=str, help="Path to the file where the modified graph will be saved.")
    parser.add_argument("K", type=int, help="Number of edges to remove from the subgraph.")

    args = parser.parse_args()
    process_graph(
        edge_file=args.edge_file,
        vertex_file=args.vertex_file,
        edges_in_S_file=args.edges_in_S_file,
        output_file=args.output_file,
        K=args.K
    )