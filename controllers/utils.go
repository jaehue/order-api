package controllers

import (
	"strconv"
	"strings"
	"time"
)

const defaultMaxResultCount = 30

//DateTermMaxValidate validate data maxterm
func DateTermMaxValidate(startAt, endAt string, term int) (result bool, err error) {
	if startAt == "" && endAt == "" {
		return true, nil
	} else if startAt != "" && endAt != "" {
		timeLayout := "2006-01-02"
		var startTime, endTime time.Time
		startTime, err = time.Parse(timeLayout, startAt)
		if err != nil {
			return false, err
		}
		endTime, err = time.Parse(timeLayout, endAt)
		if err != nil {
			return false, err
		}
		if startTime.After(endTime) {
			return false, nil
		}
		if startTime.AddDate(0, 0, term-1).Before(endTime) {
			return false, nil
		}
		return true, nil
	}
	return false, nil
}

func StringToInt64Arr(str string, sep string, removeZero bool) (result []int64) {
	strArr := strings.Split(str, ",")
	for _, item := range strArr {
		itemInt, _ := strconv.ParseInt(item, 10, 64)
		if removeZero && itemInt == 0 {
			continue
		}
		if !IsInArr(result, itemInt) {
			result = append(result, itemInt)
		}
	}
	return
}
func Int64ArrToString(intArr []int64) (result string) {
	for i, item := range intArr {
		result += strconv.FormatInt(item, 10)
		if i != len(intArr)-1 {
			result += ","
		}
	}
	return
}
func IsInArr(arr []int64, item int64) bool {
	for i, _ := range arr {
		if arr[i] == item {
			return true
		}
	}
	return false
}
func StringInArr(str string, arr []string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}
func makeTimestamp() int64 {
	return time.Now().UnixNano() / 1000000
}
