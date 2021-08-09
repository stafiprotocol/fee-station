package task

import (
	"fee-station/dao/station"
	"fee-station/pkg/db"
	"fee-station/pkg/utils"
	"fee-station/shared/substrate"
	"fmt"
	"time"

	"github.com/JFJun/go-substrate-crypto/ss58"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"
	"github.com/stafiprotocol/go-substrate-rpc-client/signature"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

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
	logrus.Infof("will pay recievers: %v \n", strFi(receives))

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

	for _, swapInfo := range swapInfoList {
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
