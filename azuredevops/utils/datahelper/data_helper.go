package datahelper

func GetAttributeValues(items []interface{}, attributeName string) ([]string, error) {
	var result []string
	for _, element := range items {
		result = append(result, element.(map[string]interface{})[attributeName].(string))
	}
	return result, nil
}
