package gorr

import (
	"testing"

	"github.com/soypete/ontology-go/gorr/owl"
)

func TestOntologyIndexBasic(t *testing.T) {
	index := NewOntologyIndex()

	if index.ClassCount() != 0 {
		t.Errorf("Expected 0 classes, got %d", index.ClassCount())
	}

	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/A"},
		SuperClass: &owl.Class{URI: "http://example.org/B"},
	})

	if index.ClassCount() != 2 {
		t.Errorf("Expected 2 classes, got %d", index.ClassCount())
	}
}

func TestOntologyIndexAddAxioms(t *testing.T) {
	index := NewOntologyIndex()

	axioms := []owl.Axiom{
		&owl.SubClassOf{
			SubClass:   &owl.Class{URI: "http://example.org/A"},
			SuperClass: &owl.Class{URI: "http://example.org/B"},
		},
		&owl.SubClassOf{
			SubClass:   &owl.Class{URI: "http://example.org/B"},
			SuperClass: &owl.Class{URI: "http://example.org/C"},
		},
	}

	index.AddAxioms(axioms)

	if index.ClassCount() != 3 {
		t.Errorf("Expected 3 classes, got %d", index.ClassCount())
	}
}

func TestOntologyIndexEquivalentClasses(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.EquivalentClasses{
		Classes: []owl.ClassExpression{
			&owl.Class{URI: "http://example.org/A"},
			&owl.Class{URI: "http://example.org/B"},
		},
	})

	if index.ClassCount() != 2 {
		t.Errorf("Expected 2 classes, got %d", index.ClassCount())
	}

	subsumers := index.GetSubsumers(index.intern("http://example.org/A"))
	if subsumers.Size() != 1 {
		t.Errorf("Expected 1 subsumer for A, got %d", subsumers.Size())
	}
}

func TestOntologyIndexDisjointClasses(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.DisjointClasses{
		Classes: []owl.ClassExpression{
			&owl.Class{URI: "http://example.org/A"},
			&owl.Class{URI: "http://example.org/B"},
		},
	})

	axioms := index.Axioms()
	if len(axioms) != 1 {
		t.Errorf("Expected 1 axiom, got %d", len(axioms))
	}
	_, ok := axioms[0].(*owl.DisjointClasses)
	if !ok {
		t.Error("Expected DisjointClasses axiom")
	}
}

func TestOntologyIndexSubObjectPropertyOf(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubObjectPropertyOf{
		SubProperty:   &owl.ObjectProperty{URI: "http://example.org/p"},
		SuperProperty: &owl.ObjectProperty{URI: "http://example.org/q"},
	})

	if index.PropertyCount() != 2 {
		t.Errorf("Expected 2 properties, got %d", index.PropertyCount())
	}
}

func TestOntologyIndexObjectPropertyRange(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.ObjectPropertyRange{
		Property: &owl.ObjectProperty{URI: "http://example.org/p"},
		Range:    &owl.Class{URI: "http://example.org/RangeClass"},
	})

	rangeHandle := index.intern("http://example.org/RangeClass")
	propHandle := index.intern("http://example.org/p")

	rng, ok := index.GetPropertyRange(propHandle)
	if !ok {
		t.Error("Expected property range to exist")
	}
	if rng != rangeHandle {
		t.Errorf("Expected range %v, got %v", rangeHandle, rng)
	}
}

func TestOntologyIndexGetClass(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/A"},
		SuperClass: &owl.Class{URI: "http://example.org/B"},
	})

	handle := index.intern("http://example.org/A")
	expr, ok := index.GetClass(handle)
	if !ok {
		t.Error("Expected class to exist")
	}
	if expr.String() != "http://example.org/A" {
		t.Errorf("Expected http://example.org/A, got %s", expr.String())
	}
}

func TestOntologyIndexURI(t *testing.T) {
	index := NewOntologyIndex()

	handle := index.intern("http://example.org/Test")
	uri := index.URI(handle)
	if uri != "http://example.org/Test" {
		t.Errorf("Expected http://example.org/Test, got %s", uri)
	}
}

func TestOntologyIndexPropertySubsumers(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubObjectPropertyOf{
		SubProperty:   &owl.ObjectProperty{URI: "http://example.org/childProp"},
		SuperProperty: &owl.ObjectProperty{URI: "http://example.org/parentProp"},
	})

	childHandle := index.intern("http://example.org/childProp")
	subsumers := index.GetPropertySubsumers(childHandle)
	if subsumers.Size() != 1 {
		t.Errorf("Expected 1 property subsumer, got %d", subsumers.Size())
	}
}

func TestOntologyIndexInternClassExpressions(t *testing.T) {
	index := NewOntologyIndex()

	exprs := []owl.ClassExpression{
		&owl.Class{URI: "http://example.org/A"},
		&owl.Class{URI: "http://example.org/B"},
		&owl.Class{URI: "http://example.org/C"},
	}

	index.mu.Lock()
	handles := index.internClassExpressionsUnlocked(exprs)
	index.mu.Unlock()

	if len(handles) != 3 {
		t.Errorf("Expected 3 handles, got %d", len(handles))
	}
}

func TestOntologyIndexInternObjectProperty(t *testing.T) {
	index := NewOntologyIndex()

	index.mu.Lock()
	handle := index.internObjectPropertyUnlocked(&owl.ObjectProperty{URI: "http://example.org/p"})
	index.mu.Unlock()

	if handle == 0 {
		t.Error("Expected non-zero handle")
	}

	prop, ok := index.GetObjectProperty(handle)
	if !ok {
		t.Error("Expected object property to exist")
	}
	if prop.URI != "http://example.org/p" {
		t.Errorf("Expected http://example.org/p, got %s", prop.URI)
	}
}
