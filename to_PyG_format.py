"""
This file transforms Graph created by this library to a format that is readable by Pytorch Geometric library.
"""

import json
import torch
from torch_geometric.data import Data
import time

start = time.perf_counter_ns()

# Read the graph data from the JSON file
with open('graph_data.json') as f:
    graph_data = json.load(f)

# Extract nodes and edges from the graph data
nodes = graph_data['Nodes']
edges = graph_data['Edges']

print(len(nodes))

keywords_to_num = {}
num_to_keywords = []
node_to_keywords = {}
count = 0

for key in nodes:
    for word in nodes[key]["Attributes"]["keywords"]:
        if word not in keywords_to_num:
            keywords_to_num[word] = count
            count+=1
            num_to_keywords.append(word)
        if key not in node_to_keywords:
            node_to_keywords[key] = []
        node_to_keywords[key].append(keywords_to_num[word])


# Convert nodes' features into a tensor (assuming features are numeric)
node_features = []
node_id_map = {}  # Map node IDs to indices
for idx, (node_id, node) in enumerate(nodes.items()):
    node_id_map[node_id] = idx
    features = [value for value in node['Attributes'].values() if isinstance(value, (int, float))]  # Flatten to numeric features
    node_features.append(features)

node_features = torch.tensor(node_features, dtype=torch.float)

# Create the edge_index tensor (assuming edges are undirected, i.e., From and To are bidirectional)
edge_index = []
edge_attr = []  # Store edge attributes (e.g., weights)
for edge in edges.values():
    if edge['First_node']['ID'] in node_id_map:
        from_idx = node_id_map[edge['First_node']['ID']]
    else:
        print(f"Key not found: {edge['Second_node']['ID']}")
        continue
    if edge['Second_node']['ID'] in node_id_map:
        to_idx = node_id_map[edge['Second_node']['ID']]
    else:
        print(f"Key not found: {edge['Second_node']['ID']}")
        continue
    
    # Since this is an undirected graph, add both directions
    edge_index.append([from_idx, to_idx])
    edge_index.append([to_idx, from_idx])
    
    # Add edge attributes (e.g., weight)
    weight = edge['Attributes'].get('weight', 1.0)  # Default weight is 1 if not found
    edge_attr.append(weight)
    edge_attr.append(weight)  # Add the reverse edge as well

edge_index = torch.tensor(edge_index, dtype=torch.long).t().contiguous()
edge_attr = torch.tensor(edge_attr, dtype=torch.float).view(-1, 1)

# Create PyTorch Geometric Data object
data = Data(x=node_features, edge_index=edge_index, edge_attr=edge_attr)

# Print the created Data object
print(data)

torch.save(data, 'citation_data_tiny_with_combined_leaves.pt')
#citation_data_tiny_with_combined_leaves_full

torch.save({'dict': node_to_keywords, 'array': num_to_keywords}, 'citation_data_tiny_with_combined_leaves_dic_arr.pt')

end = time.perf_counter_ns()
print(f"{end-start}ns")