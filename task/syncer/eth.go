package task

import (
	"encoding/json"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

var ethPath = "/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=%d&offset=%d&sort=asc&apikey=%s"

func SyncEthTx(db *db.WrapDb, ethEndpoint, apiKey string) error {
	poolAddressRes, err := dao_station.GetFeeStationPoolAddressBySymbol(db, utils.SymbolEth)
	if err != nil {
		return err
	}
	poolAddress := poolAddressRes.PoolAddress
	totalCount, err := dao_station.GetFeeStationNativeChainTxTotalCount(db, utils.SymbolEth)
	if err != nil {
		return err
	}

	usePage := totalCount/int64(pageLimit) + 1
	useUrl := fmt.Sprintf(ethEndpoint+ethPath, poolAddress, usePage, pageLimit, apiKey)
	txs, err := GetEthTxs(useUrl)
	if err != nil {
		return err
	}

	for _, tx := range txs.Result {
		useTxHash := strings.ToLower(tx.Hash)
		_, err := dao_station.GetFeeStationNativeChainTxBySymbolTxhash(db, utils.SymbolEth, useTxHash)
		//skip if exist
		if err == nil {
			continue
		}

		txStatus := 0
		if tx.IsError != "0" {
			txStatus = 1
		}
		if !strings.EqualFold(tx.To, poolAddress) {
			txStatus = 2
		}

		txTimestampDeci, err := decimal.NewFromString(tx.TimeStamp)
		if err != nil {
			return err
		}

		nativeTx := dao_station.FeeStationNativeChainTx{
			State:        0,
			TxStatus:     int64(txStatus),
			Symbol:       utils.SymbolEth,
			Blockhash:    strings.ToLower(tx.BlockHash),
			Txhash:       useTxHash,
			PoolAddress:  poolAddress,
			SenderPubkey: strings.ToLower(tx.From),
			InAmount:     tx.Value,
			TxTimestamp:  txTimestampDeci.BigInt().Int64(),
		}

		err = dao_station.UpOrInFeeStationNativeChainTx(db, &nativeTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetEthTxs(url string) (*ResEtherScan, error) {
	var res *http.Response
	var err error
	retry := 0
	for {
		if retry > BlockRetryLimit {
			return nil, fmt.Errorf("GetEthTxs reach retry limit: %s", err)
		}
		res, err = http.Get(url)
		if err != nil {
			logrus.Warnf("GetEthTxs err: %s", err)
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

	resEther := ResEtherScan{}
	err = json.Unmarshal(bodyBytes, &resEther)
	if err != nil {
		return nil, err
	}
	return &resEther, nil
}

type ResEtherScan struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Result  []struct {
		BlockNumber       string `json:"blockNumber"`
		TimeStamp         string `json:"timeStamp"`
		Hash              string `json:"hash"`
		Nonce             string `json:"nonce"`
		BlockHash         string `json:"blockHash"`
		TransactionIndex  string `json:"transactionIndex"`
		From              string `json:"from"`
		To                string `json:"to"`
		Value             string `json:"value"`
		Gas               string `json:"gas"`
		GasPrice          string `json:"gasPrice"`
		IsError           string `json:"isError"`
		TxreceiptStatus   string `json:"txreceipt_status"`
		Input             string `json:"input"`
		ContractAddress   string `json:"contractAddress"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"`
		GasUsed           string `json:"gasUsed"`
		Confirmations     string `json:"confirmations"`
	} `json:"result"`
}
