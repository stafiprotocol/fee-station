package task

import (
	"fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fee-station/shared/substrate"
	"fmt"
	"math/big"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stafiprotocol/go-substrate-rpc-client/signature"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

var batchNumberLimit = 200
var minReserveValue = big.NewInt(10e12)

func CheckPayInfo(db *db.WrapDb, fisEndpoint, swapLimit string, key *signature.KeyringPair) error {
	swapInfoList, err := dao_station.GetSwapInfoListByState(db, utils.SwapStateVerifyTxOk)
	if err != nil {
		return err
	}
	if len(swapInfoList) == 0 {
		return nil
	}
	retry := 0
	var gc *substrate.GsrpcClient
	for {
		if retry > BlockRetryLimit {
			return fmt.Errorf("substrate.NewGsrpcClient reach retry limit")
		}
		gc, err = substrate.NewGsrpcClient(fisEndpoint, substrate.AddressTypeAccountId, key)
		if err != nil {
			time.Sleep(BlockRetryInterval)
			retry++
			continue
		}
		break
	}

	swapLimitDeci, err := decimal.NewFromString(swapLimit)
	if err != nil {
		return err
	}
	accountInfo, err := gc.GetAccountInfo()
	if err != nil {
		return err
	}
	if accountInfo.Data.Free.Cmp(minReserveValue) < 0 {
		return fmt.Errorf("insufficient balance")
	}
	maxTransferAmount := new(big.Int).Sub(accountInfo.Data.Free.Int, minReserveValue)

	willTransferAmount := big.NewInt(0)
	receives := make([]*substrate.Receive, 0)
	transferMaxIndex := -1
	for i, swapInfo := range swapInfoList {
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

		tempAmount := new(big.Int).Add(willTransferAmount, outAmountDeci.BigInt())
		if tempAmount.Cmp(maxTransferAmount) > 0 {
			break
		}

		willTransferAmount = tempAmount
		transferMaxIndex = i
		receive := substrate.Receive{
			Recipient: stafiAddressBytes,
			Value:     types.NewUCompact(outAmountDeci.BigInt()),
		}
		receives = append(receives, &receive)
	}

	if len(receives) == 0 {
		return fmt.Errorf("insufficient balance")
	}
	logrus.Infof("will pay recievers: %v \n", strFi(receives))

	err = gc.BatchTransfer(receives)
	if err != nil {
		return err
	}
	tx := db.NewTransaction()
	for i, swapInfo := range swapInfoList {
		if i > transferMaxIndex {
			break
		}
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

	for i, swapInfo := range swapInfoList {
		if i > transferMaxIndex {
			break
		}
		new, err := dao_station.GetSwapInfoBySymbolBlkTx(db, swapInfo.Symbol, swapInfo.Blockhash, swapInfo.Txhash)
		if err != nil {
			return err
		}
		if new.State != utils.SwapStatePayOk {
			return fmt.Errorf("pay state in db not update")
		}
	}
	return nil
}

func strFi(recievers []*substrate.Receive) string {
	ret := ""
	for _, re := range recievers {
		bonderAddr, _ := ss58.Encode(re.Recipient, ss58.StafiPrefix)
		ret += "\n"
		ret += bonderAddr
		ret += " "
		ret += fmt.Sprintf("%v", re.Value)
	}
	return ret
}
