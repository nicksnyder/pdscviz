package main

import (
	"flag"
	"regexp"
)

type RegexpValue struct {
	regexp *regexp.Regexp
}

func (rv *RegexpValue) Set(s string) error {
	var err error
	rv.regexp, err = regexp.Compile(s)
	return err
}

func (rv *RegexpValue) String() string {
	if rv.regexp == nil {
		return ""
	}
	return rv.regexp.String()
}

var _ = flag.Value(&RegexpValue{})
