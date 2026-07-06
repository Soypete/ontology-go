// Package owl provides OWL 2 EL class expressions and axioms.
package owl

import (
	"fmt"

	"github.com/soypete/ontology-go/types"
)

var (
	OWLClass                 = URI("http://www.w3.org/2002/07/owl#Class")
	OWLObjectProperty        = URI("http://www.w3.org/2002/07/owl#ObjectProperty")
	OWLDataProperty          = URI("http://www.w3.org/2002/07/owl#DatatypeProperty")
	RDFType                  = URI("http://www.w3.org/1999/02/22-rdf-syntax-ns#type")
	RDFSSubClassOf           = URI("http://www.w3.org/2000/01/rdf-schema#subClassOf")
	RDFSSubPropertyOf        = URI("http://www.w3.org/2000/01/rdf-schema#subPropertyOf")
	RDFSDomain               = URI("http://www.w3.org/2000/01/rdf-schema#domain")
	RDFSRange                = URI("http://www.w3.org/2000/01/rdf-schema#range")
	OWLEquivalentClass       = URI("http://www.w3.org/2002/07/owl#equivalentClass")
	OWLDisjointWith          = URI("http://www.w3.org/2002/07/owl#disjointWith")
	OWLPropertyChain         = URI("http://www.w3.org/2002/07/owl#propertyChainAxiom")
	OWLInverseOf             = URI("http://www.w3.org/2002/07/owl#inverseOf")
	OWLTransitiveProperty    = URI("http://www.w3.org/2002/07/owl#TransitiveProperty")
	OWLFunctionalProperty    = URI("http://www.w3.org/2002/07/owl#FunctionalProperty")
	OWLInverseFunctionalProp = URI("http://www.w3.org/2002/07/owl#InverseFunctionalProperty")
	OWLSymmetricProperty     = URI("http://www.w3.org/2002/07/owl#SymmetricProperty")
	OWLAsymmetricProperty    = URI("http://www.w3.org/2002/07/owl#AsymmetricProperty")
	OWLReflexiveProperty     = URI("http://www.w3.org/2002/07/owl#ReflexiveProperty")
	OWLIrreflexiveProperty   = URI("http://www.w3.org/2002/07/owl#IrreflexiveProperty")
	OWLThing                 = URI("http://www.w3.org/2002/07/owl#Thing")
	OWLNothing               = URI("http://www.w3.org/2002/07/owl#Nothing")
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) ParseTriples(triples []types.Triple) ([]Axiom, error) {
	var axioms []Axiom

	for _, t := range triples {
		ax, err := p.tripleToAxiom(t)
		if err != nil {
			return nil, fmt.Errorf("failed to parse triple: %w", err)
		}
		if ax != nil {
			axioms = append(axioms, ax)
		}
	}

	return axioms, nil
}

func (p *Parser) tripleToAxiom(t types.Triple) (Axiom, error) {
	pred := URI(t.Predicate)
	switch {
	case pred == RDFType:
		return p.parseTypeTriple(t)
	case pred == RDFSSubClassOf:
		return p.parseSubClassOf(t)
	case pred == OWLEquivalentClass:
		return p.parseEquivalentClasses(t)
	case pred == OWLDisjointWith:
		return p.parseDisjointClasses(t)
	case pred == RDFSSubPropertyOf:
		return p.parseSubPropertyOf(t)
	case pred == OWLPropertyChain:
		return p.parsePropertyChain(t)
	case pred == RDFSDomain:
		return p.parsePropertyDomain(t)
	case pred == RDFSRange:
		return p.parsePropertyRange(t)
	default:
		return &UnsupportedAxiom{
			OriginalType: "unknown",
			Content:      fmt.Sprintf("%s %s %s", t.Subject, t.Predicate, t.Object),
		}, nil
	}
}

func (p *Parser) parseTypeTriple(t types.Triple) (Axiom, error) {
	obj := URI(t.Object)
	switch {
	case obj == OWLClass:
		return nil, nil
	case obj == OWLObjectProperty:
		return nil, nil
	case obj == OWLDataProperty:
		return nil, nil
	default:
		return &UnsupportedAxiom{
			OriginalType: "rdf:type",
			Content:      fmt.Sprintf("%s a %s", t.Subject, t.Object),
		}, nil
	}
}

func (p *Parser) parseSubClassOf(t types.Triple) (Axiom, error) {
	return &SubClassOf{
		SubClass:   &Class{URI: URI(t.Subject)},
		SuperClass: &Class{URI: URI(t.Object)},
	}, nil
}

func (p *Parser) parseEquivalentClasses(t types.Triple) (Axiom, error) {
	return &EquivalentClasses{
		Classes: []ClassExpression{
			&Class{URI: URI(t.Subject)},
			&Class{URI: URI(t.Object)},
		},
	}, nil
}

func (p *Parser) parseDisjointClasses(t types.Triple) (Axiom, error) {
	return &DisjointClasses{
		Classes: []ClassExpression{
			&Class{URI: URI(t.Subject)},
			&Class{URI: URI(t.Object)},
		},
	}, nil
}

func (p *Parser) parseSubPropertyOf(t types.Triple) (Axiom, error) {
	return &SubObjectPropertyOf{
		SubProperty:   &ObjectProperty{URI: URI(t.Subject)},
		SuperProperty: &ObjectProperty{URI: URI(t.Object)},
	}, nil
}

func (p *Parser) parsePropertyChain(t types.Triple) (Axiom, error) {
	return &UnsupportedAxiom{
		OriginalType: "owl:propertyChainAxiom",
		Content:      t.Subject,
	}, nil
}

func (p *Parser) parsePropertyDomain(t types.Triple) (Axiom, error) {
	return &ObjectPropertyDomain{
		Property: &ObjectProperty{URI: URI(t.Subject)},
		Domain:   &Class{URI: URI(t.Object)},
	}, nil
}

func (p *Parser) parsePropertyRange(t types.Triple) (Axiom, error) {
	return &ObjectPropertyRange{
		Property: &ObjectProperty{URI: URI(t.Subject)},
		Range:    &Class{URI: URI(t.Object)},
	}, nil
}

func (p *Parser) ParseClassExpression(uri string) ClassExpression {
	return &Class{URI: URI(uri)}
}

func (p *Parser) ParseObjectProperty(uri string) *ObjectProperty {
	return &ObjectProperty{URI: URI(uri)}
}
