package policy

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Document ... is a generic structure that is used by UnmarshalPolicy
type Document struct {
	Version   string
	Statement []*Statement
}

// Statement ... is a generic structure to hold an AWS policy statement
type Statement struct {
	Sid       string
	Effect    string
	Action    []string
	Resource  []string
	Principal *Principal
	Condition []*Condition
}

// Principal ... holds an AWS policy principal
type Principal struct {
	Type   string
	Values []string
}

//Condition ... holds an AWS policy condition
type Condition struct {
	Operator string
	Property string
	Value    []string
}

// Unmarshal ... unmarshals a raw policy document
func Unmarshal(raw string) (*Document, error) {
	data, err := url.QueryUnescape(raw)
	if err != nil {
		return nil, fmt.Errorf("failed to unescape Policy Document: %v", err)
	}
	pdoc, err := parsePolicy([]byte(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse policy document: %v", err)
	}
	return pdoc, nil
}

// setterFunc ... used by parseStatement to set each property
type setterFunc func(*Statement, interface{}) error

// parsePolicy ... takes a policy document in json format and returns a *types.PolicyDocument
func parsePolicy(data []byte) (*Document, error) {
	var m map[string]interface{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	var pdoc Document
	pdoc.Version = m["Version"].(string)
	items := m["Statement"].([]interface{})
	for i := 0; i < len(items); i++ {
		item := items[i].(map[string]interface{})
		err = parseStatement(&pdoc, item)
		if err != nil {
			return nil, err
		}
	}
	return &pdoc, nil
}

// parseStatement ... takes a *Document and the Statement value as map[string]interface{}
// then populates the *Document with the parsed policy statements
func parseStatement(doc *Document, m map[string]interface{}) error {
	statement := &Statement{}
	setters := map[string]setterFunc{
		"Sid":       setSid,
		"Effect":    setEffect,
		"Principal": setPrincipal,
		"Action":    setAction,
		"Resource":  setResource,
		"Condition": setCondition,
	}
	for field, fn := range setters {
		if val, ok := m[field]; ok {
			err := fn(statement, val)
			if err != nil {
				return fmt.Errorf("failed to set field %s: %v", field, err)
			}
		}
	}
	doc.Statement = append(doc.Statement, statement)
	return nil
}

// setSid ... converts and sets Sid in statement
func setSid(statement *Statement, m interface{}) error {
	statement.Sid = m.(string)
	return nil
}

// setEffect ... converts and sets Effect in statement
func setEffect(statement *Statement, m interface{}) error {
	statement.Effect = m.(string)
	return nil
}

// setPrincipal ... converts and sets Principal in statement
func setPrincipal(statement *Statement, m interface{}) error {
	statement.Principal = &Principal{}
	err := setPrincipalProperty(statement.Principal, m.(map[string]interface{}))
	return err
}

// setPrincipalProperty ... converts and sets Principal Type and Values
func setPrincipalProperty(principal *Principal, m map[string]interface{}) error {
	for k, v := range m {
		principal.Type = k
		switch val := v.(type) {
		case string:
			principal.Values = []string{val}
		case []string:
			principal.Values = val
		case interface{}:
			for _, item := range val.([]interface{}) {
				principal.Values = append(principal.Values, item.(string))
			}
		default:
			return fmt.Errorf("type not supported: %T", val)
		}
		break
	}
	return nil
}

// setAction ... converts and sets Action in statement
func setAction(statement *Statement, m interface{}) (err error) {
	statement.Action, err = interfaceToStringSlice(m)
	return err
}

// setResource ... converts and sets Resource in statement
func setResource(statement *Statement, m interface{}) (err error) {
	statement.Resource, err = interfaceToStringSlice(m)
	return err
}

// setCondition ... converts and sets Condition in statement
func setCondition(statement *Statement, m interface{}) (err error) {
	statement.Condition = []*Condition{}
	mm := m.(map[string]interface{})
	for k, v := range mm {
		conditions := []*Condition{}
		operator := k
		switch val := v.(type) {
		case map[string]interface{}:
			for k, v := range val {
				c := &Condition{Operator: operator, Property: k}
				c.Value, err = interfaceToStringSlice(v)
				if err != nil {
					return
				}
				conditions = append(conditions, c)
			}
		case map[string][]string:
			for k, v := range val {
				c := &Condition{Operator: operator, Property: k, Value: v}
				conditions = append(conditions, c)
			}
		case map[string]string:
			for k, v := range val {
				c := &Condition{Operator: operator, Property: k, Value: []string{v}}
				conditions = append(conditions, c)
			}
		default:
			return fmt.Errorf("type not supported: %T", val)
		}
		statement.Condition = append(statement.Condition, conditions...)
	}
	return nil
}

// interfaceToStringSlice ... converts []interface{}, []string, string to []string
func interfaceToStringSlice(m interface{}) ([]string, error) {
	switch val := m.(type) {
	case []interface{}:
		var value []string
		for _, v := range val {
			value = append(value, v.(string))
		}
		return value, nil
	case []string:
		return val, nil
	case string:
		return []string{val}, nil
	default:
		return nil, fmt.Errorf("type not supported: %T", val)
	}
}
