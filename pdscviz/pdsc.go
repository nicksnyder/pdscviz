package main

// TODO: handle includes
type PDSC struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Fields    []PDSCField
}

type PDSCField struct {
	Type interface{} `json:"type"`
}

func (f PDSCField) typeRefs() []TypeRef {
	return resolveType(f.Type, false)
}

func resolveType(t interface{}, collection bool) []TypeRef {
	switch t := t.(type) {
	case string:
		if isPrimitive(t) {
			return nil
		}
		return []TypeRef{{Name: t, Collection: collection}}
	case map[string]interface{}:
		if tt, ok := t["type"].(string); ok {
			switch tt {
			case "map":
				return resolveType(t["values"], true)
			case "record":
				return resolveType(t["fields"], false)
			case "array":
				return resolveType(t["items"], true)
			}
		}
		return resolveType(t["type"], collection)
	case []interface{}:
		var typeRefs []TypeRef
		for _, item := range t {
			for _, typeRef := range resolveType(item, collection) {
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
