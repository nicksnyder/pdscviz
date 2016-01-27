package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

type Graph struct {
	Namespace string
	Children  map[string][]Edge
	Parents   map[string][]Edge
}

func NewGraph(namespace string) *Graph {
	return &Graph{
		Namespace: namespace,
		Children:  make(map[string][]Edge),
		Parents:   make(map[string][]Edge),
	}
}

// Walk the nodes
func (g *Graph) walkParents(root string, depth int, visitor func(Edge)) {
	visited := make(map[string]struct{})
	g.walkEdges(root, g.Parents, visited, depth, func(e Edge) string {
		visitor(e)
		return e.From
	})
}

func (g *Graph) walkChildren(root string, depth int, visitor func(Edge)) {
	visited := make(map[string]struct{})
	g.walkEdges(root, g.Children, visited, depth, func(e Edge) string {
		visitor(e)
		return e.To
	})
}

// Walks all edges from the root node.
// Nodes are not visited twice.
// A negative depth will cause all edges to be walked.
func (g *Graph) walkEdges(root string, edges map[string][]Edge, visited map[string]struct{}, depth int, visitor func(Edge) string) {
	if depth == 0 {
		return
	}
	nextEdges := edges[root]
	for _, nextEdge := range nextEdges {
		nextRoot := visitor(nextEdge)
		verbosef("walking edge %s -> %s", root, nextRoot)
		if _, ok := visited[nextRoot]; !ok {
			visited[nextRoot] = struct{}{}
			g.walkEdges(nextRoot, edges, visited, depth-1, visitor)
		} else {
			verbosef("breaking cycle; already visited %s", nextRoot)
		}
	}
}

func (g *Graph) addEdge(parent, child string, collection bool) {
	verbosef("discovered edge %s -> %s", parent, child)
	e := Edge{
		From:       parent,
		To:         child,
		Collection: collection,
	}
	g.Children[parent] = append(g.Children[parent], e)
	g.Parents[child] = append(g.Parents[child], e)
}

func (g *Graph) visitPDSC(path string, info os.FileInfo, err error) error {
	if info.IsDir() {
		return nil
	}
	name := info.Name()
	if !strings.HasSuffix(name, ".pdsc") {
		return nil
	}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		fatalf("unable to read %s because %s\n", path, err)
	}

	// Strip out end of line comments since they aren't valid json.
	buf = stripRegex(buf, `//[^"\n]*`)

	// Strip out block comments since they aren't valid json.
	buf = stripRegex(buf, `/\*(\*[^/]|[^*])*\*/`)

	var pdsc PDSC
	if err := json.Unmarshal(buf, &pdsc); err != nil {
		verbosef("%s", buf)
		fatalf("unable to parse %s because %s\n", path, err)
	}
	parent := g.displayName(pdsc.fullyQualifiedName())
	for _, tr := range pdsc.typeRefs() {
		child := g.displayName(tr.Name)
		g.addEdge(parent, child, tr.Collection)
	}
	return nil
}

func stripRegex(buf []byte, re string) []byte {
	return regexp.MustCompile(re).ReplaceAllLiteral(buf, nil)
}

func (g *Graph) displayName(name string) string {
	return strings.TrimPrefix(name, g.Namespace)
}
