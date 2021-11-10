package task

import (
	"bytes"
	"encoding/json"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var substrateTxPath = "/api/scan/transfers"
var substrateBlockPath = "/api/scan/block"
var substratePageLimit = 100

func SyncDotTx(db *db.WrapDb, dotEndpoint, apiKey string) error {
	poolAddressRes, err := dao_station.GetFeeStationPoolAddressBySymbol(db, utils.SymbolDot)
	if err != nil {
		return err
	}
	poolAddress := poolAddressRes.PoolAddress

	usePage := 0
	useUrl := dotEndpoint + substrateTxPath
	txs, err := GetSubstrateTxs(useUrl, poolAddress, apiKey, int(usePage), substratePageLimit)
	if err != nil {
		return err
	}

	if txs.Code != 0 {
		logrus.Errorf("getSubstrateTxs res code %d,url: %s", txs.Code, useUrl)
		return nil
	}

	for _, tx := range txs.Data.Transfers {
		useTxHash := strings.ToLower(tx.Hash)
		_, err := dao_station.GetFeeStationNativeChainTxBySymbolTxhash(db, utils.SymbolDot, useTxHash)
		//skip if exist
		if err == nil {
			continue
		}

		txStatus := 0
		if !tx.Success {
			txStatus = 1
		}
		if !strings.EqualFold(tx.To, poolAddress) {
			txStatus = 2
		}

		amountDeci, err := decimal.NewFromString(tx.Amount)
		if err != nil {
			return err
		}
		time.Sleep(6 * time.Second)
		resBlock, err := GetSubstrateBlock(dotEndpoint+substrateBlockPath, apiKey, tx.BlockNum)
		if err != nil {
			return fmt.Errorf("GetSubstrateBlock failed: %s", err)
		}
		if resBlock.Code != 0 {
			logrus.Errorf("getSubstrateBlock res code %d,url: %s", txs.Code, useUrl)
			return nil
		}

		pubkeyBytes, err := ss58.DecodeToPub(tx.From)
		if err != nil {
			return err
		}

		nativeTx := dao_station.FeeStationNativeChainTx{
			State:        0,
			TxStatus:     int64(txStatus),
			Symbol:       utils.SymbolDot,
			Blockhash:    strings.ToLower(resBlock.Data.Hash),
			Txhash:       useTxHash,
			PoolAddress:  poolAddress,
			SenderPubkey: strings.ToLower(hexutil.Encode(pubkeyBytes)),
			InAmount:     amountDeci.Mul(decimal.New(1, 10)).StringFixed(0),
			TxTimestamp:  int64(tx.BlockTimestamp),
		}
		err = dao_station.UpOrInFeeStationNativeChainTx(db, &nativeTx)
		if err != nil {
			return err
		}

	}

	return nil
}

func GetSubstrateTxs(url, address, apiKey string, page, pageLimit int) (*ResSubstrateTxs, error) {
	reqSubTxs := ReqSubstrateTxs{
		Row:       pageLimit,
		Page:      page,
		Address:   address,
		FromBlock: 0,
		ToBlock:   999999999,
	}

	reqBts, err := json.Marshal(reqSubTxs)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBts))
	if err != nil {
		return nil, err
	}
	if len(apiKey) != 0 {
		req.Header.Add("X-API-Key", apiKey)
	}
	req.Header.Add("Content-Type", "application/json")

	var res *http.Response
	retry := 0
	for {
		if retry > BlockRetryLimit {
			return nil, fmt.Errorf("GetSubstrateTxs reach retry limit: %s", err)
		}
		res, err = client.Do(req)
		if err != nil {
			logrus.Warnf("GetSubstrateTxs err: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status err: %d", res.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) == 0 {
		return nil, fmt.Errorf("body bytes empty")
	}

	resSubTxs := ResSubstrateTxs{}

	err = json.Unmarshal(bodyBytes, &resSubTxs)
	if err != nil {
		return nil, err
	}
	return &resSubTxs, nil
}

func GetSubstrateBlock(url, apiKey string, number int) (*ResSubstrateBlock, error) {
	reqSubTxs := ReqSubstrateBlock{
		BlockNumber: number,
	}

	reqBts, err := json.Marshal(reqSubTxs)
	if err != nil {
		return nil, err
	}
	client := http.DefaultClient
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqBts))
	if err != nil {
		return nil, err
	}
	if len(apiKey) != 0 {
		req.Header.Add("X-API-Key", apiKey)
	}
	req.Header.Add("Content-Type", "application/json")

	var res *http.Response
	retry := 0
	for {
		if retry > BlockRetryLimit {
			return nil, fmt.Errorf("GetSubstrateBlock reach retry limit: %s", err)
		}
		res, err = client.Do(req)
		if err != nil {
			logrus.Warnf("GetSubstrateBlock err: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status err: %d", res.StatusCode)
	}
	bodyBytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if len(bodyBytes) == 0 {
		return nil, fmt.Errorf("body bytes empty")
	}

	resSubBlock := ResSubstrateBlock{}

	err = json.Unmarshal(bodyBytes, &resSubBlock)
	if err != nil {
		return nil, err
	}
	return &resSubBlock, nil
}

type ReqSubstrateTxs struct {
	Row       int    `json:"row"`
	Page      int    `json:"page"`
	Address   string `json:"address"`
	FromBlock int    `json:"from_block"`
	ToBlock   int    `json:"to_block"`
}

type ResSubstrateTxs struct {
	Code int `json:"code"`
	Data struct {
		Count     int `json:"count"`
		Transfers []struct {
			Amount             string `json:"amount"`
			BlockNum           int    `json:"block_num"`
			BlockTimestamp     int    `json:"block_timestamp"`
			ExtrinsicIndex     string `json:"extrinsic_index"`
			Fee                string `json:"fee"`
			From               string `json:"from"`
			FromAccountDisplay struct {
				AccountIndex  string      `json:"account_index"`
				Address       string      `json:"address"`
				Display       string      `json:"display"`
				Identity      bool        `json:"identity"`
				Judgements    interface{} `json:"judgements"`
				Parent        string      `json:"parent"`
				ParentDisplay string      `json:"parent_display"`
			} `json:"from_account_display"`
			Hash             string `json:"hash"`
			Module           string `json:"module"`
			Nonce            int    `json:"nonce"`
			Success          bool   `json:"success"`
			To               string `json:"to"`
			ToAccountDisplay struct {
				AccountIndex  string      `json:"account_index"`
				Address       string      `json:"address"`
				Display       string      `json:"display"`
				Identity      bool        `json:"identity"`
				Judgements    interface{} `json:"judgements"`
				Parent        string      `json:"parent"`
				ParentDisplay string      `json:"parent_display"`
			} `json:"to_account_display"`
		} `json:"transfers"`
	} `json:"data"`
	Message     string `json:"message"`
	GeneratedAt int    `json:"generated_at"`
}

type ReqSubstrateBlock struct {
	BlockNumber int `json:"block_num"`
}

type ResSubstrateBlock struct {
	Code        int    `json:"code"`
	Message     string `json:"message"`
	GeneratedAt int    `json:"generated_at"`
	Data        struct {
		BlockNum       int    `json:"block_num"`
		BlockTimestamp int    `json:"block_timestamp"`
		Hash           string `json:"hash"`
		ParentHash     string `json:"parent_hash"`
		StateRoot      string `json:"state_root"`
		ExtrinsicsRoot string `json:"extrinsics_root"`
	} `json:"data"`
}
