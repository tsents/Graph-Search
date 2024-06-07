
import networkx as nx
import random
from itertools import combinations



def generate_graph(num_of_colors, p=0.5, nodes_per_color=1000):
    if p is None:
        p = 1/(3*nodes_per_color)
        print(p)
    label_to_node = {}
    node_to_label = {}
    """
    label_to_node_: a dictionary of label and the nodes in this label
    node_to_label_: a dictionary of node and it's label
    """
    G = nx.erdos_renyi_graph(nodes_per_color * num_of_colors, p)
    for node in G.nodes():
        color_node = random.randint(0, num_of_colors - 1)
        node_to_label[node] = color_node
        if color_node in label_to_node:
            label_to_node[color_node].append(node)
        else:
            label_to_node[color_node] = [node]
    temp_for_save = (G, node_to_label, label_to_node)
    # nx.write_gpickle(temp_for_save, "erdos_renyi_graph_p=0.5" + "_{}_classes".format(num_of_colors))
    nx.write_gpickle(temp_for_save, "erdos_renyi_graph_" + "p={}".format(p) + "_{}_classes".format(num_of_colors))
    return G, node_to_label, label_to_node

def generate_subgraph(G, fraction=0.1):
    num_nodes = G.number_of_nodes()
    num_sub_nodes = int(num_nodes * fraction)
    sub_nodes = random.sample(list(G.nodes()), num_sub_nodes)
    subgraph = G.subgraph(sub_nodes).copy()

    return subgraph


if __name__ == '__main__':
    # G, node_to_label, label_to_node = generate_graph(5, 0.5, 400)
    # subgraph = generate_subgraph(G, 0.1)
    # sub_node_to_label = {node: node_to_label[node] for node in subgraph.nodes()}
    # sub_label_to_node = {label: [node for node in nodes if node in subgraph.nodes()] for label, nodes in
    #                      label_to_node.items()}
    # temp_for_save_sub = (subgraph, sub_node_to_label, sub_label_to_node)

    # nx.write_gpickle(temp_for_save_sub, "sub_erdos_renyi_graph_" + "p=0.5" + "_{}_classes".format(5))

    small, color_map,color_to_node = generate_graph(5,0.5,50)
    nx.set_node_attributes(small, color_map, "color")
    import json
    with open('small.json', 'w') as f:
        json.dump(nx.node_link_data(small), f)


