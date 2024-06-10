import networkx as nx
import json


def rank_edges(graph):
    # Count all optional edges in G to find rarity.
    edge_list = list(graph.edges)
    edge_ranks = {}
    for edge in edge_list:
        labeled_e = (graph.nodes[edge[0]]['color'], graph.nodes[edge[1]]['color'])
        if labeled_e in edge_ranks:
            edge_ranks[labeled_e] += 1
        elif (labeled_e[1],labeled_e[0]) in edge_ranks:
            edge_ranks[(labeled_e[1],labeled_e[0])] += 1
        else:
            edge_ranks[labeled_e] = 1
    return edge_ranks

def cheapest_hamiltonian(sub_g,edge_ranks):
    # Add ranks by rarity in G to edges in sub graph.
    edge_list = list(sub_g.edges)
    for edge in edge_list:
        labeled_e = (sub_g.nodes[edge[0]]['color'], sub_g.nodes[edge[1]]['color'])
        # If the connection between these 2 labels is in dict:
        if labeled_e in edge_ranks:
            sub_g[edge[0]][edge[1]]['weight'] = edge_ranks[labeled_e]
        elif (labeled_e[1], labeled_e[0]) in edge_ranks:
            rank = edge_ranks[(labeled_e[1], labeled_e[0])]
            sub_g[edge[0]][edge[1]]['weight'] = rank
        # If the connection between these 2 labels is not in G, exact sub isomorphism can not be found.
        else:
            return -1, edge, labeled_e
    # Find the cheapest Hamiltonian.
    tsp = nx.approximation.traveling_salesman_problem
    path = tsp(sub_g, cycle=False)
    seen = set()
    unique_path = [x for x in path if not (x in seen or seen.add(x))]
    return unique_path

def read_json_file(filename):
    with open(filename) as f:
        js_graph = json.load(f)
    g = nx.node_link_graph(js_graph)
    g = nx.convert_node_labels_to_integers(g)
    return g

def full_pipeline(f_name):
    G = read_json_file(f_name)
    G_edge_ranks = rank_edges(G)
    hamiltonian = cheapest_hamiltonian(G, G_edge_ranks)
    with open('ordering_'+f_name, 'w') as f:
        json.dump({"ordering":hamiltonian}, f)
full_pipeline('graph2.json')