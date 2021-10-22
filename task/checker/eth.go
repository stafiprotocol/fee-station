package task

import (
	"context"
	dao_station "fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
)

func CheckEthTx(db *db.WrapDb, ethEndpoint string) error {
	swapInfoList, err := dao_station.GetSwapInfoListBySymbolState(db, utils.SymbolEth, utils.SwapStateVerifySigs)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}

	retry := 0
	var client *ethclient.Client
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("ethclient.Dial reach retry limit")
		}
		client, err = ethclient.Dial(ethEndpoint)
		if err != nil {
			logrus.Warnf("ethclient dial: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		_, err = client.BlockNumber(context.Background())
		if err != nil {
			logrus.Warnf("ethclient dial: %s", err)
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	for _, swapInfo := range swapInfoList {
		var status uint8
		retry := 0
		var err error
		for {
			if retry > 10 {
				return err
			}
			status, err = TransferVerifyEth(client, swapInfo)
			if err != nil {
				logrus.Errorf("eth TransferVerify failed: %s", err)
				retry++
				time.Sleep(BlockRetryInterval)
				continue
			}
			break
		}
		swapInfo.State = status
		err = dao_station.UpOrInSwapInfo(db, swapInfo)
		if err != nil {
			logrus.Errorf("dao_station.UpOrInSwapInfo err: %s", err)
			return err
		}
	}

	return nil
}

func TransferVerifyEth(client *ethclient.Client, swapInfo *dao_station.SwapInfo) (uint8, error) {
	block, err := client.BlockByHash(context.Background(), common.HexToHash(swapInfo.Blockhash))
	if err != nil && err != ethereum.NotFound {
		return 0, err
	}
	if err != nil && err == ethereum.NotFound {
		return utils.SwapStateBlockHashFailed, nil
	}

	swapNumber := block.NumberU64()
	// wait 1 block
	retry := 0
	for {
		if retry > BlockRetryLimit {
			return 0, fmt.Errorf("wait 3 block,reach retry limit")
		}
		latestNumber, err := client.BlockNumber(context.Background())
		if err == nil && latestNumber > swapNumber+1 {
			break
		} else {
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
	}

	inAmountDeci, err := decimal.NewFromString(swapInfo.InAmount)
	if err != nil {
		return 0, err
	}
	tx, _, err := client.TransactionByHash(context.Background(), common.HexToHash(swapInfo.Txhash))
	if err != nil && err != ethereum.NotFound {
		return 0, err
	}
	if err != nil && err == ethereum.NotFound {
		return utils.SwapStateTxHashFailed, nil
	}
	//check pool address
	if !strings.EqualFold(tx.To().String(), swapInfo.PoolAddress) {
		return utils.SwapStatePoolAddressFailed, nil
	}
	//check amount
	if tx.Value().Cmp(inAmountDeci.BigInt()) != 0 {
		return utils.SwapStateAmountFailed, nil
	}
	//check blockhash
	txReceipt, err := client.TransactionReceipt(context.Background(), common.HexToHash(swapInfo.Txhash))
	if err != nil && err != ethereum.NotFound {
		return 0, err
	}
	if err != nil && err == ethereum.NotFound {
		return utils.SwapStateTxHashFailed, nil
	}
	if !strings.EqualFold(txReceipt.BlockHash.String(), swapInfo.Blockhash) {
		return utils.SwapStateBlockHashFailed, nil
	}

	//check user address
	sender, err := client.TransactionSender(context.Background(), tx, txReceipt.BlockHash, txReceipt.TransactionIndex)
	if err != nil {
		return 0, err
	}
	if !strings.EqualFold(sender.String(), swapInfo.Pubkey) {
		return utils.SwapStatePubkeyFailed, nil
	}

	return utils.SwapStateVerifyTxOk, nil
}
