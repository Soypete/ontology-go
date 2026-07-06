package gorr

import (
	"sync"

	"github.com/soypete/ontology-go/gorr/owl"
)

type OntologyIndex struct {
	mu               sync.RWMutex
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

func (idx *OntologyIndex) intern(uri string) Handle {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if h, ok := idx.uriToHandle[uri]; ok {
		return h
	}

	idx.nextHandle++
	h := idx.nextHandle
	idx.uriToHandle[uri] = h
	idx.handleToURI[h] = uri
	return h
}

func (idx *OntologyIndex) URI(handle Handle) string {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.handleToURI[handle]
}

func (idx *OntologyIndex) AddAxiom(ax owl.Axiom) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.axioms = append(idx.axioms, ax)

	switch a := ax.(type) {
	case *owl.SubClassOf:
		subHandle := idx.internClass(a.SubClass)
		superHandle := idx.internClass(a.SuperClass)
		idx.addSubsumption(subHandle, superHandle)

	case *owl.EquivalentClasses:
		handles := idx.internClassExpressions(a.Classes)
		for i := 0; i < len(handles); i++ {
			for j := i + 1; j < len(handles); j++ {
				idx.addSubsumption(handles[i], handles[j])
				idx.addSubsumption(handles[j], handles[i])
			}
		}

	case *owl.DisjointClasses:
		// Track disjointness for later processing

	case *owl.SubObjectPropertyOf:
		subProp, _ := a.SubProperty.(*owl.ObjectProperty)
		superProp, _ := a.SuperProperty.(*owl.ObjectProperty)
		if subProp != nil && superProp != nil {
			subHandle := idx.intern(string(subProp.URI))
			superHandle := idx.intern(string(superProp.URI))
			idx.addPropertySubsumption(subHandle, superHandle)
		}

	case *owl.ObjectPropertyDomain:
		prop, _ := a.Property.(*owl.ObjectProperty)
		if prop != nil {
			propHandle := idx.intern(string(prop.URI))
			domainHandle := idx.internClass(a.Domain)
			idx.objectPropertyDomain[propHandle] = domainHandle
		}

	case *owl.ObjectPropertyRange:
		prop, _ := a.Property.(*owl.ObjectProperty)
		if prop != nil {
			propHandle := idx.intern(string(prop.URI))
			rangeHandle := idx.internClass(a.Range)
			idx.objectPropertyRange[propHandle] = rangeHandle
		}
	}
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
		subHandle := idx.internClass(a.SubClass)
		superHandle := idx.internClass(a.SuperClass)
		idx.addSubsumption(subHandle, superHandle)

	case *owl.EquivalentClasses:
		handles := idx.internClassExpressions(a.Classes)
		for i := 0; i < len(handles); i++ {
			for j := i + 1; j < len(handles); j++ {
				idx.addSubsumption(handles[i], handles[j])
				idx.addSubsumption(handles[j], handles[i])
			}
		}

	case *owl.DisjointClasses:

	case *owl.SubObjectPropertyOf:
		subProp, _ := a.SubProperty.(*owl.ObjectProperty)
		superProp, _ := a.SuperProperty.(*owl.ObjectProperty)
		if subProp != nil && superProp != nil {
			subHandle := idx.intern(string(subProp.URI))
			superHandle := idx.intern(string(superProp.URI))
			idx.addPropertySubsumption(subHandle, superHandle)
		}

	case *owl.ObjectPropertyDomain:
		prop, _ := a.Property.(*owl.ObjectProperty)
		if prop != nil {
			propHandle := idx.intern(string(prop.URI))
			domainHandle := idx.internClass(a.Domain)
			idx.objectPropertyDomain[propHandle] = domainHandle
		}

	case *owl.ObjectPropertyRange:
		prop, _ := a.Property.(*owl.ObjectProperty)
		if prop != nil {
			propHandle := idx.intern(string(prop.URI))
			rangeHandle := idx.internClass(a.Range)
			idx.objectPropertyRange[propHandle] = rangeHandle
		}
	}
}

func (idx *OntologyIndex) addSubsumption(sub, super Handle) {
	if idx.classHierarchy == nil {
		idx.classHierarchy = make(map[Handle]HandleSet)
	}
	current := idx.classHierarchy[sub]
	current.Add(super)
	idx.classHierarchy[sub] = current
}

func (idx *OntologyIndex) addPropertySubsumption(sub, super Handle) {
	if idx.propertyHierarchy == nil {
		idx.propertyHierarchy = make(map[Handle]HandleSet)
	}
	current := idx.propertyHierarchy[sub]
	current.Add(super)
	idx.propertyHierarchy[sub] = current
}

func (idx *OntologyIndex) internClass(expr owl.ClassExpression) Handle {
	handle := idx.intern(expr.String())
	idx.classExprs[handle] = expr
	return handle
}

func (idx *OntologyIndex) internObjectProperty(prop *owl.ObjectProperty) Handle {
	handle := idx.intern(string(prop.URI))
	idx.objectProperties[handle] = prop
	return handle
}

func (idx *OntologyIndex) internClassExpressions(exprs []owl.ClassExpression) []Handle {
	var handles []Handle
	for _, expr := range exprs {
		handles = append(handles, idx.internClass(expr))
	}
	return handles
}

func (idx *OntologyIndex) GetClass(handle Handle) (owl.ClassExpression, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	expr, ok := idx.classExprs[handle]
	return expr, ok
}

func (idx *OntologyIndex) GetSubsumers(handle Handle) HandleSet {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.classHierarchy[handle]
}

func (idx *OntologyIndex) GetPropertySubsumers(handle Handle) HandleSet {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.propertyHierarchy[handle]
}

func (idx *OntologyIndex) GetPropertyDomain(handle Handle) (Handle, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	domain, ok := idx.objectPropertyDomain[handle]
	return domain, ok
}

func (idx *OntologyIndex) GetPropertyRange(handle Handle) (Handle, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	rng, ok := idx.objectPropertyRange[handle]
	return rng, ok
}

func (idx *OntologyIndex) Axioms() []owl.Axiom {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	result := make([]owl.Axiom, len(idx.axioms))
	copy(result, idx.axioms)
	return result
}

func (idx *OntologyIndex) ClassCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.classExprs)
}

func (idx *OntologyIndex) PropertyCount() int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return len(idx.objectProperties)
}

func (idx *OntologyIndex) GetObjectProperty(handle Handle) (*owl.ObjectProperty, bool) {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	prop, ok := idx.objectProperties[handle]
	return prop, ok
}
