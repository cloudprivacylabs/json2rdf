[![GoDoc](https://godoc.org/github.com/cloudprivacylabs/json2rdf?status.svg)](https://godoc.org/github.com/cloudprivacylabs/json2rdf)
# json2rdf - JSON-LD vs. Layered Schemas

This is a proof-of-concept that demonstrates how layered schemas can
be used to transform JSON documents into RDF instead of using JSON-LD.
There are several advantages of using layered schemas instead of
JSON-LD, some of which are that JSON schemas are widely available, and
that they can be used to describe and validate JSON documents.

A layered schema is composed of a schema that defines data structures,
and zero or more overlays that annotate the schema with semantic
information and metadata. These annotations "adjust" the schema by
adding constraints or relaxing existing ones, and "enrich" the schema
by adding processing information such as pointers to normalization
tables, or mappings to a common ontology, which is what we will use
here.

Let's consider the following simple JSON-LD document describing a
`Person`. It uses the popular `https://schema.org` ontology to
describe a person object:

``` javascript
{
    "@context": "https://schema.org",
    "@type": "Person",
    "@id": "http://linkedin.com/jane-doe",
    "address": {
        "@type": "PostalAddress",
        "addressLocality": "Denver",
        "addressRegion": "CO"
    },
    "colleague": [
        "http://example.com/John.html",
        "http://example.com/Amy.html"
    ],
    "email": "jane@example.com",
    "name": "Jane Doe",
    "sameAs" : [ "https://facebook.com/jane-doe",
                 "https://twitter.com/jane-doe"]
}
```

JSON-LD uses the `context` to map JSON keys to RDF. This is done by
mapping individual object properties to concepts in an ontology (the
`context` defines the "semantics" of the data.) The graphical RDF
representation for this object is:

![Person](person-rdf.png)

![Composition](overlays.png)

Now, let's write a JSON schema for this object. The following JSON
schema contains the definitions for two objects, a `Person` and a
`PostalAddress`, and can be used to validate either object:

``` javascript
{
    "oneOf": [
      { "$ref": "#/definitions/Person" },
      { "$ref": "#/definitions/PostalAddress" }
    ],
    "definitions": {
        "PostalAddress": {
            "type": "object",
            "properties": {
                "addressLocality": {
                    "type": "string"
                },
                "addressRegion": {
                    "type": "string"
                }
            }
        },
        "Person": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string",
                    "format": "uri"
                },
                "address": {
                    "$ref": "#/definitions/PostalAddress"
                },
                "colleague": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "email": {
                    "type": "string",
                    "format": "email"
                },
                "name": {
                    "type": "string"
                },
                "sameAs" : {
                     "type": "array",
                    "items": {
                        "type": "string",
                        "format": "uri"
                    }
                }
            }
        }
    }
}
```

Now we annotate this schema by defining mappings to the schema.org
ontology using an overlay. Let's say we will use `rdfNode` to mean the
data field should be translated as an RDF subject or object node. For instance:

``` javascript
...
"definitions": {
    "PostalAddress": {
        "x-ls": {
            "rdfNode": "http://schema.org/PostalAddress"
        },
    }
...
```

The `definitions/PostalAddress` will match the `PostalAddress`
definitions in the original schema.

``` javascript
{
    "definitions": {
        "PostalAddress": {
            "type": "object",
            "x-ls": {
                "rdfNode": "http://schema.org/PostalAddress"
            },
            "properties": {
                "addressLocality": {
                    "x-ls": {
                        "rdfPredicate" : "http://schema.org/addressLocality"
                    }
                },
                "addressRegion": {
                    "x-ls": {
                        "rdfPredicate": "http://schema.org/addressRegion"
                    }
                }
            }
        },
        "Person": {
            "type": "object",
            "x-ls": {
                "rdfNode": "http://schema.org/Person"
            },
            "properties": {
                "id": {
                    "x-ls": {
                        "rdfId": "true"
                    }
                },
                "address": {
                    "x-ls": {
                        "rdfPredicate": "http://schema.org/address"
                    }
                },
                "colleague": {
                    "x-ls": {
                        "rdfPredicate": "http://schema.org/colleague"
                    }
                },
                "email": {
                    "x-ls": {
                        "rdfPredicate": "http://schema.org/email"
                    }
                },
                "name": {
                    "x-ls": {
                        "rdfPredicate": "http://schema.org/name"
                    }
                },
	            "sameAs" : {
                    "x-ls": {
                        "rdfPredicate": "http://schema.org/sameAs"
                    }
                }
            }
        }
    }
}
```




layers ingest json --bundle person.bundle.yaml --type http://schema.org/Person person-sample.json 
