package task

import (
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func SyncKsmTx(db *db.WrapDb, ksmEndpoint, apiKey string) error {
	poolAddressRes, err := dao_station.GetFeeStationPoolAddressBySymbol(db, utils.SymbolKsm)
	if err != nil {
		return err
	}
	poolAddress := poolAddressRes.PoolAddress

	usePage := 1
	useUrl := ksmEndpoint + substrateTxPath
	txs, err := GetSubstrateTxs(useUrl, poolAddress, apiKey, int(usePage), pageLimit)
	if err != nil {
		return err
	}

	if txs.Code != 0 {
		logrus.Errorf("getSubstrateTxs res code %d,url: %s", txs.Code, useUrl)
		return nil
	}

	pageMax := txs.Data.Count/pageLimit + 1

	for i := 1; i <= pageMax; i++ {
		time.Sleep(6 * time.Second)
		txs, err := GetSubstrateTxs(useUrl, poolAddress, apiKey, i, pageLimit)
		if err != nil {
			return err
		}

		if txs.Code != 0 {
			logrus.Errorf("getSubstrateTxs res code %d,url: %s", txs.Code, useUrl)
			return nil
		}

		for _, tx := range txs.Data.Transfers {
			useTxHash := strings.ToLower(tx.Hash)
			_, err := dao_station.GetFeeStationNativeChainTxBySymbolTxhash(db, utils.SymbolKsm, useTxHash)
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
			resBlock, err := GetSubstrateBlock(ksmEndpoint+substrateBlockPath, apiKey, tx.BlockNum)
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
				Symbol:       utils.SymbolKsm,
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
	}

	return nil
}
