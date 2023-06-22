[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/json2rdf?status.svg)](https://godoc.org/github.com/cloudprivacylabs/json2rdf)
# json2rdf

This is a simple proof-of-concept that demonstrates how layered
schemas can be used to transform JSON documents into RDF. In a way, it
shows a JSON-schema based replacement for JSON-LD.

JSON-LD uses a `context` to map JSON constructs to RDF. This is done
by mapping individual object properties to concepts in an ontology
(the `context` defines the "semantics" of the data.) Use of `context`
mappings allows translating certain structurally different data to
comparable RDF models. Because of this, JSON-LD and in particular, RDF
appears in many settings where interoperability between multiple
systems is desired.

JSON-LD context does not define or provide tooling to validate the
structural correctness of the underlying data. For that, JSON schemas
are used. A JSON schema defines and documents a correct JSON encoding
of a data object. Given a JSON document and a schema, one can validate
if the document conforms to that schema, and if it does, data can be
processed without additional structual controls. However, a JSON
schema does not define the semantics (i.e. the ontology mappings) of
the data elements.

A layered schema is a JSON schema combined with additional overlays
that define the semantics of the underlying data elements. Because of
this, a JSON schema combined with an overlay can be used in place of a
JSON-LD document and its context. The added benefit of this approach
is that the JSON schema also defines the structure and format of the
underlying data elements.
