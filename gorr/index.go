package gorr

import (
	"sync"

	"github.com/soypete/ontology-go/gorr/owl"
)

type OntologyIndex struct {
	mu               sync.Mutex
	classExprs       map[Handle]owl.ClassExpression
	objectProperties map[Handle]*owl.ObjectProperty

	classHierarchy       map[Handle]HandleSet
	propertyHierarchy    map[Handle]HandleSet
	objectPropertyDomain map[Handle]Handle
	objectPropertyRange  map[Handle]Handle

	axioms      []owl.Axiom
	nextHandle  Handle
	handleToURI map[Handle]string
	uriToHandle map[string]Handle
}

func NewOntologyIndex() *OntologyIndex {
	return &OntologyIndex{
		classExprs:           make(map[Handle]owl.ClassExpression),
		objectProperties:     make(map[Handle]*owl.ObjectProperty),
		classHierarchy:       make(map[Handle]HandleSet),
		propertyHierarchy:    make(map[Handle]HandleSet),
		objectPropertyDomain: make(map[Handle]Handle),
		objectPropertyRange:  make(map[Handle]Handle),
		handleToURI:          make(map[Handle]string),
		uriToHandle:          make(map[string]Handle),
	}
}

func (idx *OntologyIndex) AddAxiom(ax owl.Axiom) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.axioms = append(idx.axioms, ax)
	idx.addAxiomUnlocked(ax)
}

func (idx *OntologyIndex) AddAxioms(axs []owl.Axiom) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	for _, ax := range axs {
		idx.axioms = append(idx.axioms, ax)
		idx.addAxiomUnlocked(ax)
	}
}

func (idx *OntologyIndex) addAxiomUnlocked(ax owl.Axiom) {
	switch a := ax.(type) {
	case *owl.SubClassOf:
		subHandle := idx.internClassUnlocked(a.SubClass)
		superHandle := idx.internClassUnlocked(a.SuperClass)
		idx.addSubsumptionUnlocked(subHandle, superHandle)

	case *owl.EquivalentClasses:
		handles := idx.internClassExpressionsUnlocked(a.Classes)
		for i := 0; i < len(handles); i++ {
			for j := i + 1; j < len(handles); j++ {
				idx.addSubsumptionUnlocked(handles[i], handles[j])
				idx.addSubsumptionUnlocked(handles[j], handles[i])
			}
		}

	case *owl.DisjointClasses:

	case *owl.SubObjectPropertyOf:
		subProp, _ := a.SubProperty.(*owl.ObjectProperty)
		superProp, _ := a.SuperProperty.(*owl.ObjectProperty)
		if subProp != nil && superProp != nil {
			subHandle := idx.internUnlocked(string(subProp.URI))
			superHandle := idx.internUnlocked(string(superProp.URI))
			idx.addPropertySubsumptionUnlocked(subHandle, superHandle)
		}

	case *owl.ObjectPropertyDomain:
		prop, _ := a.Property.(*owl.ObjectProperty)
		if prop != nil {
			propHandle := idx.internUnlocked(string(prop.URI))
			domainHandle := idx.internClassUnlocked(a.Domain)
			idx.objectPropertyDomain[propHandle] = domainHandle
		}

	case *owl.ObjectPropertyRange:
		prop, _ := a.Property.(*owl.ObjectProperty)
		if prop != nil {
			propHandle := idx.internUnlocked(string(prop.URI))
			rangeHandle := idx.internClassUnlocked(a.Range)
			idx.objectPropertyRange[propHandle] = rangeHandle
		}
	}
}

func (idx *OntologyIndex) internUnlocked(uri string) Handle {
	if h, ok := idx.uriToHandle[uri]; ok {
		return h
	}
	idx.nextHandle++
	h := idx.nextHandle
	idx.uriToHandle[uri] = h
	idx.handleToURI[h] = uri
	return h
}

func (idx *OntologyIndex) internClassUnlocked(expr owl.ClassExpression) Handle {
	handle := idx.internUnlocked(expr.String())
	idx.classExprs[handle] = expr
	return handle
}

func (idx *OntologyIndex) internClassExpressionsUnlocked(exprs []owl.ClassExpression) []Handle {
	var handles []Handle
	for _, expr := range exprs {
		handles = append(handles, idx.internClassUnlocked(expr))
	}
	return handles
}

func (idx *OntologyIndex) internObjectPropertyUnlocked(prop *owl.ObjectProperty) Handle {
	handle := idx.internUnlocked(string(prop.URI))
	idx.objectProperties[handle] = prop
	return handle
}

func (idx *OntologyIndex) addSubsumptionUnlocked(sub, super Handle) {
	if idx.classHierarchy == nil {
		idx.classHierarchy = make(map[Handle]HandleSet)
	}
	current := idx.classHierarchy[sub]
	current.Add(super)
	idx.classHierarchy[sub] = current
}

func (idx *OntologyIndex) addPropertySubsumptionUnlocked(sub, super Handle) {
	if idx.propertyHierarchy == nil {
		idx.propertyHierarchy = make(map[Handle]HandleSet)
	}
	current := idx.propertyHierarchy[sub]
	current.Add(super)
	idx.propertyHierarchy[sub] = current
}

func (idx *OntologyIndex) URI(handle Handle) string {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.handleToURI[handle]
}

func (idx *OntologyIndex) intern(uri string) Handle {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.internUnlocked(uri)
}

func (idx *OntologyIndex) internClass(expr owl.ClassExpression) Handle {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.internClassUnlocked(expr)
}

func (idx *OntologyIndex) GetClass(handle Handle) (owl.ClassExpression, bool) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	expr, ok := idx.classExprs[handle]
	return expr, ok
}

func (idx *OntologyIndex) GetSubsumers(handle Handle) HandleSet {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.classHierarchy[handle]
}

func (idx *OntologyIndex) GetPropertySubsumers(handle Handle) HandleSet {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return idx.propertyHierarchy[handle]
}

func (idx *OntologyIndex) GetPropertyDomain(handle Handle) (Handle, bool) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	domain, ok := idx.objectPropertyDomain[handle]
	return domain, ok
}

func (idx *OntologyIndex) GetPropertyRange(handle Handle) (Handle, bool) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	rng, ok := idx.objectPropertyRange[handle]
	return rng, ok
}

func (idx *OntologyIndex) Axioms() []owl.Axiom {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	result := make([]owl.Axiom, len(idx.axioms))
	copy(result, idx.axioms)
	return result
}

func (idx *OntologyIndex) ClassCount() int {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return len(idx.classExprs)
}

func (idx *OntologyIndex) PropertyCount() int {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	return len(idx.objectProperties)
}

func (idx *OntologyIndex) GetObjectProperty(handle Handle) (*owl.ObjectProperty, bool) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	prop, ok := idx.objectProperties[handle]
	return prop, ok
}
