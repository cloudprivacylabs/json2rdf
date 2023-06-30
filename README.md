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



```
<http://linkedin.com/jane-doe> <http://schema.org/address> _:b0 .
<http://linkedin.com/jane-doe> <http://schema.org/birthDate> "1972-11-12"^^<http://schema.org/Date> .
<http://linkedin.com/jane-doe> <http://schema.org/birthPlace> "Boulder, CO" .
<http://linkedin.com/jane-doe> <http://schema.org/colleague> <http://www.example.com/Jane.html> .
<http://linkedin.com/jane-doe> <http://schema.org/colleague> <http://www.example.com/John.html> .
<http://linkedin.com/jane-doe> <http://schema.org/email> "info@example.com" .
<http://linkedin.com/jane-doe> <http://schema.org/gender> "female" .
<http://linkedin.com/jane-doe> <http://schema.org/height> "71" .
<http://linkedin.com/jane-doe> <http://schema.org/name> "Jane Doe" .
<http://linkedin.com/jane-doe> <http://schema.org/sameAs> <http://twitter.com/> .
<http://linkedin.com/jane-doe> <http://schema.org/sameAs> <https://www.facebook.com/> .
<http://linkedin.com/jane-doe> <http://schema.org/sameAs> <https://www.linkedin.com/> .
<http://linkedin.com/jane-doe> <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schema.org/Person> .
_:b0 <http://schema.org/addressLocality> "Denver" .
_:b0 <http://schema.org/addressRegion> "CO" .
_:b0 <http://schema.org/postalCode> "80123" .
_:b0 <http://schema.org/streetAddress> "100 Main Street" .
_:b0 <http://www.w3.org/1999/02/22-rdf-syntax-ns#type> <http://schema.org/PostalAddress> .
