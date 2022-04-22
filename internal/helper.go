package internal

import "regexp"

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

func matchAny(str string, regExps []string) (bool, error) {
	for _, regExpStr := range regExps {
		regExp, err := regexp.Compile(regExpStr)
		if err != nil {
			return false, err
		}
		if regExp.MatchString(str) {
			return true, nil
		}
	}
	return false, nil
}
