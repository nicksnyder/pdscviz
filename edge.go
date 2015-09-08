package main

import (
	"fmt"
	"strings"
)

type Edge struct {
	From, To   string
	Collection bool
}

func (e *Edge) graphvizFormat() string {
	attrs := []string{}
	if e.Collection {
		attrs = append(attrs, `color="#ff0000"`)
	}
	return fmt.Sprintf("\"%s\"->\"%s\" [%s]", e.From, e.To, strings.Join(attrs, ", "))
}
