package tag

import "fmt"

type Tags map[string]interface{}

// Info transforms tags into user-readable strings
// An error is returned if a tag value is not a string
func (t Tags) Info() ([]string, error) {
	var str []string
	for key, value := range t {
		if valStr, ok := value.(string); ok {
			str = append(str, key+": "+valStr)
		} else {
			return nil, fmt.Errorf("value of tag `%s` should be of type `string` but is of type `%T`", key, value)
		}
	}
	return str, nil
}
