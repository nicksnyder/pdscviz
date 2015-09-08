package main

type PDSC struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Fields    []PDSCField
}

type PDSCField struct {
	Type interface{} `json:"type"`
}

func (f PDSCField) typeRef() *TypeRef {
	switch t := f.Type.(type) {
	case string:
		return &TypeRef{Name: t, Collection: false}
	case map[string]interface{}:
		collection := false
		if tt, _ := t["type"].(string); tt == "array" {
			collection = true
		}
		if name, ok := t["items"].(string); ok {
			return &TypeRef{Name: name, Collection: collection}
		}
	}
	return nil
}

type TypeRef struct {
	Name       string
	Collection bool
}

func (tr *TypeRef) isPrimitive() bool {
	switch tr.Name {
	case "int", "long", "float", "double", "bytes", "string", "null", "boolean":
		return true
	}
	return false
}
