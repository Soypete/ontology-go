// Package reasoner provides inference over SKOS thesaurus hierarchies.
//
// Deprecated: Use the gorr package instead. This package provides SKOS-specific
// transitive closure reasoning and will be removed in a future release.
// The gorr package provides a full OWL 2 EL consequence-based reasoner with
// incremental reasoning, proof generation, and concurrent saturation.
//
// Example migration:
//
//	Old: reasoner.New(doc, hierarchy)
//	New: gorr.NewReasoner(ctx, gorr.WithSource(doc))
package reasoner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/soypete/ontology-go/skosast"
	"github.com/soypete/ontology-go/ttlast"
	"github.com/soypete/ontology-go/types"
)

type Reasoner struct {
	doc        *ttlast.Document
	hierarchy  *skosast.Hierarchy
	factSet    *FactSet
	rules      []Rule
	sourceFile string
}

func New(doc *ttlast.Document, hierarchy *skosast.Hierarchy, sourceFile string) *Reasoner {
	return &Reasoner{
		doc:        doc,
		hierarchy:  hierarchy,
		factSet:    &FactSet{},
		sourceFile: sourceFile,
		rules: []Rule{
			&TransitiveBroaderRule{},
			&TransitiveNarrowerRule{},
			&SymmetricRelatedRule{},
			&TransitiveExactMatchRule{},
			&InconsistencyRule{},
		},
	}
}

func (r *Reasoner) Run(ctx context.Context) error {
	for _, rule := range r.rules {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := rule.Apply(r.doc, r.hierarchy, r.factSet); err != nil {
			return fmt.Errorf("rule %s failed: %w", rule.Name(), err)
		}
	}

	for i := range r.factSet.Facts {
		if r.factSet.Facts[i].Provenance != nil && r.factSet.Facts[i].Provenance.SourceFile == "" {
			r.factSet.Facts[i].Provenance.SourceFile = r.sourceFile
		}
	}

	return nil
}

func (r *Reasoner) Facts() []Fact {
	return r.factSet.Facts
}

func (r *Reasoner) FactSet() *FactSet {
	return r.factSet
}

func (r *Reasoner) Inconsistencies() []Inconsistency {
	return r.factSet.Inconsistencies
}

func (r *Reasoner) AddRule(rule Rule) {
	r.rules = append(r.rules, rule)
}

func (r *Reasoner) ClearRules() {
	r.rules = nil
}

// OWLClassInfo represents an OWL class with its hierarchy information.
type OWLClassInfo struct {
	ID            string
	IRI           string
	ParentClasses []string
	Label         string
}

// Ontology provides OWL class resolution and type hierarchy inference.
type Ontology struct {
	Classes map[string]*OWLClassInfo
	Doc     *ttlast.Document
}

// LoadOntology loads OWL classes from Turtle files in the given directory.
func LoadOntology(dir string) (*Ontology, error) {
	ont := &Ontology{
		Classes: make(map[string]*OWLClassInfo),
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read ontology directory: %w", err)
	}

	parser := ttlast.NewParser()
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ttl" {
			continue
		}

		doc, err := parser.ParseFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", entry.Name(), err)
		}

		if ont.Doc == nil {
			ont.Doc = doc
		}

		if err := ont.extractClasses(doc); err != nil {
			return nil, fmt.Errorf("failed to extract classes from %s: %w", entry.Name(), err)
		}
	}

	return ont, nil
}

func (o *Ontology) extractClasses(doc *ttlast.Document) error {
	for _, stmt := range doc.Statements {
		triple := stmt.Triple

		subj := o.resolveTerm(triple.Subject)
		pred := o.resolveTerm(triple.Predicate)
		obj := o.resolveTerm(triple.Object)

		if pred == types.RDFSSubClassOf {
			classIRI := subj
			parentIRI := obj

			if _, exists := o.Classes[classIRI]; !exists {
				o.Classes[classIRI] = &OWLClassInfo{
					ID:  classIRI,
					IRI: classIRI,
				}
			}
			o.Classes[classIRI].ParentClasses = append(o.Classes[classIRI].ParentClasses, parentIRI)
		}

		if pred == types.RDFType && (obj == types.RDFSClass || obj == types.OWLClass) {
			if _, exists := o.Classes[subj]; !exists {
				o.Classes[subj] = &OWLClassInfo{
					ID:  subj,
					IRI: subj,
				}
			}
		}

		if pred == types.RDFSLabel {
			if class, exists := o.Classes[subj]; exists {
				class.Label = obj
			}
		}
	}
	return nil
}

var defaultPrefixes = map[string]string{
	"rdf":  "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
	"owl":  "http://www.w3.org/2002/07/owl#",
	"xsd":  "http://www.w3.org/2001/XMLSchema#",
}

func (o *Ontology) resolveTerm(term ttlast.Term) string {
	switch t := term.(type) {
	case *ttlast.IRI:
		return t.Value
	case *ttlast.PrefixedName:
		// Check document prefixes first
		if o.Doc != nil {
			for _, p := range o.Doc.Prefixes {
				if p.Prefix == t.Prefix {
					return p.IRI.Value + t.Local
				}
			}
		}
		// Fall back to default prefixes
		if base, ok := defaultPrefixes[t.Prefix]; ok {
			return base + t.Local
		}
		return ""
	case *ttlast.Literal:
		return t.Value
	default:
		return ""
	}
}

// ResolveClass returns the OWL class info for the given IRI.
func (o *Ontology) ResolveClass(iri string) (*OWLClassInfo, bool) {
	cls, ok := o.Classes[iri]
	return cls, ok
}

// GetClassAncestors returns all ancestor classes (transitive closure of subClassOf).
func (o *Ontology) GetClassAncestors(iri string) []OWLClassInfo {
	visited := make(map[string]bool)
	var ancestors []OWLClassInfo

	o.collectAncestors(iri, visited, &ancestors)

	return ancestors
}

func (o *Ontology) collectAncestors(iri string, visited map[string]bool, ancestors *[]OWLClassInfo) {
	if visited[iri] {
		return
	}
	visited[iri] = true

	class, ok := o.Classes[iri]
	if !ok {
		return
	}

	for _, parentIRI := range class.ParentClasses {
		if parentClass, exists := o.Classes[parentIRI]; exists {
			*ancestors = append(*ancestors, *parentClass)
			o.collectAncestors(parentIRI, visited, ancestors)
		}
	}
}
