// Copyright 2020 tpkeeper
// SPDX-License-Identifier: LGPL-3.0-only

package utils

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	uuid "github.com/satori/go.uuid"
)

var location *time.Location
var dayLayout = "20060102"

func init() {
	var err error
	location, err = time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
}

func StrToFloat(str string) float64 {
	v, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0
	}
	return v
}

func StrToInt64(str string) (int64, error) {
	ret, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0, err
	}
	return ret, nil
}

func FloatToStr(f float64) string {
	v := strconv.FormatFloat(f, 'f', -1, 64)
	return v
}

func Uuid() string {
	return uuid.NewV4().String()
}
func IsImageExt(extName string) bool {
	var supportExtNames = map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".ico": true, ".svg": true, ".bmp": true, ".gif": true,
	}
	return supportExtNames[extName]
}

func GetSwapHash(swapType, sender string, created int64) string {
	return "0xswap" + hex.EncodeToString(
		crypto.Keccak256Hash([]byte(swapType+sender+strconv.FormatInt(created, 10))).Bytes())
}

func ToUpperList(list []string) []string {
	for i := range list {
		list[i] = strings.ToUpper(list[i])
	}
	return list
}

func GetNowUTC8Date() string {
	return time.Now().In(location).Format(dayLayout)
}

func GetNewDayUtc8Seconds() int64 {
	hour, min, sec := time.Now().In(location).Clock()
	return int64(hour*60*60 + min*60 + sec)
}

func GetYesterdayUTC8Date() string {
	return time.Now().In(location).AddDate(0, 0, -1).Format(dayLayout)
}

func AddOneDay(day string) (string, error) {
	timeParse, err := time.Parse(dayLayout, day)
	if err != nil {
		return "", err
	}
	return timeParse.AddDate(0, 0, 1).Format(dayLayout), nil
}

func SubOneDay(day string) (string, error) {
	timeParse, err := time.Parse(dayLayout, day)
	if err != nil {
		return "", err
	}
	return timeParse.AddDate(0, 0, -1).Format(dayLayout), nil
}

const DropRate10 = "10000000000000000000"
const DropRate7 = "7000000000000000000"
const DropRate4 = "4000000000000000000"

func GetDropRateFromTimestamp(startDay, stamp string) (string, error) {
	stampSec, err := strconv.Atoi(stamp)
	if err != nil {
		return "", err
	}
	stampDate := time.Unix(int64(stampSec), 0).In(location).Format(dayLayout)
	return GetDropRate(startDay, stampDate)
}

func GetDropRate(startDayStr, nowDayStr string) (string, error) {
	if startDayStr > nowDayStr {
		return "0", nil
	}
	startDay, err := time.Parse(dayLayout, startDayStr)
	if err != nil {
		return "", err
	}
	nowDay, err := time.Parse(dayLayout, nowDayStr)
	if err != nil {
		return "", err
	}
	interDays := nowDay.Sub(startDay).Milliseconds() / (24 * 60 * 60 * 1000)
	switchDay := interDays%30 + 1

	switch {
	case switchDay >= 1 && switchDay <= 5:
		return DropRate10, nil
	case switchDay >= 6 && switchDay <= 20:
		return DropRate7, nil
	case switchDay >= 21 && switchDay <= 30:
		return DropRate4, nil
	}
	return "", fmt.Errorf("switchDay err:%d", switchDay)
}
