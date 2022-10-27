package util

//slice、map深度拷贝函数
func DeepCopy(value interface{}) interface{} {
	if valueMap, ok := value.(map[string]interface{}); ok {
		newMap := make(map[string]interface{})
		for k, v := range valueMap {
			newMap[k] = DeepCopy(v)
		}

		return newMap
	} else if valueSlice, ok := value.([]interface{}); ok {
		newSlice := make([]interface{}, len(valueSlice))
		for k, v := range valueSlice {
			newSlice[k] = DeepCopy(v)
		}

		return newSlice
	}

	return value
}

//map拷贝
func MapCopy(valueMap map[string]interface{}) map[string]interface{} {
	newMap := make(map[string]interface{})
	for k, v := range valueMap {
		newMap[k] = v
	}
	return newMap
}
