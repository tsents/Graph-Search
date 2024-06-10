import os
import networkx as nx
import json

mapping = {}

def replace_colors(graph):
    for node, data in graph.nodes(data=True):
        if not data['color'] in mapping:
            mapping[data['color']] = len(mapping)
        data['color'] = mapping[data['color']]


for file_name in os.listdir('graphs'):
    g = nx.read_graphml(f'graphs/{file_name}')
    replace_colors(g)
    g = nx.convert_node_labels_to_integers(g)
    with open('json/graph' + file_name[6] + '.json', 'w') as f:
        json.dump(nx.node_link_data(g), f)


with open('json/mapping.json','w') as file:
    json.dump(mapping,file)

print(len(mapping))