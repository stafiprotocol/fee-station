// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package utils_test

import (
	"fee-station/pkg/utils"
	"strconv"
	"testing"
	"time"
)

func TestGetSwapHash(t *testing.T) {
	timeNow := time.Now().UnixNano()
	t.Log(timeNow)
	t.Log(strconv.FormatInt(timeNow, 10))
	t.Log(utils.GetSwapHash("swap", "swap.Sender", time.Now().Unix()))
}

func TestGetNowUTC8Date(t *testing.T) {
	t.Log(utils.GetNowUTC8Date())
	t.Log(utils.GetYesterdayUTC8Date())
	timeParse,_:=time.Parse("20060102","0")
	t.Log(timeParse.String())
	timeParse2,_:=time.Parse("20060102","20200714")
	t.Log(timeParse2.Sub(timeParse).Hours()/24)


	t.Log("20210714" > "20200714")
	t.Log("20200814" > "20200714")
	t.Log("20200715" > "20200714")
	t.Log(utils.GetNewDayUtc8Seconds())
	t.Log(utils.GetDropRate("20200715","20200714"))
	t.Log(utils.GetDropRate("20200715","20200715"))
	t.Log(utils.GetDropRate("20200715","20200717"))
	t.Log(utils.GetDropRate("20200715","20200720"))
	t.Log(utils.GetDropRate("20200715","20200813"))
	t.Log(utils.GetDropRate("20200715","20200814"))
}
