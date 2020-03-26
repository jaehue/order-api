package models

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/hublabs/common/api"
	"github.com/hublabs/common/auth"

	"github.com/sirupsen/logrus"
)

func JsonToString(param interface{}) string {
	jsonStr, _ := json.Marshal(param)
	return string(jsonStr)
}
func LogParamErrorMsg(methodName string, param interface{}, err api.Error) {
	logrus.WithFields(logrus.Fields{
		"errorCode":    err.Code,
		"errorMessage": err.Message,
		"errorDetails": err.Details,
	}).Error(methodName+"error pramams==", JsonToString(param))
}

//DateParseToUtc parses a formatted local time string
//and returns the UTC time value it represents.
func DateParseToUtc(date string) (timeUtc time.Time, err error) {
	timeLayout := "2006-01-02"
	timeLoc, err := time.Parse(timeLayout, date)
	if err != nil {
		return
	}
	timeUtc = timeLoc.Add(time.Hour * -8)
	return
}
func StringInArr(str string, arr []string) bool {
	for _, item := range arr {
		if item == str {
			return true
		}
	}
	return false
}
func IsItemIdInOffer(id int64, offer OrderOffer) bool {
	result := false
	idStr := strconv.FormatInt(id, 10)
	if idStr != "" {
		arr := strings.Split(offer.ItemIds+","+offer.TargetItemIds, ",")
		for _, v := range arr {
			if idStr == v {
				result = true
			}
		}
	}
	return result
}

// TODO: github.com/hublabs/common/auth에 아래 로직 반영 & tenantCode 처리
type UserClaim auth.UserClaim

func (u UserClaim) isCustomer() bool {
	// return u.Iss == auth.IssMembership
	return u.Issuer == "membership"
}
func (u UserClaim) customerId() int64 {
	return 0
}
func (u UserClaim) tenantCode() string {
	return "hublabs"
}
func (u UserClaim) channelId() int64 {
	return 0
}
