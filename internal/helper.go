package internal

func contains(arr []string, elem string) bool {
	for _, itm := range arr {
		if itm == elem {
			return true
		}
	}
	return false
}

func uniqueAppend(arr []string, elem string) []string {
	if !contains(arr, elem) {
		return append(arr, elem)
	}
	return arr
}
