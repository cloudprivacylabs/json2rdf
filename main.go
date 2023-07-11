package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/cloudprivacylabs/lpg"
	"github.com/cloudprivacylabs/lsa/layers/cmd"
	jsoningest "github.com/cloudprivacylabs/lsa/pkg/json"
	"github.com/cloudprivacylabs/lsa/pkg/ls"
	"github.com/piprate/json-gold/ld"
)

func printHelp() {
	fmt.Println(`Translate a JSON file to RDF using a layered JSON schema.

Usage:

  json2rdf -  (read graph from stdin)
  json2rdf graphFile
  json2rdf --bundle bundleFile --type typeName JSONFile
`)
}

// No args: read graph from stdin
// 1 args: read graph from file
// 1 arg with --bundle  --type flags: ingest and read file
func main() {
	bundle := flag.String("bundle", "", "Schema bundle")
	typ := flag.String("type", "", "Type name")
	flag.Parse()
	var dec *json.Decoder
	var g *lpg.Graph
	if len(os.Args) == 2 && os.Args[1] == "-" {
		dec = json.NewDecoder(os.Stdin)
	} else if len(os.Args) == 2 {
		if *bundle == "" && *typ == "" {
			f, err := os.Open(os.Args[1])
			if err != nil {
				panic(err)
			}
			dec = json.NewDecoder(f)
			defer f.Close()
		} else {
			printHelp()
			return
		}
	} else if *bundle != "" && *typ != "" {
		lsctx := ls.NewContext(context.Background())
		g = ls.NewDocumentGraph()
		layer, err := cmd.LoadSchemaFromFileOrRepo(lsctx, "", "", "", *typ, []string{*bundle})
		if err != nil {
			panic(err)
		}
		parser := jsoningest.Parser{
			OnlySchemaAttributes: true,
			IngestNullValues:     true,
			Layer:                layer,
		}
		ingester := &ls.Ingester{Schema: layer}
		builder := ls.NewGraphBuilder(g, ls.GraphBuilderOptions{
			EmbedSchemaNodes:     true,
			OnlySchemaAttributes: true,
		})
		f, err := os.Open(flag.Args()[0])
		if err != nil {
			panic(err)
		}
		defer f.Close()

		_, err = jsoningest.IngestStream(lsctx, "", f, parser, builder, ingester)
		if err != nil {
			panic(err)
		}
	} else {
		printHelp()
		return
	}
	if g == nil {
		g = ls.NewDocumentGraph()
		m := ls.JSONMarshaler{}
		if err := m.Decode(g, dec); err != nil {
			panic(err)
		}
	}
	g2r := NewGraph2RDF(g)
	if err := g2r.Convert(); err != nil {
		panic(err)
	}
	quads := g2r.Quads()
	serializer := ld.NQuadRDFSerializer{}
	dataset := ld.NewRDFDataset()
	dataset.Graphs["@default"] = quads

	serializer.SerializeTo(os.Stdout, dataset)
}
