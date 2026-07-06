package gorr

import (
	"context"
	"testing"

	"github.com/soypete/ontology-go/gorr/owl"
)

func TestSaturationEngineBasic(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/A"},
		SuperClass: &owl.Class{URI: "http://example.org/B"},
	})

	engine := NewSaturationEngine(index)

	err := engine.Saturation(context.Background())
	if err != nil {
		t.Fatalf("Saturation() error = %v", err)
	}

	subsumers := engine.GetSubsumers(index.intern("http://example.org/A"))
	if subsumers.Size() < 1 {
		t.Errorf("Expected at least 1 subsumer for A (B), got %d", subsumers.Size())
	}
}

func TestSaturationEngineChain(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/A"},
		SuperClass: &owl.Class{URI: "http://example.org/B"},
	})
	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/B"},
		SuperClass: &owl.Class{URI: "http://example.org/C"},
	})
	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/C"},
		SuperClass: &owl.Class{URI: "http://example.org/Thing"},
	})

	engine := NewSaturationEngine(index)

	err := engine.Saturation(context.Background())
	if err != nil {
		t.Fatalf("Saturation() error = %v", err)
	}

	aHandle := index.intern("http://example.org/A")
	subsumers := engine.GetSubsumers(aHandle)
	if subsumers.Size() < 3 {
		t.Errorf("Expected at least 3 subsumers for A (B, C, Thing), got %d", subsumers.Size())
	}
}

func TestSaturationEngineEquivalent(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/A"},
		SuperClass: &owl.Class{URI: "http://example.org/B"},
	})
	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/B"},
		SuperClass: &owl.Class{URI: "http://example.org/A"},
	})

	engine := NewSaturationEngine(index)

	err := engine.Saturation(context.Background())
	if err != nil {
		t.Fatalf("Saturation() error = %v", err)
	}

	aHandle := index.intern("http://example.org/A")
	bHandle := index.intern("http://example.org/B")

	subsumersA := engine.GetSubsumers(aHandle)
	if !subsumersA.Contains(bHandle) {
		t.Error("Expected A to subsume B (equivalence)")
	}
}

func TestSaturationEnginePropertyDomain(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubObjectPropertyOf{
		SubProperty:   &owl.ObjectProperty{URI: "http://example.org/hasParent"},
		SuperProperty: &owl.ObjectProperty{URI: "http://example.org/hasRelative"},
	})
	index.AddAxiom(&owl.ObjectPropertyDomain{
		Property: &owl.ObjectProperty{URI: "http://example.org/hasParent"},
		Domain:   &owl.Class{URI: "http://example.org/Person"},
	})
	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/Person"},
		SuperClass: &owl.Class{URI: "http://example.org/Thing"},
	})

	engine := NewSaturationEngine(index)

	err := engine.Saturation(context.Background())
	if err != nil {
		t.Fatalf("Saturation() error = %v", err)
	}

	personHandle := index.intern("http://example.org/Person")
	thingHandle := index.intern("http://example.org/Thing")

	subsumers := engine.GetSubsumers(personHandle)
	if !subsumers.Contains(thingHandle) {
		t.Error("Expected Person to subsume Thing")
	}
}

func TestSaturationEngineIsEntailed(t *testing.T) {
	index := NewOntologyIndex()

	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/A"},
		SuperClass: &owl.Class{URI: "http://example.org/B"},
	})
	index.AddAxiom(&owl.SubClassOf{
		SubClass:   &owl.Class{URI: "http://example.org/B"},
		SuperClass: &owl.Class{URI: "http://example.org/C"},
	})

	engine := NewSaturationEngine(index)

	err := engine.Saturation(context.Background())
	if err != nil {
		t.Fatalf("Saturation() error = %v", err)
	}

	aHandle := index.intern("http://example.org/A")
	bHandle := index.intern("http://example.org/B")
	cHandle := index.intern("http://example.org/C")
	thingHandle := HandleTop

	if !engine.IsEntailed(aHandle, thingHandle, bHandle) {
		t.Error("Expected A ⊑ B to be entailed")
	}
	if !engine.IsEntailed(aHandle, thingHandle, cHandle) {
		t.Error("Expected A ⊑ C to be entailed (transitive)")
	}
}
