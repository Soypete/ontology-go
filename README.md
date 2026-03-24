# ontology-parser

A Go library for parsing RDF/OWL/TTL ontology files and generating RDF triples from relational data.

## Overview

ontology-parser provides tools for working with RDF ontologies, including:

- **Parsing**: Load RDF/OWL/TTL ontology files
- **Storage**: In-memory triple store with named graph support
- **SPARQL**: Query triples using SPARQL
- **TBOX/ABOX Mapping**: Map relational database tables to RDF triples

## TBOX and ABOX Mapping

This library supports the two-layer ontology mapping pattern common in knowledge representation:

### TBOX (Terminological Box)

TBOX defines the schema/ontology - the concepts and relationships. Use the library to:

- Parse ontology files (OWL, TTL, RDF/XML) to extract class hierarchies and properties
- Define mappings from database columns to ontology predicates
- Generate RDF triples representing schema-level knowledge

Example TBOX mapping:
```yaml
mappings:
  - table: product_categories
    triples:
      - subject: "https://example.org/category/{id}"
        predicate: "rdf:type"
        object: "schema:ProductCategory"
      - subject: "https://example.org/category/{id}"
        predicate: "rdfs:label"
        object: "{name}"
```

### ABOX (Assertional Box)

ABOX contains the actual assertions/instances - data about individuals. The library:

- Maps database rows to RDF individuals
- Generates instance triples from relational data
- Supports inference through ontology-based reasoning

Example ABOX mapping:
```yaml
mappings:
  - table: products
    graph: "https://example.org/data/products"
    triples:
      - subject: "https://example.org/product/{sku}"
        predicate: "rdf:type"
        object: "schema:Product"
      - subject: "https://example.org/product/{sku}"
        predicate: "schema:name"
        object: "{name}"
        datatype: "xsd:string"
      - subject: "https://example.org/product/{sku}"
        predicate: "schema:price"
        object: "{price}"
        datatype: "xsd:decimal"
```

## Installation

```bash
go get github.com/soypete/ontology-go
```

## Usage

### Parse Ontology Files

```go
package main

import (
    "os"
    "github.com/soypete/ontology-go/rdf"
)

func main() {
    f, _ := os.Open("ontology.ttl")
    defer f.Close()
    
    parser := rdf.NewXMLParser("https://example.org/graph")
    triples, err := parser.Parse(f)
    // Handle error...
}
```

### Store and Query Triples

```go
store := store.NewMemoryStore()

store.Register("graph1", []types.Triple{
    {Subject: "https://example.org/item1", Predicate: "rdf:type", Object: "schema:Product"},
})

results := store.Match("", "rdf:type", "schema:Product")
```

## Modules

- `rdf` - RDF/OWL/TTL parsing
- `store` - Triple storage with named graphs
- `sparql` - SPARQL query support
- `ttl` - Turtle format parsing
- `fetch` - Remote ontology fetching
- `types` - Core RDF type definitions

## License

MIT