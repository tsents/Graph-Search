import networkx as nx
import random
import matplotlib.pyplot as plt
from networkx.algorithms import isomorphism
import time
import json


# Debugging:
def print_graph(graph, colormap, g):
    # Draw the graph with colored nodes
    pos = nx.spring_layout(g, seed=42)  # Position nodes using spring layout
    # pos = {i:[i,pow(-1,i)] for i in range(100)} displays sub graphs nicely
    nx.draw(graph, pos, node_color=list(colormap.values()), cmap=plt.cm.get_cmap("viridis"), with_labels=True)

    nodes_data_list = list(graph.nodes(data=True))
    print(nodes_data_list)

    # Show the plot
    plt.show()


def format_as_array(graph_dict):
    output = [i for i in range((len(graph_dict)))]
    for key, val in graph_dict.items():
        output[val] = int(key)
    return output


def json_to_networkx(graph_file):
    with open(graph_file) as f:
        js_graph = json.load(f)
    # Create a graph
    networkx_graph = nx.Graph()

    # Add nodes with attributes
    for node in js_graph["nodes"]:
        networkx_graph.add_node(node["id"], color=node["color"])

    # Add edges
    for edge in js_graph["links"]:
        networkx_graph.add_edge(edge["source"], edge["target"])
    return networkx_graph


def calculate_prior(prior_data_graph):
    # Precompute the degree of each node
    degrees = dict(prior_data_graph.degree())

    # Use a dictionary comprehension to calculate the sum of degrees of neighbors for each node
    degree_sum_dict = {
        node: sum(degrees[neighbor] for neighbor in prior_data_graph.neighbors(node))
        for node in prior_data_graph.nodes()
    }

    return degree_sum_dict


def search_all_subgraphs(graph, subgraph, output_file):
    all_subgraphs = []
    prior = calculate_prior(graph)
    first_node = find_next_node(subgraph, [], {}, prior)
    for node in graph.nodes():
        if graph.nodes[node]["color"] == subgraph.nodes[first_node]["color"]:
            output = recursion_search(graph, subgraph, node, first_node, {},
                                      {}, output_file, prior)
            all_subgraphs.extend(output)
    return all_subgraphs


def find_next_node(subgraph, added_nodes, restrictions, prior=None):
    if prior is None:
        if len(added_nodes) == 0:
            nodes = list(subgraph.nodes())
            # Randomly choose a node
            return random.choice(nodes)
        neighbors = [node for node in restrictions if node not in added_nodes]
        return list(neighbors)[0]
    neighbors_s = restrictions.keys()
    if len(neighbors_s) == 0:
        return max(subgraph.nodes, key=subgraph.degree)
    max_score = 0
    next_node = 0
    for node_s, options_in_g in restrictions.items():
        if node_s not in added_nodes:
            if len(options_in_g) <= 1:
                return node_s
            score = sum(prior[node_g] for node_g in options_in_g)
            if score > max_score:
                max_score = score
                next_node = node_s
    return next_node


# Graph and Subgraph don't change, node_g is the current node in G,node_s is the corresponding position in S
# restriction - dictionary of lists, path - set of chosen nodes
def recursion_search(graph, subgraph, node_g, node_s, restrictions, path, output_file, prior):
    if node_g in path:
        return []
    if len(path) >= len(subgraph.nodes()) - 1:
        copy = path.copy()
        copy[node_g] = node_s
        with open(output_file, "a") as file:
            # Convert dictionary to JSON format and write it as a new line in the file
            file.write(json.dumps(copy) + "\n")
        return []
    output = []
    inverse_restrictions, empty_set = restriction_update(graph, subgraph, node_g, node_s, restrictions)
    path[node_g] = node_s

    added_nodes = path.values()
    node_next = find_next_node(subgraph, added_nodes, restrictions, prior)

    if not empty_set:
        for u in restrictions[node_next]:
            output.extend(recursion_search(graph, subgraph, u, node_next, restrictions, path, output_file, prior))

    # do the union of the changes, i.e. reverse the new imposed restrictions
    for u in inverse_restrictions:
        if len(inverse_restrictions[u]) > 0 and inverse_restrictions[u][0] == "G":
            restrictions.pop(u)
        else:
            restrictions[u].extend(inverse_restrictions[u])

    path.pop(node_g)
    return output


def restriction_update(graph, subgraph, node_g, node_s, restrictions):
    empty = False
    inverse_restrictions = {}
    for u in subgraph.neighbors(node_s):
        # if u > node_s: #Optional! makes it quicker to retreat in the DFS
        if u in restrictions:  # if there are restrictions on u
            marginal_rest, inverse_marginal_rest = [], []  # A_i in the paper
            for u_instance in restrictions[u]:
                (inverse_marginal_rest, marginal_rest)[graph.has_edge(node_g, u_instance)].append(u_instance)
            if len(marginal_rest) == 0:
                empty = True
            restrictions[u] = marginal_rest
            inverse_restrictions[u] = inverse_marginal_rest
        else:
            restrictions[u] = base_restrictions(node_g, graph, subgraph.nodes[u]["color"], subgraph.degree[u])
            inverse_restrictions[u] = ["G"]
    return inverse_restrictions, empty


def base_restrictions(node, graph, color, degree):
    optional_colored_neighbors = [n for n in graph.neighbors(node) if graph.nodes[n]["color"] == color and
                                  graph.degree[n] >= degree]
    return optional_colored_neighbors


def colors_match(n1_attrib, n2_attrib):
    # returns False if either does not have a color or if the colors do not match'''
    try:
        return n1_attrib['color'] == n2_attrib['color']
    except KeyError:
        return False


def check_matching(graph, subgraph):

    start1 = time.time()
    search_all_subgraphs(graph, subgraph, "output_d.txt")
    time1 = time.time() - start1

    with open("output_d.txt", 'r') as f:
        for line in f:
            # Convert each line (JSON string) back to a dictionary
            g = json.loads(line.strip())
            for key, val in g.items():
                if not graph.nodes[int(key)]["color"] == subgraph.nodes[val]["color"]:
                    print("fucking up color", graph.nodes[key]["color"], subgraph.nodes[val]["color"])
            temp = format_as_array(g)
            for u, v in subgraph.edges():
                if not graph.has_edge(temp[u], temp[v]):
                    print("fucking up edge")

    start2 = time.time()
    graph_matcher = isomorphism.GraphMatcher(graph, subgraph, node_match=colors_match)
    output2 = graph_matcher.subgraph_monomorphisms_iter()
    output2 = list(output2)
    time2 = time.time() - start2

    with open("output_d.txt", 'r') as file:
        output_len = sum(1 for line in file)
    if output_len != len(output2):
        print("wrong answer :(")
    print(output_len, len(output2), time1, time2)
    return time1, time2


if __name__ == "__main__":

    data_graph = json_to_networkx('graphs/graph0.json')
    sub_g = json_to_networkx('graphs/graph2.json')
    check_matching(data_graph, sub_g)
