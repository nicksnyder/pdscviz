package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

var usage = `%[1]s visualizes dependencies between Pegasus Data Schema (PDSC) files using Graphviz.

Usage:

    %[1]s [options]
		    Graphs all models.

    %[1]s [options] usages <root entity>
		    Graphs all models that transitively depend on <root entity>.

    %[1]s [options] dependencies <root entity>
		    Graphs all models that <root entity> transitively depends on.

Options:

`

var verbose bool

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
	}
	var out, dir, trimPrefix, graphAttrs string
	var exclude RegexpValue
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.StringVar(&out, "out", "", "the output file (defaults to stdout)")
	flag.StringVar(&dir, "dir", ".", "the directory to scan for PDSC files (defaults to the current directory)")
	flag.StringVar(&trimPrefix, "trimPrefix", "", "the prefix to remove from each type name")
	flag.StringVar(&graphAttrs, "graphAttrs", "", "extra attributes for the graph (see http://www.graphviz.org/content/attrs)")
	flag.Var(&exclude, "exclude", "nodes matching this regular expression will be excluded")
	flag.Parse()

	var recommendedCommand string
	var commandFunc func(*Graph) *GraphvizData
	switch command := flag.Arg(0); command {
	case "usages":
		commandFunc = func(g *Graph) *GraphvizData {
			root := flag.Arg(1)
			var edges []string
			g.walkParents(root, func(e Edge) {
				edges = append(edges, e.graphvizFormat())
			})
			return &GraphvizData{
				Root:  root,
				Edges: edges,
			}
		}
		recommendedCommand = "twopi"
	case "dependencies":
		commandFunc = func(g *Graph) *GraphvizData {
			root := flag.Arg(1)
			var edges []string
			g.walkChildren(root, func(e Edge) {
				edges = append(edges, e.graphvizFormat())
			})
			return &GraphvizData{
				Root:  root,
				Edges: edges,
			}
		}
		recommendedCommand = "twopi"
	case "":
		commandFunc = func(g *Graph) *GraphvizData {
			var edges []string
			for _, es := range g.Children {
				for _, e := range es {
					edges = append(edges, e.graphvizFormat())
				}
			}
			return &GraphvizData{
				Edges: edges,
			}
		}
		recommendedCommand = "neato"
	default:
		fatalf("unknown command: %s", command)
	}

	g := NewGraph(trimPrefix)
	infof("walking %s", dir)
	if err := filepath.Walk(dir, g.visitPDSC); err != nil {
		fatalf("finished walking with error: %s", err)
	}

	templateData := commandFunc(g)
	templateData.GraphAttrs = graphAttrs

	if exclude.regexp != nil {
		var edges []string
		for _, edge := range templateData.Edges {
			if !exclude.regexp.MatchString(edge) {
				edges = append(edges, edge)
			}
		}
		templateData.Edges = edges
	}

	t := template.Must(template.New("").Parse(`digraph G {
	node [shape="box"];
	fontsize=11.0;
	overlap=prism;
	{{if .GraphAttrs}}{{.GraphAttrs}};{{end}}
	{{if .Root}}root="{{.Root}}";{{end}}
	{{range .Edges}}{{.}};
	{{end}}
}
`))

	var graph bytes.Buffer
	if err := t.Execute(&graph, templateData); err != nil {
		fatalf("unable to execute template because %s", err)
	}

	if out != "" {
		outFile, err := os.OpenFile(out, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			fatalf("failed to write file %s because %s", out, err)
		}
		outFile.Write(graph.Bytes())
		infof("wrote graph to %s", out)
		infof("example: %s -Tpng %s > /tmp/pdsc.png && open /tmp/pdsc.png", recommendedCommand, out)
	} else {
		os.Stdout.Write(graph.Bytes())
		infof("wrote graph to stdout (redirect to file or use -out parameter)")
		infof("recommended graphviz command: %s", recommendedCommand)
		infof("example: %s | %s -Tpng > /tmp/pdsc.png && open /tmp/pdsc.png", strings.Join(os.Args, " "), recommendedCommand)
	}
}

func infof(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
}

func verbosef(format string, args ...interface{}) {
	if verbose {
		infof(format, args...)
	}
}

func fatalf(format string, args ...interface{}) {
	infof("fatal: "+format+"\n", args...)
	os.Exit(1)
}
