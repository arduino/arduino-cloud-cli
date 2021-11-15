package tag

import "fmt"

// TagsInfo transforms tags into user-readable strings
// An error is returned if a tag value is not a string
func TagsInfo(tags map[string]interface{}) ([]string, error) {
	var str []string
	for key, value := range tags {
		valStr, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("value of tag `%s` should be of type `string` but is of type `%T`", key, value)
		}
		str = append(str, key+": "+valStr)
	}
	return str, nil
}
