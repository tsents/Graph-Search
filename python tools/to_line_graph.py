import networkx as nx
import matplotlib.pyplot as plt
import itertools

def rank_edges(graph, node_to_label_g):
    # Count all optional edges in G to find rarity.
    edge_list = list(graph.edges)
    edge_ranks = {}
    for edge in edge_list:
        labeled_e = (node_to_label_g[edge[0]], node_to_label_g[edge[1]])
        if labeled_e in edge_ranks:
            edge_ranks[labeled_e] += 1
        elif (labeled_e[1],labeled_e[0]) in edge_ranks:
            edge_ranks[(labeled_e[1],labeled_e[0])] += 1
        else:
            edge_ranks[labeled_e] = 1
    return edge_ranks

def cheapest_hamiltonian(sub_g, node_to_label_s, edge_ranks):
    # Add ranks by rarity in G to edges in sub graph.
    edge_list = list(sub_g.edges)
    for edge in edge_list:
        labeled_e = (node_to_label_s[edge[0]], node_to_label_s[edge[1]])
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



import json

if __name__ == "__main__":
    graph, node_to_label_g, label_to_node_g = nx.read_gpickle("erdos_renyi_graph_p=0.5" + "_{}_classes".format(5))
    sub_g, node_to_label_s , label_to_node_s = nx.read_gpickle("sub_erdos_renyi_graph_p=0.5" + "_{}_classes".format(5))

    #plot_graph(sub_g, node_to_label_s, "Subgraph", "subgraph.png")
    G_edge_ranks = rank_edges(graph, node_to_label_g)
    hamiltonian = cheapest_hamiltonian(sub_g, node_to_label_s, G_edge_ranks)

    # nx.set_node_attributes(graph, node_to_label_g, "color")
    # nx.set_node_attributes(sub_g, node_to_label_s, "color")
    # with open('graph1.json', 'w') as f:
    #     json.dump(nx.node_link_data(graph), f)
    # with open('graph2.json', 'w') as f:
    #     json.dump(nx.node_link_data(sub_g), f)
    with open('ordering.json', 'w') as f:
        json.dump({"ordering" : hamiltonian}, f)
