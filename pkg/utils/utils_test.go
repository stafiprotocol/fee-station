// Copyright 2021 stafiprotocol
// SPDX-License-Identifier: LGPL-3.0-only

package utils_test

import (
	"fee-station/pkg/utils"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	ethCrypto "github.com/ethereum/go-ethereum/crypto"
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
	timeParse, _ := time.Parse("20060102", "0")
	t.Log(timeParse.String())
	timeParse2, _ := time.Parse("20060102", "20200714")
	t.Log(timeParse2.Sub(timeParse).Hours() / 24)

	t.Log("20210714" > "20200714")
	t.Log("20200814" > "20200714")
	t.Log("20200715" > "20200714")
	t.Log(utils.GetNewDayUtc8Seconds())
	t.Log(utils.GetDropRate("20200715", "20200714"))
	t.Log(utils.GetDropRate("20200715", "20200715"))
	t.Log(utils.GetDropRate("20200715", "20200717"))
	t.Log(utils.GetDropRate("20200715", "20200720"))
	t.Log(utils.GetDropRate("20200715", "20200813"))
	t.Log(utils.GetDropRate("20200715", "20200814"))
}

func TestVerifySigsEth(t *testing.T) {
	sigs, err := hexutil.Decode("0xe95bf9f5600771161308183a43e7b5a3a5ef410912cde5fbd1382293deec88146815f155df18c33a16f86f0d48b9ca170c3ac65e9919c5816b012a9c40edfafc1b")
	if err!=nil{
		t.Fatal(err)
	}
	msg, err := hexutil.Decode("0x66d410cde3a337cf45b171dbb9b90762cc0a6c60cff3b8229befdd7678afa669")
	if err!=nil{
		t.Fatal(err)
	}
	ok := utils.VerifySigsEth(sigs, msg, common.HexToAddress("0x3aab5AE578FA45744aFe8224DdA506cFE67c508b"))
	msgHash := ethCrypto.Keccak256(msg)
	t.Log(hexutil.Encode(msgHash))
	t.Log(ok)
}
