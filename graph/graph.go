package graph

import (
	_struct "github.com/golang/protobuf/ptypes/struct"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	KnowledgeRelationship = "kr"
)

// NewGraph Create graph.
func NewGraph() *KnowledgeGraph {
	kg := new(KnowledgeGraph)
	kg.Nodes = make(map[string]*_struct.Struct)
	kg.Edges = make(map[string]*_struct.Struct)
	return kg
}

// AddNode Add node id to graph, return true if added (ID's are unique).
func (g *KnowledgeGraph) AddNode(id string) bool {
	if _, ok := g.Nodes[id]; ok {
		return false
	}
	nodeInfo := map[string]interface{}{}
	// Here node info is added as string but this can be struct as well
	nodeInfo["InfoID"+id] = "Info" + id

	g.Nodes[id], _ = structpb.NewStruct(nodeInfo)
	return true
}

// AddEdge Add an edge from u to v.
func (g *KnowledgeGraph) AddEdge(from, to, label string) {
	if _, ok := g.Nodes[from]; !ok {
		g.AddNode(from)
	}
	if _, ok := g.Nodes[to]; !ok {
		g.AddNode(to)
	}

	// Check for presence of from node in edges
	var edgesFromVal map[string]interface{}
	if eVal, eOk := g.Edges[from]; eOk {
		edgesFromVal = eVal.AsMap()
	} else {
		edgesFromVal = map[string]interface{}{}
	}

	// Check for presence of to node in from edge
	var edgesToVal map[string]interface{}
	if eVal, eOk := edgesFromVal[to]; eOk {
		edgesToVal = eVal.(map[string]interface{})
	} else {
		edgesToVal = map[string]interface{}{}
	}

	// Adding Knowledge relationship
	edgesToVal[KnowledgeRelationship] = label

	edgesFromVal[to] = edgesToVal
	g.Edges[from], _ = structpb.NewStruct(edgesFromVal)
}
