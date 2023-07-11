package main

import (
	"fmt"
	"strings"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/piprate/json-gold/ld"
)

const RDFTypeTerm = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"

// Top-level RDF nodes:
//
//	A top-level RDF node has rdfIRI, and optionally,  rdfType
//
//	     * rdfIRI: "blank" creates a blank node
//	     * rdfIRI: "." uses current node value as IRI
//	     * rdfIRI: "ref:reference" uses node with schemaNodeId:reference as IRI
//	     * rdfIRI: <value> uses value as IRI
//

type Graph2RDF struct {
	input     *lpg.Graph
	nodeMap   map[*lpg.Node]ld.Node
	processed map[*lpg.Node]struct{}

	blankNodeIx int
	quads       []*ld.Quad
}

func (gr *Graph2RDF) Quads() []*ld.Quad { return gr.quads }

func (gr *Graph2RDF) isProcessed(node *lpg.Node) bool {
	_, ok := gr.processed[node]
	return ok
}

func (gr *Graph2RDF) newBlankNode() ld.Node {
	nextId := fmt.Sprintf("_:b%d", gr.blankNodeIx)
	gr.blankNodeIx++
	return ld.NewBlankNode(nextId)
}

func (gr *Graph2RDF) newQuad(subject, predicate, object ld.Node) {
	gr.quads = append(gr.quads, ld.NewQuad(subject, predicate, object, ""))
}

// BuildTopLevelNodes builds the nodes for objects marked as such in
// the input using rdfIRI or rdfType or both. Returns the graph nodes
// that has assigned RDF nodes.
func (gr *Graph2RDF) BuildTopLevelNodes() ([]*lpg.Node, error) {
	ret := make([]*lpg.Node, 0)
	for nodes := gr.input.GetNodes(); nodes.Next(); {
		node := nodes.Node()
		iri, ok := node.GetProperty("rdfIRI")
		if !ok {
			continue
		}
		rdfIRI := ls.AsPropertyValue(iri, true).AsString()
		rdfType := ls.AsPropertyValue(node.GetProperty("rdfType")).AsString()
		iriIsReference := false
		if strings.HasPrefix(rdfIRI, "ref:") {
			iriIsReference = true
			rdfIRI = rdfIRI[4:]
		}
		makeBlankNode := func() {
			bnode := gr.newBlankNode()
			gr.nodeMap[node] = bnode
			gr.processed[node] = struct{}{}
			ret = append(ret, node)
			if rdfType != "" {
				gr.newQuad(bnode, ld.NewIRI(RDFTypeTerm), ld.NewIRI(rdfType))
			}
		}
		makeNode := func(str string) {
			newNode := ld.NewIRI(str)
			gr.nodeMap[node] = newNode
			gr.processed[node] = struct{}{}
			ret = append(ret, node)
			if rdfType != "" {
				gr.newQuad(newNode, ld.NewIRI(RDFTypeTerm), ld.NewIRI(rdfType))
			}
		}
		if !iriIsReference {
			switch rdfIRI {
			case ".":
				v, err := ls.GetNodeValue(node)
				if err != nil {
					return nil, err
				}
				if v == nil {
					makeBlankNode()
					break
				}
				str, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("rdfIRI value not string: %v", v)
				}
				if str == "" {
					makeBlankNode()
				} else {
					makeNode(str)
				}

			case "", "blank":
				makeBlankNode()

			default:
				makeNode(rdfIRI)
			}
		} else {
			var refNode *lpg.Node
			ls.IterateDescendants(node, func(nd *lpg.Node) bool {
				if ls.AsPropertyValue(nd.GetProperty(ls.SchemaNodeIDTerm)).AsString() == rdfIRI {
					refNode = nd
					return false
				}
				return true
			}, ls.FollowEdgesInEntity, false)
			if refNode == nil {
				makeBlankNode()
				continue
			}
			v, err := ls.GetNodeValue(refNode)
			if err != nil {
				return nil, err
			}
			if v == nil {
				makeBlankNode()
			} else {
				str, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("rdfIRI reference not string: %v", v)
				}
				if str == "" {
					makeBlankNode()
				} else {
					makeNode(str)
				}
			}
		}
	}
	return ret, nil
}

// literalNode will return a literal node if the input node is literal, nil otherwise
func literalNode(node *lpg.Node) ld.Node {
	val, ok := ls.GetRawNodeValue(node)
	if !ok {
		return nil
	}
	typ := ls.AsPropertyValue(node.GetProperty("rdfType")).AsString()
	lang := ls.AsPropertyValue(node.GetProperty("rdfLang")).AsString()
	return ld.NewLiteral(val, typ, lang)
}

func (gr *Graph2RDF) extend(node *lpg.Node) (map[*lpg.Node]struct{}, error) {
	ret := make(map[*lpg.Node]struct{})
	ldNode := gr.nodeMap[node]
	for edges := node.GetEdges(lpg.OutgoingEdge); edges.Next(); {
		nextNode := edges.Edge().GetTo()

		predicate := ls.AsPropertyValue(nextNode.GetProperty("rdfPredicate")).AsString()
		var literal ld.Node
		if predicate != "" {
			literal = literalNode(nextNode)
		}
		if literal != nil {
			// Link to a literal node
			gr.newQuad(ldNode, ld.NewIRI(predicate), literal)
			gr.processed[nextNode] = struct{}{}
			continue
		}

		// nextNode is not literal

		// If it has predicate, it is an intermediate node between this node and nextNode.children
		if predicate != "" {
			gr.processed[nextNode] = struct{}{}

			ldNextNode := gr.nodeMap[nextNode]
			if ldNextNode != nil {
				// There is an rdf node for the next node already, and it also has rdfPredicate
				gr.newQuad(ldNode, ld.NewIRI(predicate), ldNextNode)
				continue
			}

			for childEdges := nextNode.GetEdges(lpg.OutgoingEdge); childEdges.Next(); {
				childNode := childEdges.Edge().GetTo()
				childLdNode := gr.nodeMap[childNode]
				if childLdNode != nil {
					gr.newQuad(ldNode, ld.NewIRI(predicate), childLdNode)
					gr.processed[childNode] = struct{}{}
				} else {
					literal := literalNode(childNode)
					if literal != nil {
						gr.newQuad(ldNode, ld.NewIRI(predicate), literal)
						gr.processed[childNode] = struct{}{}
					} else {
						ret[childNode] = struct{}{}
					}
				}
			}
		}
	}
	return ret, nil
}

func (gr *Graph2RDF) Extend(nodes []*lpg.Node) ([]*lpg.Node, error) {
	newNodes := make(map[*lpg.Node]struct{})
	for _, node := range nodes {
		n, err := gr.extend(node)
		if err != nil {
			return nil, err
		}
		for x := range n {
			newNodes[x] = struct{}{}
		}
	}
	ret := make([]*lpg.Node, 0, len(newNodes))
	for n := range newNodes {
		ret = append(ret, n)
	}
	return ret, nil
}

func (gr *Graph2RDF) Convert() error {
	nodes, err := gr.BuildTopLevelNodes()
	if err != nil {
		return err
	}
	for len(nodes) > 0 {
		nodes, err = gr.Extend(nodes)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewGraph2RDF(input *lpg.Graph) *Graph2RDF {
	return &Graph2RDF{
		input:     input,
		nodeMap:   make(map[*lpg.Node]ld.Node),
		processed: make(map[*lpg.Node]struct{}),
	}
}
