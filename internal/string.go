package internal

import (
	"encoding/json"
	"regexp"

	"github.com/steffakasid/eslog"
)

func Contains(arr []string, elem string) bool {
	for _, itm := range arr {
		if itm == elem {
			return true
		}
	}
	return false
}

func UniqueAppend(arr []string, elem string) []string {
	if !Contains(arr, elem) {
		return append(arr, elem)
	}
	return arr
}

func MatchAny(str string, regExps []string) (bool, error) {
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

func ToJSONString(obj any) string {
	bt, err := json.Marshal(obj)
	eslog.LogIfError(err, eslog.Error, err)

	return string(bt)
}
