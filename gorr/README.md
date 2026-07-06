# gorr - Go RDF Reasoner

A pure Go OWL EL reasoner with incremental saturation-based reasoning.

## Installation

```bash
go get github.com/soypete/ontology-go/gorr
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"

    "github.com/soypete/ontology-go/gorr"
    "github.com/soypete/ontology-go/gorr/owl"
)

func main() {
    // Create an ontology index
    index := gorr.NewOntologyIndex()

    // Add class hierarchy axioms
    index.AddAxiom(&owl.SubClassOf{
        SubClass:   &owl.Class{URI: "http://example.org/Person"},
        SuperClass: &owl.Class{URI: "http://example.org/Thing"},
    })
    index.AddAxiom(&owl.SubClassOf{
        SubClass:   &owl.Class{URI: "http://example.org/Man"},
        SuperClass: &owl.Class{URI: "http://example.org/Person"},
    })

    // Add property hierarchy
    index.AddAxiom(&owl.SubObjectPropertyOf{
        SubProperty:   &owl.ObjectProperty{URI: "http://example.org/hasSon"},
        SuperProperty: &owl.ObjectProperty{URI: "http://example.org/hasChild"},
    })

    // Add domain constraints
    index.AddAxiom(&owl.ObjectPropertyDomain{
        Property: &owl.ObjectProperty{URI: "http://example.org/hasChild"},
        Domain:   &owl.Class{URI: "http://example.org/Person"},
    })

    // Create saturation engine and run reasoning
    engine := gorr.NewSaturationEngine(index)
    if err := engine.Saturation(context.Background()); err != nil {
        panic(err)
    }

    // Query subsumers
    manHandle := index.intern("http://example.org/Man")
    subsumers := engine.GetSubsumers(manHandle)
    fmt.Println("Man is subsumed by:", subsumers.Size(), "classes")

    // Check entailment
    personHandle := index.intern("http://example.org/Person")
    thingHandle := gorr.HandleTop

    if engine.IsEntailed(manHandle, thingHandle, personHandle) {
        fmt.Println("Man ⊑ Person is entailed")
    }
}
```

## Core Concepts

### Handles

The reasoner uses dense integer handles (`Handle`) to efficiently reference ontology entities. Handles are assigned automatically when axioms are added:

- `HandleBottom` (0) - owl:Nothing
- `HandleTop` (1) - owl:Thing
- User handles start at 2

### OntologyIndex

The `OntologyIndex` stores the ontology and provides:
- `AddAxiom(ax owl.Axiom)` - Add a single axiom
- `AddAxioms(axs []owl.Axiom)` - Add multiple axioms
- `intern(uri string) Handle` - Get handle for a URI
- `ClassCount() int` - Number of classes indexed
- `PropertyCount() int` - Number of properties indexed

### SaturationEngine

The `SaturationEngine` performs forward chaining reasoning:

```go
engine := gorr.NewSaturationEngine(index)

// With custom logger
engine := gorr.NewSaturationEngine(index, gorr.WithLogger(logger))

// Run saturation
engine.Saturation(ctx)

// Query results
subsumers := engine.GetSubsumers(handle)
entailed := engine.IsEntailed(subj, pred, obj)
```

## Supported Axioms

### Class Axioms
- `SubClassOf` - Class hierarchy
- `EquivalentClasses` - Class equivalence
- `DisjointClasses` - Class disjointness

### Property Axioms
- `SubObjectPropertyOf` - Property hierarchy
- `ObjectPropertyDomain` - Domain constraints
- `ObjectPropertyRange` - Range constraints

### Not Supported (EL Profile Limitation)
- Universal quantification (`forall`)
- Negation
- Exact cardinality restrictions

## API Reference

### owl.ClassExpression

```go
// Atomic class
&owl.Class{URI: "http://example.org/Person"}

// Object intersection (conjunction)
&owl.ObjectIntersectionOf{
    A: &owl.Class{URI: "http://example.org/Person"},
    B: &owl.Class{URI: "http://example.org/Mortal"},
}

// Object union (disjunction)
&owl.ObjectUnionOf{...}

// Existential restriction
&owl.ObjectSomeValuesFrom{
    Property: "http://example.org/hasChild",
    Filler:   &owl.Class{URI: "http://example.org/Person"},
}

// Self restriction
&owl.ObjectHasSelf{Property: "http://example.org/hasRelative"}

// Enumeration
&owl.ObjectOneOf{Individuals: []owl.URI{"http://example.org/Alice"}}
```

### owl.Axiom

```go
// SubClassOf
&owl.SubClassOf{
    SubClass:   &owl.Class{URI: "http://example.org/Man"},
    SuperClass: &owl.Class{URI: "http://example.org/Person"},
}

// EquivalentClasses
&owl.EquivalentClasses{
    Classes: []owl.ClassExpression{
        &owl.Class{URI: "http://example.org/Human"},
        &owl.Class{URI: "http://example.org/Person"},
    },
}

// DisjointClasses
&owl.DisjointClasses{
    Classes: []owl.ClassExpression{
        &owl.Class{URI: "http://example.org/Man"},
        &owl.Class{URI: "http://example.org/Woman"},
    },
}

// SubObjectPropertyOf
&owl.SubObjectPropertyOf{
    SubProperty:   &owl.ObjectProperty{URI: "http://example.org/hasSon"},
    SuperProperty: &owl.ObjectProperty{URI: "http://example.org/hasChild"},
}

// ObjectPropertyDomain
&owl.ObjectPropertyDomain{
    Property: &owl.ObjectProperty{URI: "http://example.org/hasChild"},
    Domain:   &owl.Class{URI: "http://example.org/Person"},
}

// ObjectPropertyRange
&owl.ObjectPropertyRange{
    Property: &owl.ObjectProperty{URI: "http://example.org/hasChild"},
    Range:    &owl.Class{URI: "http://example.org/Person"},
}
```

### Validation

All axioms support validation:

```go
ax := &owl.SubClassOf{SubClass: &owl.Class{URI: "A"}, SuperClass: nil}
warnings, ok := ax.Validate()
if !ok {
    fmt.Println("Invalid:", warnings)
}
```

## Performance

- **URI Interning**: O(1) lookup via hash map
- **Hybrid HandleSet**: Sorted slice for <16 elements, map for larger sets
- **Incremental Reasoning**: Can be extended to support incremental saturation

## Related Packages

- `gorr/owl` - OWL type definitions and axioms
- `gorr/owl.Parser` - Parse RDF triples to axioms