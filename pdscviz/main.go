package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

var usage = `%[1]s visualizes dependencies between Pegasus Data Schema (PDSC) files using Graphviz.

Usage:

    %[1]s [options] usages <root entity>
    %[1]s [options] dependencies <root entity>

Options:

`

var verbose bool

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usage, os.Args[0])
		flag.PrintDefaults()
	}
	var out, dir, trimPrefix, graphAttrs string
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.StringVar(&out, "out", "/tmp/pdsc.dot", "the output file")
	flag.StringVar(&dir, "dir", "", "the directory to scan for PDSC files (defaults to the current directory)")
	flag.StringVar(&trimPrefix, "trimPrefix", "", "the prefix to remove from each type name")
	flag.StringVar(&graphAttrs, "graphAttrs", "", "extra attributes for the graph (see http://www.graphviz.org/content/attrs)")
	flag.Parse()

	var commandFunc func(*Graph) map[string]interface{}
	command := flag.Arg(0)
	switch command {
	case "usages":
		commandFunc = func(g *Graph) map[string]interface{} {
			root := flag.Arg(1)
			var edges []string
			g.walkParents(root, func(e Edge) {
				edges = append(edges, e.graphvizFormat())
			})
			return map[string]interface{}{
				"Root":  root,
				"Edges": edges,
			}
		}
	case "dependencies":
		commandFunc = func(g *Graph) map[string]interface{} {
			root := flag.Arg(1)
			var edges []string
			g.walkChildren(root, func(e Edge) {
				edges = append(edges, e.graphvizFormat())
			})
			return map[string]interface{}{
				"Root":  root,
				"Edges": edges,
			}
		}
	default:
		fatalf("unknown command %s", command)
	}

	if dir == "" {
		var err error
		dir, err = os.Getwd()
		if err != nil {
			fatalf("unable to get current working directory: %s", err)
		}
	}

	g := NewGraph(trimPrefix)
	infof("walking %s", dir)
	if err := filepath.Walk(dir, g.visitPDSC); err != nil {
		fatalf("finished walking with error: %s", err)
	}

	templateData := commandFunc(g)
	templateData["GraphAttrs"] = graphAttrs

	t := template.Must(template.New("").Parse(`digraph G {
	fontsize=11.0;
	overlap=prism;
	{{if .GraphAttrs}}{{.GraphAttrs}};{{end}}
	{{if .Root}}root="{{.Root}}";{{end}}
	{{range .Edges}}
	  {{.}};
	{{end}}
}`))

	var graph bytes.Buffer
	if err := t.Execute(&graph, templateData); err != nil {
		fatalf("unable to execute template because %s", err)
	}
	if err := ioutil.WriteFile(out, graph.Bytes(), 0644); err != nil {
		fatalf("failed to write file %s because %s", out, err)
	}

	infof("wrote graph to %s", out)
	infof("cat %s | twopi -Tpng > /tmp/pdsc.png && open /tmp/pdsc.png", out)
}

func infof(format string, args ...interface{}) {
	fmt.Printf(format+"\n", args...)
}

func verbosef(format string, args ...interface{}) {
	if verbose {
		fmt.Printf(format+"\n", args...)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf("fatal: "+format+"\n", args...)
	os.Exit(1)
}
