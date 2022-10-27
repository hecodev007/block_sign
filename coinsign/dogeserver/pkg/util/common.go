package util

// 通过两重循环过滤重复元素
func StingArrayToRemoveRepeat(slc []string) []string {
	result := []string{} // 存放结果

	for i, v := range slc {
		if v == "" {
			continue
		}
		flag := true
		for j, _ := range result {
			if slc[i] == result[j] {
				flag = false // 存在重复元素，标识为false
				break
			}
		}
		if flag { // 标识为false，不添加进结果
			result = append(result, slc[i])
		}
	}
	return result
}

//map去重
func StringArrayRemoveRepeatByMap(slc []string) []string {
	result := make([]string, 0)
	mapdata := make(map[string]string, 0)
	for _, v := range slc {
		mapdata[v] = v
	}
	for k, _ := range mapdata {
		result = append(result, k)
	}
	return result
}
