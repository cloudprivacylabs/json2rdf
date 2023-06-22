package main

import (
	"encoding/json"
	"os"

	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/piprate/json-gold/ld"
)

func main() {
	dec := json.NewDecoder(os.Stdin)
	g := ls.NewDocumentGraph()
	m := ls.JSONMarshaler{}
	if err := m.Decode(g, dec); err != nil {
		panic(err)
	}
	quads, err := Graph2RDF(g)
	if err != nil {
		panic(err)
	}
	serializer := ld.NQuadRDFSerializer{}
	dataset := ld.NewRDFDataset()
	dataset.Graphs["@default"] = quads

	serializer.SerializeTo(os.Stdout, dataset)
}
