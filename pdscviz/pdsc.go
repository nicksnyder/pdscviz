package main

import "strings"

type PDSC struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Ref       interface{}   `json:"ref"`
	Include   []interface{} `json:"include"`
	Fields    []PDSCField
}

type PDSCField struct {
	Type interface{} `json:"type"`
}

func (p *PDSC) fullyQualifiedName() string {
	return fullyQualifiedName(p.Namespace, p.Name)
}

func (p *PDSC) typeRefs() []TypeRef {
	var typeRefs []TypeRef
	if p.Ref != nil {
		typeRefs = append(typeRefs, p.resolveType(p.Ref, false)...)
	}
	if p.Include != nil {
		typeRefs = append(typeRefs, p.resolveType(p.Include, false)...)
	}
	for _, field := range p.Fields {
		typeRefs = append(typeRefs, p.resolveType(field.Type, false)...)
	}
	return typeRefs
}

func (p *PDSC) resolveType(t interface{}, collection bool) []TypeRef {
	switch t := t.(type) {
	case string:
		if isPrimitive(t) {
			return nil
		}
		fqn := fullyQualifiedName(p.Namespace, t)
		return []TypeRef{{Name: fqn, Collection: collection}}
	case map[string]interface{}:
		if tt, ok := t["type"].(string); ok {
			switch tt {
			case "map":
				return p.resolveType(t["values"], true)
			case "record":
				return p.resolveType(t["fields"], collection)
			case "array":
				return p.resolveType(t["items"], true)
			case "typeref":
				return p.resolveType(t["ref"], collection)
			}
		}
		return p.resolveType(t["type"], collection)
	case []interface{}:
		var typeRefs []TypeRef
		for _, item := range t {
			for _, typeRef := range p.resolveType(item, collection) {
				typeRefs = append(typeRefs, typeRef)
			}
		}
		return typeRefs
	}
	fatalf("unsupported type %#v\n", t)
	return nil
}

type TypeRef struct {
	Name       string
	Collection bool
}

func isPrimitive(name string) bool {
	switch name {
	case "int", "long", "float", "double", "bytes", "string", "null", "boolean", "fixed", "enum":
		return true
	}
	return false
}

func fullyQualifiedName(namespace, name string) string {
	if !strings.ContainsRune(name, '.') && len(namespace) > 0 {
		// Name is not fully qualified so prefix with namespace
		name = namespace + "." + name
	}
	return name
}
