package jsonld

import (
	"encoding/json"
	"strings"

	"github.com/soypete/ontology-go/reasoner"
)

var DefaultContext = map[string]string{
	"rdf":  "http://www.w3.org/1999/02/22-rdf-syntax-ns#",
	"rdfs": "http://www.w3.org/2000/01/rdf-schema#",
	"owl":  "http://www.w3.org/2002/07/owl#",
	"xsd":  "http://www.w3.org/2001/XMLSchema#",
}

type Vertex struct {
	ID         string          `json:"id"`
	OWLClass   string          `json:"owl_class"`
	Properties json.RawMessage `json:"properties,omitempty"`
	CreatedAt  string          `json:"created_at,omitempty"`
	UpdatedAt  string          `json:"updated_at,omitempty"`
}

type Edge struct {
	VertexIRI  string          `json:"vertex_iri"`
	EdgeIRI    string          `json:"edge_iri"`
	NodeIRI    string          `json:"node_iri"`
	Properties json.RawMessage `json:"properties,omitempty"`
}

type Node struct {
	ID         string          `json:"@id"`
	Type       []string        `json:"@type"`
	Properties json.RawMessage `json:"properties,omitempty"`
	CreatedAt  string          `json:"created_at,omitempty"`
	UpdatedAt  string          `json:"updated_at,omitempty"`
}

type EdgeNode struct {
	Vertex     string          `json:"vertex"`
	Edge       string          `json:"edge"`
	Node       string          `json:"node"`
	Properties json.RawMessage `json:"properties,omitempty"`
}

type Response struct {
	Context map[string]string `json:"@context"`
	Graph   []Node            `json:"@graph"`
	Edges   []EdgeNode        `json:"edges,omitempty"`
}

type Builder struct {
	prefixMap map[string]string
	baseURL   string
}

type Option func(*Builder)

func WithPrefixMap(m map[string]string) Option {
	return func(b *Builder) {
		b.prefixMap = m
	}
}

func WithBaseURL(base string) Option {
	return func(b *Builder) {
		b.baseURL = base
	}
}

func NewBuilder(opts ...Option) *Builder {
	b := &Builder{
		prefixMap: make(map[string]string),
		baseURL:   "http://data.local",
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

func (b *Builder) Build(vertices []Vertex, edges []Edge, ont *reasoner.Ontology) *Response {
	type group struct {
		seed    Vertex
		classes []string
	}
	order := make([]string, 0, len(vertices))
	groups := make(map[string]*group, len(vertices))
	for _, v := range vertices {
		g, ok := groups[v.ID]
		if !ok {
			order = append(order, v.ID)
			g = &group{seed: v}
			groups[v.ID] = g
		}
		g.classes = append(g.classes, v.OWLClass)
	}

	nodes := make([]Node, 0, len(order))
	for _, id := range order {
		g := groups[id]
		nodes = append(nodes, Node{
			ID:         b.ExpandCompactIRI(g.seed.ID),
			Type:       b.unionExpandedTypes(g.classes, ont),
			Properties: g.seed.Properties,
			CreatedAt:  g.seed.CreatedAt,
			UpdatedAt:  g.seed.UpdatedAt,
		})
	}

	edgeNodes := make([]EdgeNode, 0, len(edges))
	for _, e := range edges {
		edgeNodes = append(edgeNodes, EdgeNode{
			Vertex:     b.ExpandCompactIRI(e.VertexIRI),
			Edge:       e.EdgeIRI,
			Node:       b.ExpandCompactIRI(e.NodeIRI),
			Properties: e.Properties,
		})
	}

	context := make(map[string]string)
	for k, v := range DefaultContext {
		context[k] = v
	}
	for k, v := range b.prefixMap {
		context[k] = v
	}

	return &Response{
		Context: context,
		Graph:   nodes,
		Edges:   edgeNodes,
	}
}

func (b *Builder) expandTypes(owlClass string, ont *reasoner.Ontology) []string {
	types := []string{owlClass}
	if ont == nil {
		return types
	}
	cls, ok := ont.ResolveClass(owlClass)
	if !ok {
		return types
	}
	ancestors := ont.GetClassAncestors(cls.ID)
	for _, a := range ancestors {
		types = append(types, a.IRI)
	}
	return types
}

func (b *Builder) unionExpandedTypes(classes []string, ont *reasoner.Ontology) []string {
	seen := make(map[string]struct{}, len(classes)*2)
	out := make([]string, 0, len(classes)*2)
	for _, c := range classes {
		for _, t := range b.expandTypes(c, ont) {
			if _, dup := seen[t]; dup {
				continue
			}
			seen[t] = struct{}{}
			out = append(out, t)
		}
	}
	return out
}

func (b *Builder) ExpandCompactIRI(iri string) string {
	if strings.HasPrefix(iri, "http://") || strings.HasPrefix(iri, "https://") {
		return iri
	}

	// Try both : and # as separators
	var prefix, local string

	// Check for # first (e.g., edu#Person_1)
	if idx := strings.Index(iri, "#"); idx != -1 {
		prefix = iri[:idx]
		local = iri[idx+1:]
	} else if idx := strings.Index(iri, ":"); idx != -1 {
		// Fall back to : (e.g., edu:Person_1)
		prefix = iri[:idx]
		local = iri[idx+1:]
	} else {
		// No separator, treat as local name
		return b.baseURL + "/" + iri
	}

	base, ok := b.prefixMap[prefix]
	if !ok {
		return b.baseURL + "/" + iri
	}

	return base + local
}

func Build(vertices []Vertex, edges []Edge, ont *reasoner.Ontology, opts ...Option) *Response {
	b := NewBuilder(opts...)
	return b.Build(vertices, edges, ont)
}
