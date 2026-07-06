package owl

import (
	"testing"

	"github.com/soypete/ontology-go/types"
)

func TestParserParseTriples(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/A", Predicate: string(RDFType), Object: string(OWLClass)},
		{Subject: "http://example.org/A", Predicate: string(RDFSSubClassOf), Object: "http://example.org/B"},
		{Subject: "http://example.org/C", Predicate: string(RDFSSubClassOf), Object: "http://example.org/D"},
		{Subject: "http://example.org/X", Predicate: string(OWLEquivalentClass), Object: "http://example.org/Y"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	if len(axioms) != 3 {
		t.Errorf("Expected 3 axioms (skipping rdf:type owl:Class), got %d", len(axioms))
	}
}

func TestParserSubClassOf(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/A", Predicate: string(RDFSSubClassOf), Object: "http://example.org/B"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	if len(axioms) != 1 {
		t.Fatalf("Expected 1 axiom, got %d", len(axioms))
	}

	sco, ok := axioms[0].(*SubClassOf)
	if !ok {
		t.Fatalf("Expected SubClassOf, got %T", axioms[0])
	}

	subClass := sco.SubClass.(*Class)
	superClass := sco.SuperClass.(*Class)
	if subClass.URI != "http://example.org/A" {
		t.Errorf("SubClass = %v, want http://example.org/A", subClass.URI)
	}
	if superClass.URI != "http://example.org/B" {
		t.Errorf("SuperClass = %v, want http://example.org/B", superClass.URI)
	}
}

func TestParserEquivalentClasses(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/A", Predicate: string(OWLEquivalentClass), Object: "http://example.org/B"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	eq, ok := axioms[0].(*EquivalentClasses)
	if !ok {
		t.Fatalf("Expected EquivalentClasses, got %T", axioms[0])
	}

	if len(eq.Classes) != 2 {
		t.Fatalf("Expected 2 classes, got %d", len(eq.Classes))
	}
}

func TestParserDisjointClasses(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/A", Predicate: string(OWLDisjointWith), Object: "http://example.org/B"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	dc, ok := axioms[0].(*DisjointClasses)
	if !ok {
		t.Fatalf("Expected DisjointClasses, got %T", axioms[0])
	}

	if len(dc.Classes) != 2 {
		t.Fatalf("Expected 2 classes, got %d", len(dc.Classes))
	}
}

func TestParserSubObjectPropertyOf(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/p", Predicate: string(RDFSSubPropertyOf), Object: "http://example.org/q"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	sop, ok := axioms[0].(*SubObjectPropertyOf)
	if !ok {
		t.Fatalf("Expected SubObjectPropertyOf, got %T", axioms[0])
	}

	subProp := sop.SubProperty.(*ObjectProperty)
	superProp := sop.SuperProperty.(*ObjectProperty)
	if subProp.URI != "http://example.org/p" {
		t.Errorf("SubProperty = %v, want http://example.org/p", subProp.URI)
	}
	if superProp.URI != "http://example.org/q" {
		t.Errorf("SuperProperty = %v, want http://example.org/q", superProp.URI)
	}
}

func TestParserPropertyDomain(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/p", Predicate: string(RDFSDomain), Object: "http://example.org/C"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	pd, ok := axioms[0].(*ObjectPropertyDomain)
	if !ok {
		t.Fatalf("Expected ObjectPropertyDomain, got %T", axioms[0])
	}

	prop := pd.Property.(*ObjectProperty)
	domain := pd.Domain.(*Class)
	if prop.URI != "http://example.org/p" {
		t.Errorf("Property = %v, want http://example.org/p", prop.URI)
	}
	if domain.URI != "http://example.org/C" {
		t.Errorf("Domain = %v, want http://example.org/C", domain.URI)
	}
}

func TestParserPropertyRange(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/p", Predicate: string(RDFSRange), Object: "http://example.org/D"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	pr, ok := axioms[0].(*ObjectPropertyRange)
	if !ok {
		t.Fatalf("Expected ObjectPropertyRange, got %T", axioms[0])
	}

	prop := pr.Property.(*ObjectProperty)
	rng := pr.Range.(*Class)
	if prop.URI != "http://example.org/p" {
		t.Errorf("Property = %v, want http://example.org/p", prop.URI)
	}
	if rng.URI != "http://example.org/D" {
		t.Errorf("Range = %v, want http://example.org/D", rng.URI)
	}
}

func TestParserUnsupportedAxiom(t *testing.T) {
	p := NewParser()

	triples := []types.Triple{
		{Subject: "http://example.org/A", Predicate: "http://example.org/customPred", Object: "http://example.org/B"},
	}

	axioms, err := p.ParseTriples(triples)
	if err != nil {
		t.Fatalf("ParseTriples() error = %v", err)
	}

	ua, ok := axioms[0].(*UnsupportedAxiom)
	if !ok {
		t.Fatalf("Expected UnsupportedAxiom, got %T", axioms[0])
	}

	if ua.OriginalType != "unknown" {
		t.Errorf("OriginalType = %v, want unknown", ua.OriginalType)
	}
}
