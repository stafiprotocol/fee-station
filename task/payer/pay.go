package task

import (
	"fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fee-station/shared/substrate"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/stafiprotocol/go-substrate-rpc-client/signature"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

func CheckPayInfo(db *db.WrapDb, fisTypesPath, fisEndpoint, swapLimit string, key *signature.KeyringPair) error {
	swapInfoList, err := dao_station.GetSwapInfoListByState(db, utils.SwapStateVerifyTxOk)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}

	gc, err := substrate.NewGsrpcClient(fisEndpoint, substrate.AddressTypeAccountId, key)
	if err != nil {
		return err
	}

	swapLimitDeci, err := decimal.NewFromString(swapLimit)
	if err != nil {
		return err
	}
	receives := make([]*substrate.Receive, 0)
	for _, swapInfo := range swapInfoList {
		stafiAddressBytes, err := hexutil.Decode(swapInfo.StafiAddress)
		if err != nil {
			return err
		}
		outAmountDeci, err := decimal.NewFromString(swapInfo.OutAmount)
		if err != nil {
			return err
		}
		if outAmountDeci.Cmp(swapLimitDeci) > 0 {
			return fmt.Errorf("outAmount > swapLimit,out: %s", outAmountDeci.StringFixed(0))
		}

		receive := substrate.Receive{
			Recipient: stafiAddressBytes,
			Value:     types.NewUCompact(outAmountDeci.BigInt()),
		}
		receives = append(receives, &receive)
	}
	err = gc.BatchTransfer(receives)
	if err != nil {
		return err
	}
	tx := db.NewTransaction()
	for _, swapInfo := range swapInfoList {
		swapInfo.State = utils.SwapStatePayOk
		err := dao_station.UpOrInSwapInfo(tx, swapInfo)
		if err != nil {
			tx.RollbackTransaction()
			return err
		}
	}
	err = tx.CommitTransaction()
	if err != nil {
		return fmt.Errorf("tx.CommitTransaction err: %s", err)
	}

	return nil
}
