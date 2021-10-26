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
)

var path = "/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&page=%d&offset=%d&sort=asc&apikey=%s"

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
	useUrl := fmt.Sprintf(ethEndpoint+path, poolAddress, usePage, pageLimit, apiKey)
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
			txStatus = 1
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
		}

		err = dao_station.UpOrInFeeStationNativeChainTx(db, &nativeTx)
		if err != nil {
			return err
		}
	}

	return nil
}

func GetEthTxs(url string) (*ResEtherScan, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
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
