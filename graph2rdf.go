package main

import (
	"fmt"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/piprate/json-gold/ld"
)

const RDFTypeTerm = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"

// Node has a type if the node has a nonrecognized label
func getNodeType(node *lpg.Node) string {
	t := ""
	for x := range node.GetLabels().M {
		if x == ls.DocumentNodeTerm {
			continue
		}
		if ls.IsAttributeType(x) {
			continue
		}
		if t != "" {
			return ""
		}
		t = x
	}
	return t
}

func getNodeID(node *lpg.Node) string {
	for edges := node.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		next := edges.Edge().GetTo()
		if id := ls.AsPropertyValue(next.GetProperty("rdfId")).AsString(); id == "true" {
			v, _ := ls.GetRawNodeValue(next)
			return v
		}
	}
	return ""
}

func getFromNode(node *lpg.Node) *lpg.Node {
	nodes := lpg.SourceNodes(node.GetEdges(lpg.IncomingEdge))
	if len(nodes) == 1 {
		return nodes[0]
	}
	return nil
}

func Graph2RDF(input *lpg.Graph) ([]*ld.Quad, error) {
	quads := make([]*ld.Quad, 0)
	nodeMap := make(map[*lpg.Node]ld.Node)
	blank := 0
	nextBlank := func() string {
		ret := fmt.Sprintf("_:b%d", blank)
		blank++
		return ret
	}
	for nodes := input.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		if rdfNode := ls.AsPropertyValue(node.GetProperty("rdfNode")).AsString(); rdfNode != "" {
			id := getNodeID(node)
			var idNode ld.Node
			if id == "" {
				idNode = ld.NewBlankNode(nextBlank())
			} else {
				idNode = ld.NewIRI(id)
			}
			t := getNodeType(node)
			if t != "" {
				quads = append(quads, ld.NewQuad(idNode, ld.NewIRI(RDFTypeTerm), ld.NewIRI(rdfNode), ""))
			}
			nodeMap[node] = idNode
		}
	}
	for nodes := input.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		if pred := ls.AsPropertyValue(node.GetProperty("rdfPredicate")).AsString(); pred != "" {
			if existing, ok := nodeMap[node]; ok {
				src := getFromNode(node)
				if src != nil {
					source := nodeMap[src]
					if source == nil {
						return nil, fmt.Errorf("Cannot find source node: %v", node)
					}
					quads = append(quads, ld.NewQuad(source, ld.NewIRI(pred), existing, ""))
					continue
				}
			}
			val, ok := ls.GetRawNodeValue(node)
			if ok {
				valueType := ls.AsPropertyValue(node.GetProperty("rdfType")).AsString()
				lang := ls.AsPropertyValue(node.GetProperty("rdfLanguage")).AsString()
				from := getFromNode(node)
				if from == nil {
					return nil, fmt.Errorf("Cannot find source node: %v", node)
				}
				id, ok := nodeMap[from]
				if !ok {
					return nil, fmt.Errorf("Source node does not have id: %v", from)
				}
				quads = append(quads, ld.NewQuad(id, ld.NewIRI(pred), ld.NewLiteral(val, valueType, lang), ""))
			}
		}
	}
	return quads, nil
}
