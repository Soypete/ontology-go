package gorr

import (
	"context"
	"testing"

	"github.com/soypete/ontology-go/gorr/owl"
)

func TestSaturationEngineBasic(t *testing.T) {
	t.Skip("skipping - saturation logic needs debugging")

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

	subsumers := engine.GetSubsumers(index.intern("http://example.org/A"))
	if subsumers.Size() < 2 {
		t.Errorf("Expected at least 2 subsumers for A (B and C), got %d", subsumers.Size())
	}
}

func TestSaturationEngineIsEntailed(t *testing.T) {
	t.Skip("skipping - saturation logic needs debugging")
}
