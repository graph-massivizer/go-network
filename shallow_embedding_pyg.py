"""
This file creates graph embedding from graph, trasformed into format readable by Pytorch Geometric python library.
"""


import torch
from torch_geometric.nn import Node2Vec
import warnings
import time
import torch.nn as nn
import torch.nn.functional as F
from sklearn.metrics import accuracy_score, roc_auc_score
from sklearn.model_selection import train_test_split
from sklearn.cluster import KMeans


class Classifier(torch.nn.Module):
    def __init__(self, input_dim, hidden_dim, output_dim):
        super(Classifier, self).__init__()
        self.fc1 = torch.nn.Linear(input_dim, hidden_dim)
        self.fc2 = torch.nn.Linear(hidden_dim, output_dim)

    def forward(self, x):
        x = F.relu(self.fc1(x))
        x = self.fc2(x)
        return x



def main():
    device = 'cuda' if torch.cuda.is_available() else 'cpu'
    # Suppress specific warning
    warnings.filterwarnings("ignore", category=FutureWarning)

    data = torch.load('citation_data_tiny_with_combined_leaves.pt')
    dic_arr_data = torch.load('citation_data_tiny_with_combined_leaves_dic_arr.pt')
    dict_data = dic_arr_data["dict"]
    arr_data = dic_arr_data['array']

    #mp.set_start_method('spawn', force=True)
    model = Node2Vec(
        data.edge_index,
        embedding_dim=128,
        walks_per_node=10,
        walk_length=20,
        context_size=10,
        p=1.0,
        q=1.0,
        num_negative_samples=1,
    ).to(device)

    optimizer = torch.optim.Adam(model.parameters(), lr=0.01)
    loader = model.loader(batch_size=128, shuffle=True, num_workers=0)

    def train():
        model.train()
        total_loss = 0
        for pos_rw, neg_rw in loader:
            optimizer.zero_grad()
            loss = model.loss(pos_rw.to(device), neg_rw.to(device))
            loss.backward(retain_graph=True)
            optimizer.step()
            total_loss += loss.item()
        return total_loss / len(loader)
    
    embeddings = model(torch.arange(model.num_nodes, device=device))
    print(embeddings)
    print(embeddings.shape)

    #

    graph_embedding = embeddings.mean(dim=0, keepdim=True)  # Mean pooling
    print("Graph-level embedding shape:", graph_embedding.shape)

    

    kmeans = KMeans(n_clusters=2, random_state=42).fit(embeddings.cpu().detach().numpy())
    labels = torch.tensor(kmeans.labels_, dtype=torch.long).to(device)

    # Assign graph-level label (e.g., majority vote of node labels)
    graph_label = torch.tensor([labels.float().mean().round().item()], device=device)  # Example logic
    print(f"Pseudo-label for the graph: {graph_label}")

    classifier = Classifier(input_dim=embeddings.shape[1], hidden_dim=64, output_dim=2).to(device)
    optimizer = torch.optim.Adam(classifier.parameters(), lr=0.01)
    criterion = torch.nn.CrossEntropyLoss()


    
    X_train, X_test, y_train, y_test = train_test_split(
        embeddings, labels, test_size=0.2, random_state=42
    )
    
    def train_classifier():
        classifier.train()
        optimizer.zero_grad()
        out = classifier(X_train)
        loss = criterion(out, y_train)
        loss.backward(retain_graph=True)
        optimizer.step()
        return loss.item()

    # Train the classifier
    for epoch in range(1, 101):  # Train for 100 epochs
        loss = train_classifier()
        if epoch % 10 == 0:
            print(f"Epoch {epoch}, Loss: {loss:.4f}")

    
    def evaluate(X_test, y_test):
        classifier.eval()
        with torch.no_grad():
            out = classifier(X_test)
            y_pred = out.argmax(dim=1)
            acc = accuracy_score(y_test.cpu(), y_pred.cpu())
            auc = roc_auc_score(y_test.cpu(), F.softmax(out, dim=1)[:, 1].cpu())
            print(f"Test Accuracy: {acc:.4f}, AUC: {auc:.4f}")

    #X_test, y_test = prepare_test_data()
    evaluate(X_test, y_test)




if __name__ == "__main__":
    start = time.perf_counter_ns()
    main()
    end = time.perf_counter_ns()
    print(f"{end-start}ns")


# najdi dodatem embedding model, ki bi upošteval tudi vsebino noda
# če ni tega samo dodaj nakoncu node-embeddinga prilimaj še text-embedding
# kot pri nodetovec se sprehaja po grafu, ampak ne gleda id nodov ampak primerja text embeddinge
# glede na text embedding streniramo model da predicta možne key worde
# kot klasifikacijo damo vse možne key-words
# damo posebi node za embedding da vemo ker je ker
# dodamo stolpce z vsemi moznimi key-wordi, in dodamo 1, če je notri in 0 če ni v temu nodu
# gxboost
