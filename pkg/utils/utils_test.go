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
	if err != nil {
		t.Fatal(err)
	}
	msg, err := hexutil.Decode("0x66d410cde3a337cf45b171dbb9b90762cc0a6c60cff3b8229befdd7678afa669")
	if err != nil {
		t.Fatal(err)
	}
	ok := utils.VerifySigsEth(sigs, msg, common.HexToAddress("0x3aab5AE578FA45744aFe8224DdA506cFE67c508b"))
	msgHash := ethCrypto.Keccak256(msg)
	t.Log(hexutil.Encode(msgHash))
	t.Log(ok)
}

func TestVerifySigs25519(t *testing.T) {
	sigs, err := hexutil.Decode("0xf47b28366ad87a65a802d596533d33f869fff163bb7582fe5c86ad3e0151950121700ba862869ffa482b9d74955785fbe2c5075619926ec24f7eda9d53d40681")
	if err != nil {
		t.Fatal(err)
	}
	msg, err := hexutil.Decode("0xcca4cef73a88bcc6b1ef7087393f5ca226ce795bf20b5868d140690c7d419156")
	if err != nil {
		t.Fatal(err)
	}
	pubkey, err := hexutil.Decode("0xcca4cef73a88bcc6b1ef7087393f5ca226ce795bf20b5868d140690c7d419156")
	if err != nil {
		t.Fatal(err)
	}
	ok := utils.VerifiySigsSr25519(sigs, pubkey, msg)
	msgHash := ethCrypto.Keccak256(msg)
	t.Log(hexutil.Encode(msgHash))
	t.Log(ok)
}

func TestGetPrice(t *testing.T){
	url:="https://api.coingecko.com/api/v3/simple/price?ids=ethereum,polkadot,cosmos,stafi,kusama&vs_currencies=usd"
	prices,err:=utils.GetPriceFromCoinGecko(url)
	if err!=nil{
		t.Fatal(err)
	}
	t.Log(prices)
}