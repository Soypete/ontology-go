package jsonld_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/soypete/ontology-go/jsonld"
	"github.com/soypete/ontology-go/reasoner"
)

const testTTL = `
@prefix owl:  <http://www.w3.org/2002/07/owl#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix edu:  <https://ontology.example.org/edu#> .

<https://ontology.example.org/edu>
    a owl:Ontology .

edu:Person a owl:Class ; rdfs:label "Person" .
edu:Student a owl:Class ; rdfs:subClassOf edu:Person ; rdfs:label "Student" .
edu:Teacher a owl:Class ; rdfs:subClassOf edu:Person ; rdfs:label "Teacher" .
edu:Course a owl:Class ; rdfs:label "Course" .
`

func loadOnt(t *testing.T) *reasoner.Ontology {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "test.ttl"), []byte(testTTL), 0o600); err != nil {
		t.Fatal(err)
	}
	ont, err := reasoner.LoadOntology(dir)
	if err != nil {
		t.Fatal(err)
	}
	return ont
}

func TestBuild_ExpandsTypes(t *testing.T) {
	ont := loadOnt(t)
	vertices := []jsonld.Vertex{
		{
			ID:         "https://data.example.org/Student/123",
			OWLClass:   "https://ontology.example.org/edu#Student",
			Properties: json.RawMessage(`{"name":"Alice"}`),
		},
	}
	resp := jsonld.Build(vertices, nil, ont)

	if len(resp.Graph) != 1 {
		t.Fatalf("expected 1 node, got %d", len(resp.Graph))
	}
	node := resp.Graph[0]
	if node.ID != "https://data.example.org/Student/123" {
		t.Errorf("node @id = %s", node.ID)
	}
	hasStudent := false
	hasPerson := false
	for _, typ := range node.Type {
		if typ == "https://ontology.example.org/edu#Student" {
			hasStudent = true
		}
		if typ == "https://ontology.example.org/edu#Person" {
			hasPerson = true
		}
	}
	if !hasStudent {
		t.Error("missing Student type")
	}
	if !hasPerson {
		t.Error("missing inferred Person type")
	}
}

func TestBuild_IncludesEdges(t *testing.T) {
	ont := loadOnt(t)
	edges := []jsonld.Edge{
		{
			VertexIRI: "https://data.example.org/Student/123",
			EdgeIRI:   "https://ontology.example.org/edu#enrolledIn",
			NodeIRI:   "https://data.example.org/Course/456",
		},
	}
	resp := jsonld.Build(nil, edges, ont)

	if len(resp.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
	}
	if resp.Edges[0].Edge != "https://ontology.example.org/edu#enrolledIn" {
		t.Errorf("edge edge_iri = %s", resp.Edges[0].Edge)
	}
}

func TestBuild_GroupsDuplicateIRIs(t *testing.T) {
	ont := loadOnt(t)
	const iri = "http://data.example.org/edu#teacher-fixture"
	vertices := []jsonld.Vertex{
		{
			ID:         iri,
			OWLClass:   "https://ontology.example.org/edu#Person",
			Properties: json.RawMessage(`{"name":"Pat"}`),
			CreatedAt:  "2026-04-27 23:28:22.041974+00",
			UpdatedAt:  "2026-04-29 21:54:44.249779+00",
		},
		{
			ID:         iri,
			OWLClass:   "https://ontology.example.org/edu#Student",
			Properties: json.RawMessage(`{"name":"Pat"}`),
			CreatedAt:  "2026-04-27 23:28:22.041974+00",
			UpdatedAt:  "2026-04-29 21:54:44.249779+00",
		},
	}
	resp := jsonld.Build(vertices, nil, ont)

	if len(resp.Graph) != 1 {
		t.Fatalf("expected 1 grouped node, got %d", len(resp.Graph))
	}
	got := resp.Graph[0]
	if got.ID != iri {
		t.Errorf("node @id = %s", got.ID)
	}

	want := map[string]bool{
		"https://ontology.example.org/edu#Student": false,
		"https://ontology.example.org/edu#Person":  false,
	}
	for _, typ := range got.Type {
		if _, expected := want[typ]; expected {
			want[typ] = true
		}
	}
	for class, seen := range want {
		if !seen {
			t.Errorf("@type missing %s; got %v", class, got.Type)
		}
	}

	counts := map[string]int{}
	for _, typ := range got.Type {
		counts[typ]++
	}
	for class, n := range counts {
		if n > 1 {
			t.Errorf("@type contains %s %d times; should be deduplicated", class, n)
		}
	}
}

func TestBuild_GroupsPreserveFirstAppearanceOrder(t *testing.T) {
	ont := loadOnt(t)
	vertices := []jsonld.Vertex{
		{ID: "edu#Person_A", OWLClass: "https://ontology.example.org/edu#Person"},
		{ID: "edu#Person_B", OWLClass: "https://ontology.example.org/edu#Person"},
		{ID: "edu#Person_A", OWLClass: "https://ontology.example.org/edu#Student"},
	}
	resp := jsonld.Build(vertices, nil, ont)
	if len(resp.Graph) != 2 {
		t.Fatalf("expected 2 grouped nodes, got %d", len(resp.Graph))
	}
	const (
		wantA = "http://data.local/edu#Person_A"
		wantB = "http://data.local/edu#Person_B"
	)
	if resp.Graph[0].ID != wantA || resp.Graph[1].ID != wantB {
		t.Errorf("node order = [%s, %s]; want [%s, %s]", resp.Graph[0].ID, resp.Graph[1].ID, wantA, wantB)
	}
}

func TestBuild_ContextPrefixes(t *testing.T) {
	ont := loadOnt(t)
	resp := jsonld.Build(nil, nil, ont)

	if resp.Context["rdf"] != "http://www.w3.org/1999/02/22-rdf-syntax-ns#" {
		t.Errorf("context[rdf] = %s", resp.Context["rdf"])
	}
	if resp.Context["rdfs"] != "http://www.w3.org/2000/01/rdf-schema#" {
		t.Errorf("context[rdfs] = %s", resp.Context["rdfs"])
	}
	if resp.Context["owl"] != "http://www.w3.org/2002/07/owl#" {
		t.Errorf("context[owl] = %s", resp.Context["owl"])
	}
}

func TestBuild_WithCustomPrefixMap(t *testing.T) {
	ont := loadOnt(t)
	customPrefixes := map[string]string{
		"edu": "https://ontology.example.org/edu#",
		"sai": "https://ontology.example.org/sai#",
	}
	vertices := []jsonld.Vertex{
		{ID: "edu#Person_user-1", OWLClass: "https://ontology.example.org/edu#Person"},
	}
	resp := jsonld.Build(vertices, nil, ont, jsonld.WithPrefixMap(customPrefixes))

	if resp.Context["edu"] != "https://ontology.example.org/edu#" {
		t.Errorf("context[edu] = %s", resp.Context["edu"])
	}
	if resp.Context["sai"] != "https://ontology.example.org/sai#" {
		t.Errorf("context[sai] = %s", resp.Context["sai"])
	}
	if resp.Graph[0].ID != "https://ontology.example.org/edu#Person_user-1" {
		t.Errorf("expanded ID = %s", resp.Graph[0].ID)
	}
}

func TestBuild_WithCustomBaseURL(t *testing.T) {
	ont := loadOnt(t)
	vertices := []jsonld.Vertex{
		{ID: "myresource-1", OWLClass: "https://ontology.example.org/edu#Person"},
	}
	resp := jsonld.Build(vertices, nil, ont, jsonld.WithBaseURL("https://custom.data.example.org"))

	if resp.Graph[0].ID != "https://custom.data.example.org/myresource-1" {
		t.Errorf("expanded ID = %s", resp.Graph[0].ID)
	}
}

func TestBuild_IRIExpansion_FullURLPassThrough(t *testing.T) {
	ont := loadOnt(t)
	vertices := []jsonld.Vertex{
		{ID: "http://example.org/Person_1", OWLClass: "https://ontology.example.org/edu#Person"},
		{ID: "https://example.org/Person_2", OWLClass: "https://ontology.example.org/edu#Person"},
	}
	resp := jsonld.Build(vertices, nil, ont)

	if resp.Graph[0].ID != "http://example.org/Person_1" {
		t.Errorf("expected http://example.org/Person_1, got %s", resp.Graph[0].ID)
	}
	if resp.Graph[1].ID != "https://example.org/Person_2" {
		t.Errorf("expected https://example.org/Person_2, got %s", resp.Graph[1].ID)
	}
}

func TestBuild_WithoutOntology(t *testing.T) {
	vertices := []jsonld.Vertex{
		{ID: "https://data.example.org/Student/123", OWLClass: "https://ontology.example.org/edu#Student"},
	}
	edges := []jsonld.Edge{
		{
			VertexIRI: "https://data.example.org/Student/123",
			EdgeIRI:   "https://ontology.example.org/edu#enrolledIn",
			NodeIRI:   "https://data.example.org/Course/456",
		},
	}
	resp := jsonld.Build(vertices, edges, nil)

	if len(resp.Graph) != 1 {
		t.Fatalf("expected 1 node, got %d", len(resp.Graph))
	}
	if len(resp.Edges) != 1 {
		t.Fatalf("expected 1 edge, got %d", len(resp.Edges))
	}
	if len(resp.Graph[0].Type) != 1 {
		t.Errorf("expected 1 type without ontology, got %d", len(resp.Graph[0].Type))
	}
	if resp.Graph[0].Type[0] != "https://ontology.example.org/edu#Student" {
		t.Errorf("expected Student type, got %s", resp.Graph[0].Type[0])
	}
}

func TestBuild_EdgeExpansion(t *testing.T) {
	ont := loadOnt(t)
	customPrefixes := map[string]string{
		"edu": "https://ontology.example.org/edu#",
	}
	edges := []jsonld.Edge{
		{
			VertexIRI: "edu#Person_user-1",
			EdgeIRI:   "https://ontology.example.org/edu#enrolledIn",
			NodeIRI:   "edu#Course_course-1",
		},
	}
	resp := jsonld.Build(nil, edges, ont, jsonld.WithPrefixMap(customPrefixes))

	if resp.Edges[0].Vertex != "https://ontology.example.org/edu#Person_user-1" {
		t.Errorf("Edge.Vertex = %s", resp.Edges[0].Vertex)
	}
	if resp.Edges[0].Node != "https://ontology.example.org/edu#Course_course-1" {
		t.Errorf("Edge.Node = %s", resp.Edges[0].Node)
	}
	if resp.Edges[0].Edge != "https://ontology.example.org/edu#enrolledIn" {
		t.Errorf("Edge.Edge = %s", resp.Edges[0].Edge)
	}
}

func TestBuild_EmptyInputs(t *testing.T) {
	ont := loadOnt(t)

	resp := jsonld.Build(nil, nil, ont)
	if len(resp.Graph) != 0 {
		t.Errorf("expected 0 nodes, got %d", len(resp.Graph))
	}
	if len(resp.Edges) != 0 {
		t.Errorf("expected 0 edges, got %d", len(resp.Edges))
	}
}

func TestExpandCompactIRI_WithPrefixMap(t *testing.T) {
	b := jsonld.NewBuilder(jsonld.WithPrefixMap(map[string]string{
		"edu": "https://ontology.example.org/edu#",
	}))

	tests := []struct {
		name     string
		iri      string
		expected string
	}{
		{"compact with prefix", "edu#Person_1", "https://ontology.example.org/edu#Person_1"},
		{"full URL", "https://example.org/Person_1", "https://example.org/Person_1"},
		{"no prefix", "Person_1", "http://data.local/Person_1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := b.ExpandCompactIRI(tt.iri)
			if result != tt.expected {
				t.Errorf("ExpandCompactIRI(%q) = %q, want %q", tt.iri, result, tt.expected)
			}
		})
	}
}

func TestExpandCompactIRI_WithBaseURL(t *testing.T) {
	b := jsonld.NewBuilder(jsonld.WithBaseURL("https://data.example.org"))

	result := b.ExpandCompactIRI("resource-123")
	expected := "https://data.example.org/resource-123"
	if result != expected {
		t.Errorf("ExpandCompactIRI(%q) = %q, want %q", "resource-123", result, expected)
	}
}

func TestBuild_TransitiveTypeExpansion(t *testing.T) {
	const deepTTL = `
@prefix owl:  <http://www.w3.org/2002/07/owl#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix edu:  <https://ontology.example.org/edu#> .

<https://ontology.example.org/edu> a owl:Ontology .

edu:Thing a owl:Class .
edu:Entity a owl:Class ; rdfs:subClassOf edu:Thing .
edu:Person a owl:Class ; rdfs:subClassOf edu:Entity .
edu:Student a owl:Class ; rdfs:subClassOf edu:Person .
	`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "deep.ttl"), []byte(deepTTL), 0o600); err != nil {
		t.Fatal(err)
	}
	ont, err := reasoner.LoadOntology(dir)
	if err != nil {
		t.Fatal(err)
	}

	vertices := []jsonld.Vertex{
		{ID: "https://data.example.org/Student/1", OWLClass: "https://ontology.example.org/edu#Student"},
	}
	resp := jsonld.Build(vertices, nil, ont)

	hasStudent := false
	hasPerson := false
	hasEntity := false
	hasThing := false

	for _, typ := range resp.Graph[0].Type {
		switch typ {
		case "https://ontology.example.org/edu#Student":
			hasStudent = true
		case "https://ontology.example.org/edu#Person":
			hasPerson = true
		case "https://ontology.example.org/edu#Entity":
			hasEntity = true
		case "https://ontology.example.org/edu#Thing":
			hasThing = true
		}
	}

	if !hasStudent {
		t.Error("missing Student type")
	}
	if !hasPerson {
		t.Error("missing Person type")
	}
	if !hasEntity {
		t.Error("missing Entity type")
	}
	if !hasThing {
		t.Error("missing Thing type")
	}
}

func TestBuild_PropertiesPreserved(t *testing.T) {
	ont := loadOnt(t)
	props := json.RawMessage(`{"name":"Alice","age":25,"active":true}`)

	vertices := []jsonld.Vertex{
		{
			ID:         "https://data.example.org/Person/1",
			OWLClass:   "https://ontology.example.org/edu#Person",
			Properties: props,
		},
	}
	resp := jsonld.Build(vertices, nil, ont)

	var gotProps map[string]interface{}
	if err := json.Unmarshal(resp.Graph[0].Properties, &gotProps); err != nil {
		t.Fatalf("failed to unmarshal properties: %v", err)
	}

	if gotProps["name"] != "Alice" {
		t.Errorf("name = %v", gotProps["name"])
	}
	if gotProps["age"] != 25.0 {
		t.Errorf("age = %v", gotProps["age"])
	}
	if gotProps["active"] != true {
		t.Errorf("active = %v", gotProps["active"])
	}
}

func TestBuild_EdgePropertiesPreserved(t *testing.T) {
	ont := loadOnt(t)
	props := json.RawMessage(`{"startDate":"2026-01-01","endDate":"2026-12-31"}`)

	edges := []jsonld.Edge{
		{
			VertexIRI:  "https://data.example.org/Student/1",
			EdgeIRI:    "https://ontology.example.org/edu#enrolledIn",
			NodeIRI:    "https://data.example.org/Course/1",
			Properties: props,
		},
	}
	resp := jsonld.Build(nil, edges, ont)

	var gotProps map[string]interface{}
	if err := json.Unmarshal(resp.Edges[0].Properties, &gotProps); err != nil {
		t.Fatalf("failed to unmarshal edge properties: %v", err)
	}

	if gotProps["startDate"] != "2026-01-01" {
		t.Errorf("startDate = %v", gotProps["startDate"])
	}
}

func TestBuild_TimestampsPreserved(t *testing.T) {
	ont := loadOnt(t)
	vertices := []jsonld.Vertex{
		{
			ID:        "https://data.example.org/Person/1",
			OWLClass:  "https://ontology.example.org/edu#Person",
			CreatedAt: "2026-01-01T00:00:00Z",
			UpdatedAt: "2026-06-01T00:00:00Z",
		},
	}
	resp := jsonld.Build(vertices, nil, ont)

	if resp.Graph[0].CreatedAt != "2026-01-01T00:00:00Z" {
		t.Errorf("CreatedAt = %s", resp.Graph[0].CreatedAt)
	}
	if resp.Graph[0].UpdatedAt != "2026-06-01T00:00:00Z" {
		t.Errorf("UpdatedAt = %s", resp.Graph[0].UpdatedAt)
	}
}

func TestBuilder_OptionOverrides(t *testing.T) {
	ont := loadOnt(t)

	b := jsonld.NewBuilder(
		jsonld.WithBaseURL("https://base1.example.org"),
		jsonld.WithBaseURL("https://base2.example.org"),
		jsonld.WithPrefixMap(map[string]string{"test": "https://test.example.org#"}),
	)

	vertices := []jsonld.Vertex{
		{ID: "test#resource", OWLClass: "https://ontology.example.org/edu#Person"},
	}
	resp := b.Build(vertices, nil, ont)

	if resp.Graph[0].ID != "https://test.example.org#resource" {
		t.Errorf("ID = %s", resp.Graph[0].ID)
	}
	if resp.Context["test"] != "https://test.example.org#" {
		t.Errorf("context[test] = %s", resp.Context["test"])
	}
}

func TestBuild_MultipleEdgesFromSameVertex(t *testing.T) {
	ont := loadOnt(t)
	vertices := []jsonld.Vertex{
		{ID: "https://data.example.org/Student/1", OWLClass: "https://ontology.example.org/edu#Student"},
	}
	edges := []jsonld.Edge{
		{VertexIRI: "https://data.example.org/Student/1", EdgeIRI: "https://ontology.example.org/edu#enrolledIn", NodeIRI: "https://data.example.org/Course/1"},
		{VertexIRI: "https://data.example.org/Student/1", EdgeIRI: "https://ontology.example.org/edu#likes", NodeIRI: "https://data.example.org/Course/2"},
		{VertexIRI: "https://data.example.org/Student/1", EdgeIRI: "https://ontology.example.org/edu#attends", NodeIRI: "https://data.example.org/School/1"},
	}
	resp := jsonld.Build(vertices, edges, ont)

	if len(resp.Edges) != 3 {
		t.Fatalf("expected 3 edges, got %d", len(resp.Edges))
	}

	expectedPredicates := []string{
		"https://ontology.example.org/edu#enrolledIn",
		"https://ontology.example.org/edu#likes",
		"https://ontology.example.org/edu#attends",
	}
	for i, expected := range expectedPredicates {
		if resp.Edges[i].Edge != expected {
			t.Errorf("Edge[%d] = %s, want %s", i, resp.Edges[i].Edge, expected)
		}
	}
}

func TestBuild_LargeDataset(t *testing.T) {
	ont := loadOnt(t)

	vertices := make([]jsonld.Vertex, 100)
	for i := 0; i < 100; i++ {
		vertices[i] = jsonld.Vertex{
			ID:         fmt.Sprintf("https://data.example.org/Person/%d", i),
			OWLClass:   "https://ontology.example.org/edu#Person",
			Properties: json.RawMessage(`{"index":` + fmt.Sprint(i) + `}`),
		}
	}

	resp := jsonld.Build(vertices, nil, ont)

	if len(resp.Graph) != 100 {
		t.Errorf("expected 100 nodes, got %d", len(resp.Graph))
	}
}

func TestBuild_ComplexTypeHierarchy(t *testing.T) {
	const complexTTL = `
@prefix owl:  <http://www.w3.org/2002/07/owl#> .
@prefix rdfs: <http://www.w3.org/2000/01/rdf-schema#> .
@prefix edu:  <https://ontology.example.org/edu#> .

<https://ontology.example.org/edu> a owl:Ontology .

edu:Agent a owl:Class .
edu:Person a owl:Class ; rdfs:subClassOf edu:Agent .
edu:Organization a owl:Class ; rdfs:subClassOf edu:Agent .
edu:Student a owl:Class ; rdfs:subClassOf edu:Person .
edu:Instructor a owl:Class ; rdfs:subClassOf edu:Person .
	`

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "complex.ttl"), []byte(complexTTL), 0o600); err != nil {
		t.Fatal(err)
	}
	ont, err := reasoner.LoadOntology(dir)
	if err != nil {
		t.Fatal(err)
	}

	vertices := []jsonld.Vertex{
		{ID: "https://data.example.org/Student/1", OWLClass: "https://ontology.example.org/edu#Student"},
		{ID: "https://data.example.org/Instructor/1", OWLClass: "https://ontology.example.org/edu#Instructor"},
	}
	resp := jsonld.Build(vertices, nil, ont)

	studentTypes := resp.Graph[0].Type
	instructorTypes := resp.Graph[1].Type

	studentTypesSet := make(map[string]bool)
	for _, t := range studentTypes {
		studentTypesSet[t] = true
	}

	instructorTypesSet := make(map[string]bool)
	for _, t := range instructorTypes {
		instructorTypesSet[t] = true
	}

	if !studentTypesSet["https://ontology.example.org/edu#Student"] {
		t.Error("Student missing its own type")
	}
	if !studentTypesSet["https://ontology.example.org/edu#Person"] {
		t.Error("Student missing Person type")
	}
	if !studentTypesSet["https://ontology.example.org/edu#Agent"] {
		t.Error("Student missing Agent type")
	}

	if !instructorTypesSet["https://ontology.example.org/edu#Instructor"] {
		t.Error("Instructor missing its own type")
	}
	if !instructorTypesSet["https://ontology.example.org/edu#Person"] {
		t.Error("Instructor missing Person type")
	}
	if !instructorTypesSet["https://ontology.example.org/edu#Agent"] {
		t.Error("Instructor missing Agent type")
	}
}
